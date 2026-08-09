package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/richo542/sneak/internal/config"
	"github.com/richo542/sneak/internal/ui"
	"github.com/spf13/cobra"
)

func newListCmd(app *App) *cobra.Command {
	var (
		refresh    bool
		typeFilter string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work items assigned to you",
		Long: `Displays work items under the configured bindings.

Uses a local cache (1hr TTL) for fast results.
Use --refresh to force a live fetch from the provider.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if app.LocalContext == nil {
				return fmt.Errorf("not initialized: run 'sneak init' first")
			}

			return runList(app, refresh, typeFilter)
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "force live fetch from provider")
	cmd.Flags().StringVarP(&typeFilter, "types", "t", "", "filter by work item type (e.g. Story, Bug)")

	return cmd
}

func runList(app *App, refresh bool, typeFilter string) error {

	refreshRequired, err := CheckAndRefreshCache(app, refresh)
	if refreshRequired && err != nil {
		return err
	}

	state := app.State
	// Shouldn't happen as setting bindings is mandatory during init,
	// however, keeping it as a saftey net.
	if len(state.Cache.Bindings) == 0 {
		return fmt.Errorf("No bindings set, please run 'sneak init' first for setup.")
	}

	items := state.Cache.Items

	types := parseTypes(typeFilter)
	if len(types) > 0 {
		items = filterByType(items, types)
	}

	if len(items) == 0 {
		fmt.Println("No work items found.")
		return nil
	}

	ui.PrintTableOfItems(items)

	fmt.Printf("\n%d work items", len(items))

	// Log out the age of this state for keep user informed
	age := time.Since(state.Cache.FetchedAt).Truncate(time.Second)
	fmt.Printf(" (cached, fetched %s ago)", age)
	fmt.Println()

	return nil
}

func parseTypes(typeFilter string) []string {
	var types []string
	typesSplits := strings.SplitSeq(typeFilter, ",")
	for p := range typesSplits {
		if p = strings.TrimSpace(p); p != "" {
			types = append(types, p)
		}
	}

	return types
}

func filterByType(items []config.CacheItem, types []string) []config.CacheItem {
	var filtered []config.CacheItem
	for _, item := range items {
		for _, t := range types {
			if strings.EqualFold(item.Type, t) {
				filtered = append(filtered, item)
			}
		}
	}
	return filtered
}
