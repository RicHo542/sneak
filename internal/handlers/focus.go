package handlers

import (
	"errors"
	"fmt"

	"github.com/richo542/sneak/internal/app"
	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/todos"
	"github.com/richo542/sneak/internal/ui"
)

// TaskFocus selects the lifecycle target pool. The todo leg is a status set
// (a todo has exactly one status, so pools are disjoint); the provider leg
// selects between the active and non-active cache pools.
type TaskFocus uint8

const (
	// FocusClose targets items that can be closed: open or active todos and
	// active provider items.
	FocusClose TaskFocus = iota
	// FocusStart targets items that can be started: open todos and non-active
	// provider items.
	FocusStart
)

func (f TaskFocus) todoStatuses() []string {
	switch f {
	case FocusStart:
		return []string{todos.TodoStatusOpen}
	default:
		return []string{todos.TodoStatusOpen, todos.TodoStatusActive}
	}
}

func (f TaskFocus) providerAll(state *config.State) ([]*config.CacheItem, error) {
	switch f {
	case FocusStart:
		return state.GetNonActiveCacheItems()
	default:
		return state.GetActiveCacheItems()
	}
}

func HasInvalidTasks(app *app.App, tasks []string) ([]string, bool) {
	// validate that task names are actually referring to
	// existing tasks
	var invalidTasks []string
	for _, task := range tasks {
		valid := false
		for _, item := range app.State.Cache.Items {
			if task == item.Key {
				valid = true
			}
		}

		if !valid {
			invalidTasks = append(invalidTasks, task)
		}
	}

	return invalidTasks, len(invalidTasks) > 0
}

// ContainsProviderKeys checks if a list of keys contain keys belonging to
// provider items. It infers this by checking if all keys are prefixed as local
// todo items "sn:XY"
func ContainsProviderKeys(keys []string) bool {
	for _, key := range keys {
		if !todos.HasTodoRef(key) {
			return true
		}
	}
	return false
}

func separateKeys(keys []string) ([]string, []string) {
	var todoKeys []string
	var providerKeys []string

	for _, key := range keys {
		if todos.HasTodoRef(key) {
			todoKeys = append(todoKeys, key)
		} else {
			providerKeys = append(providerKeys, key)
		}
	}
	return todoKeys, providerKeys
}

func filterSelected(cacheItems []*config.CacheItem, todoItems []*todos.Todo, keys map[string]struct{}) ([]*config.CacheItem, []*todos.Todo, error) {
	selCache := []*config.CacheItem{}
	for _, item := range cacheItems {
		if _, ok := keys[item.Key]; ok {
			selCache = append(selCache, item)
		}
	}
	selTodos := []*todos.Todo{}
	for _, item := range todoItems {
		if _, ok := keys[item.Key]; ok {
			selTodos = append(selTodos, item)
		}
	}
	return selCache, selTodos, nil
}

func itemsToSelectItems(cacheItems []*config.CacheItem, todoItems []*todos.Todo) []*ui.SelectItem {
	var selectItems []*ui.SelectItem
	for _, item := range cacheItems {
		selectItems = append(selectItems, &ui.SelectItem{
			Key:   item.Key,
			Label: fmt.Sprintf("[%s] %s: %s", item.Key, item.Type, item.Summary),
		})
	}

	for _, item := range todoItems {
		selectItems = append(selectItems, &ui.SelectItem{
			Key:   item.Key,
			Label: fmt.Sprintf("[%s] %s: %s", item.Key, item.Status, item.Title),
		})
	}

	return selectItems
}

func resolveTaskKeys(instance *app.App, keys []string) ([]*config.CacheItem, []*todos.Todo, error) {
	selectedCacheItems := []*config.CacheItem{}
	selectedTodos := []*todos.Todo{}

	todoKeys, providerKeys := separateKeys(keys)

	if len(providerKeys) > 0 {
		cacheItems, err := instance.State.Cache.GetByKeyBatch(providerKeys)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"unable to find tasks in cache. "+
					"Consider to run 'sneak list --refresh' to force a refresh.: %w",
				err,
			)
		}
		selectedCacheItems = append(selectedCacheItems, cacheItems...)
	}

	if len(todoKeys) > 0 {
		todos, err := instance.TodoStore.GetByRefBatch(
			instance.ProjectScope, todoKeys,
		)
		if err != nil {
			return nil, nil, err
		}
		selectedTodos = append(selectedTodos, todos...)
	}

	return selectedCacheItems, selectedTodos, nil
}

func ResolveTaskSelection(
	instance *app.App, focus TaskFocus,
	keys []string, all bool,
) ([]*config.CacheItem, []*todos.Todo, error) {

	// --all --> return the full focus pool depending on context.
	// We do not refresh the cache in this case to avoid acting on items
	// unknown to the user yet.
	if all {
		return resolveAll(instance, focus)
	}

	// Provider cache is only refreshed when the operation touches provider
	// items from within a project scope.
	if instance.InProjectScope() && ContainsProviderKeys(keys) {
		refreshRequired, err := CheckAndRefreshCache(instance, false)
		if refreshRequired && err != nil {
			return nil, nil, err
		}
	}

	// No keys → interactive multi-select over the focus pool.
	// In global context cacheItems stay empty.
	if len(keys) < 1 {
		cacheItems, todoItems, err := resolveAll(instance, focus)
		if err != nil {
			return nil, nil, err
		}

		selected, err := ui.InteractiveSelectItem(itemsToSelectItems(cacheItems, todoItems))
		if err != nil {
			return nil, nil, err
		}
		if len(selected) == 0 {
			return nil, nil, errors.New("no items selected")
		}

		selectedKeys := make(map[string]struct{}, len(selected))
		for _, s := range selected {
			selectedKeys[s.Key] = struct{}{}
		}
		return filterSelected(cacheItems, todoItems, selectedKeys)
	}

	// Explicitly supplied keys → resolve per domain. The focus is irrelevant
	// here: resolution is by key, not by pool.
	return resolveTaskKeys(instance, keys)
}

// resolveAll returns the full focus pool. Todos are always resolved; provider
// items are only resolved within a project scope.
func resolveAll(instance *app.App, focus TaskFocus) ([]*config.CacheItem, []*todos.Todo, error) {
	todoItems, err := instance.TodoStore.GetByStatus(
		instance.ProjectScope, focus.todoStatuses(),
	)
	if err != nil {
		return nil, nil, err
	}

	cacheItems := []*config.CacheItem{}
	if instance.InProjectScope() {
		cacheItems, err = focus.providerAll(instance.State)
		if err != nil {
			return nil, nil, err
		}
	} else {
		ui.Printfln("outside of project context. Provider work items will not be resolved.")
	}

	return cacheItems, todoItems, nil
}

// ResolveCommentTargets splits keys into provider and todo targets, applying
// the provider cache refresh gate. Provider items are only resolved within a
// project scope; todo refs resolve against the current scope bucket.
func ResolveCommentTargets(instance *app.App, keys []string) ([]*config.CacheItem, []*todos.Todo, error) {
	if instance.InProjectScope() && ContainsProviderKeys(keys) {
		refreshRequired, err := CheckAndRefreshCache(instance, false)
		if refreshRequired && err != nil {
			return nil, nil, err
		}
	}

	return resolveTaskKeys(instance, keys)
}

func ResolveShipTaskFocus(app *app.App, taskKeys []string, all bool) ([]*config.CacheItem, error) {
	return resolveTaskFocus(app, taskKeys, all, app.State.GetActiveCacheItems)
}

func resolveTaskFocus(
	app *app.App, taskKeys []string, all bool,
	candidateFunc func() ([]*config.CacheItem, error),
) ([]*config.CacheItem, error) {

	if len(taskKeys) > 0 {
		cacheItems, err := app.State.Cache.GetByKeyBatch(taskKeys)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to find tasks in cache. "+
					"Consider to run 'sneak list --refresh' to force a refresh.: %w",
				err,
			)
		}
		return cacheItems, nil
	}

	cacheItems, err := candidateFunc()
	if err != nil {
		return nil, err
	}

	if all {
		return cacheItems, nil
	}

	// Prompt user with interactive selection if no key was provided.
	selection, tuiSelectErr := ui.InteractiveSelectCacheItem(cacheItems)
	if tuiSelectErr != nil {
		return nil, fmt.Errorf("failed task selection.")
	}
	return selection, nil
}
