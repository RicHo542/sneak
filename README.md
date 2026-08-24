<p align="center">
  <img src="assets/logo.svg" alt="sneak" width="160">
</p>

# sneak

Sneak lets you claim, update and close tasks - User Stories, Bugs and more -
right from the terminal. It reduces organizational red tape for work item tracking
and allows you to manage your tasks right where you are - the terminal.

## Commands

- `sneak config` - Set up your provider connection and user handle.
- `sneak init` - Initialise the project context in `.sneak/` and discover the workflow.
- `sneak start KEY...` - Start work on one or more tasks and track them as active.
- `sneak close KEY...` - Transition tasks to done and stop tracking them.

## Examples

Initially setup a provider (azure/jira), by running the config command and following the setup wizard

```bash
sneak config
```

Initialize sneak in your current directory (root of your repository ideally)

```bash
sneak init
```

List the work items in your project:

```bash
$ sneak list
KEY           ASSIGNED    TYPE         STATUS           SUMMARY
--------------------------------------------------------------------------------
STORY-42      x           Story        In Progress      Refactor the login flow
BUG-7                     Bug          Open             Fix flaky test runner
```

Start several tasks at once and create a feature branch for them:

```bash
sneak start STORY-42 BUG-7 -b
```

Close every task you are currently tracking:

```bash
sneak close -a
```

Run `sneak --help` for the full command list.
