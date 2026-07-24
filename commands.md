## Global configuration for different providers (not in project repo)

- Right now only for listing, will be useful later

```bash
sneak config
--list -> List available providers with working status
```

```bash
sneak login
-> cli process to setup/update a provider (host, alias, user, pass, pat)
```

## Initializing new sneak config in project directory

```bash
sneak init <path>
    --provider <provider> -> For which provider
    --ids <ids> -> Epic/Feature Ids to link as parent
```

## List all tasks

```bash
sneak list
--type -t -> Filter by type
--refresh
```

## Shows the currently active tasks

```bash
sneak status
```

## Start task

- Assigns, if not already
- Moves to in progress / active
- Optionally comments and creates a feature branch for it

```bash
sneak start <task1> <task2> <task3> (cli selector, if none)
-m -> Optional: Comment to add to the started tasks
-b -> Optional: Create a branch with taskids and checkout
```

## Complete task

- Will set task states to closed / completed
- How to handle the branch here? Only accept -b without or just with branchname-matching task ids? -> Sounds fuzzy

```bash
sneak close <task1> <task2> <task3> (cli selector, if none)
-m -> Will add comment to the task to be completed.
-b -> Also create a PR for the branch
```

## Unassign a task again to "Unassigned"

- Only possible when currently assigned.
- Sets assignee back to unassigned
- Sets state back to Open

```bash
sneak unassign <task1> <task2> <task3>
```

## Just adding a comment to the task

```bash
sneak comment <task>
-m -> Comment to make
```
