Here is a complete summary of sneak—the vision, the core mechanics, and the key decisions made along the way.
🦝 Tool Overview: sneak
Elevator Pitch: A lightweight, developer-first CLI tool designed to eliminate corporate red tape, bypass heavy web UI dashboards, and keep developers in their terminal flow state by seamlessly automating task management overhead (Jira, Azure Boards, GitHub Issues).
Core Vision & Value Proposition
Zero Web UI Drag: Developers pick up tasks, create branches, update ticket statuses, and open PRs without ever opening a web browser or filling out manual forms.
Flow State Preservation: Shifts the administrative burden into background automation triggered by natural developer actions (e.g., git commits and PRs).
Developer Independence: Lets devs focus entirely on building software while automatically generating a complete audit trail of ticket movements for managers and project tracking.
🛠️ Typical Command Workflow

Bash
sneak init # Binds the local repository to an Epic or Area Path
sneak start # Fuzzy-selects a story, checks out a branch, sets "In Progress"
sneak ship # Opens PR, links commits, moves ticket to "Done", and wraps up

📌 Summary of Key Design Decisions

1. Naming Philosophy & Selection
   The Goal: Single-word, 3-to-5-letter names that are fast to type and evoke a sense of speed, stealth, or clearing process friction.
   Why sneak Won:
   It captures the playful, under-the-radar vibe of sneaking past corporate bureaucracy to focus on actual code.
   It's memorable, short, and fun on the command line.
2. Ecosystem & Collision Checks
   System Binaries: We avoided squid due to conflicts with the widely pre-installed Linux HTTP caching proxy server.
   Major Platforms: We avoided flow due to heavy collisions with Meta’s JavaScript Type Checker (flow) and the Flow Blockchain CLI.
   Category Clutter: We avoided snip because it is already ubiquitous for code-snippet managers (phlx0/snip, snippet downloaders).
   sneak Status: Fully clear of major global executable collisions. (Note: While it sounds like the security tool Snyk, its binary command is uniquely spelled sneak).
3. Visual Identity (Brand & Logo)
   Theme: Features a sleek raccoon mascot—the ultimate symbol of being clever, nimble, and operating quietly in the background.
   Palette: Terminal-dark aesthetics with modern neon green/cyan accents and crisp code typography (sneak\_).
   What’s Next?
   Now that the identity, logo, name, and CLI ergonomics are locked in, you have a rock-solid foundation to start scaffolding the project architecture (API connectors for Jira/Azure, terminal UI, and Git hooks)!

Here is the updated, comprehensive summary incorporating the initial system design details, command structure, and end-to-end flow we mapped out.
🦝 Complete Project Summary: sneak
Elevator Pitch: A lightweight, developer-first CLI tool designed to eliminate corporate red tape, bypass heavy web UI dashboards, and keep developers in their terminal flow state by seamlessly automating task management overhead (Jira, Azure Boards, GitHub Issues).
📐 System Design & Architecture
sneak is designed to act as an invisible integration bridge between your local terminal session and external enterprise task management systems.
┌─────────────────────────────────┐
│ Local Terminal │
│ (sneak CLI + Git Workspace) │
└───────────────┬─────────────────┘
│
[sneak Orchestration]
│
┌───────────────────────┼───────────────────────┐
▼ ▼ ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│ Jira Cloud / │ │ Azure Boards │ │ GitHub / GitLab │
│ Data Center │ │ (ADO) │ │ PR Engine │
└──────────────────┘ └──────────────────┘ └──────────────────┘

Central Point of Integration: Like a octopus reaching out, sneak hooks into various external APIs simultaneously to grab context and manage state while maintaining a tiny footprint locally.
Context Persistence: Maintains lightweight local state (.sneak/ or global config) to remember active Epics, assigned tickets, and workspace parameters so you never re-enter flags.
Zero-Friction Execution: Operates entirely within your shell session, intercepting Git operations to execute status transitions silently in the background.
🔄 End-to-End Workflow
Context Initialization: You link your local repository to a parent Epic, Area Path, or Board milestone once.
Interactive Selection: You fetch assigned or unassigned tasks directly inside the terminal via an interactive selector—no browser tab needed.
Automated Workspace Setup: sneak automatically creates a standardized Git branch (e.g., feat/PROJ-123-short-description) and shifts the ticket status to In Progress.
Flow State Coding: You write code and commit as usual. sneak can auto-append issue keys or metadata to commits if configured.
Seamless Handoff & Wrap-up: A single command creates the Pull Request, links commit history to the issue, moves the ticket to In Review / Done, and cleans up local branches.
🛠️ CLI Command Reference
The ergonomics rely on short, expressive action verbs:
Command
Purpose & Action
sneak init
Binds the local repository to a target Jira/Azure Board, Epic, or milestone.
sneak start
Launches an interactive fuzzy selector to pick a task, checks out a branch, and moves the ticket to In Progress.
sneak status
Shows your current active task, linked branch, and remote ticket state directly in the shell.
sneak ink
(Optional) Quickly appends progress notes or comments to the remote ticket without leaving the terminal.
sneak ship
Pushes code, opens a Pull Request, updates the ticket to Done or In Review, and cleans up workspace state.

📌 Summary of Naming & Identity Discussions
The Problem Solved: Devs spend too much time navigating heavy web boards, manually updating ticket statuses, and creating branches.
Name Selection (sneak): Chosen over names like squid (collides with the Linux Squid proxy), flow (collides with Meta Flow / Flow Blockchain), and snip (crowded with code-snippet vault tools). sneak is clear of binary collisions and perfectly embodies quietly bypassing process drag.
Visual Identity: Represented by a sleek raccoon mascot—symbolizing speed, cleverness, and working quietly in the background to get things done.

Here is the updated System Design, Architecture, and Workflows section, updated to clearly highlight those two distinct operational modes:
🔀 The Two Core Workflows
sneak is designed around two main ways developers interact with backlogs in their day-to-day work:

1. The "Backlog Discovery" Mode (Claim, Context & Start)
   Best for: Picking up new work from an Epic, milestone, or team queue.
   Discover & Claim: You run sneak start to launch an interactive search across unassigned or team-level open tickets.
   Context Binding: sneak assigns the selected ticket to you, sets the local repository context to that ticket, and automatically branches out (feat/PROJ-123-description).
   Status Sync: Moves the remote ticket status from To Do to In Progress in the background.
   Bash
   sneak start --unassigned # Fuzzy-finds from backlog, assigns to you, makes branch

2. The "Fast-Track / Standup" Mode (Sprint Execution)
   Best for: Knocking out tasks already assigned to you during sprint execution.
   Instant Fetch: You run sneak list (or sneak start) to view a lightweight list of tasks already on your plate for the active sprint.
   One-Shot Execution: Select the task directly in the terminal to jump straight into the branch without navigating menus or filling forms.
   Quick Wrap-Up: Once complete, run sneak ship or sneak close to transition the task directly to Done or In Reviewright from where you are, without ever breaking terminal flow.
   Bash
   sneak list # Shows your assigned tasks in 0.1s
   sneak start PROJ-123 # Jumps right into the assigned task
   sneak ship # Opens PR and checks off the ticket on the spot

📐 System Design & Architecture
sneak acts as an invisible integration bridge between your local terminal session and external enterprise task management systems.
┌─────────────────────────────────┐
│ Local Terminal │
│ (sneak CLI + Git Workspace) │
└───────────────┬─────────────────┘
│
[sneak Orchestration]
│
┌───────────────────────┼───────────────────────┐
▼ ▼ ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│ Jira Cloud / │ │ Azure Boards │ │ GitHub / GitLab │
│ Data Center │ │ (ADO) │ │ PR Engine │
└──────────────────┘ └──────────────────┘ └──────────────────┘

Central Point of Integration: Like an octopus reaching out, sneak hooks into various external APIs simultaneously to grab context and manage state while maintaining a tiny footprint locally.
Context Persistence: Maintains lightweight local state (.sneak/ or global config) to remember active Epics, assigned tickets, and workspace parameters so you never re-enter flags.
Zero-Friction Execution: Operates entirely within your shell session, intercepting Git operations to execute status transitions silently in the background.
🛠️ CLI Command Reference
The ergonomics rely on short, expressive action verbs supporting both workflows:
Command
Purpose & Action
sneak init
Binds the local repository to a target Jira/Azure Board, Epic, or milestone.
sneak list
Fast-Track: Quickly displays tickets already assigned to you for fast selection.
sneak start
Discovery or Start: Fuzzy-selects an open backlog item or assigned task, assigns it, creates a branch, and sets status to In Progress.
sneak status
Shows your active task context, linked branch, and remote ticket state directly in the shell.
sneak ship
Pushes code, opens a Pull Request, updates the ticket to Done or In Review, and cleans up workspace state.

📌 Summary of Naming & Identity
Name (sneak): Chosen because it is clear of system binary collisions, avoids overcrowded spaces (like code snippet tools or typecheckers), and captures the playful, under-the-radar feeling of sneaking past administrative drag.
Visual Identity: Modern neon green typography and a sleek raccoon mascot—symbolizing speed, cleverness, and working quietly in the background.

Here is a complete architecture blueprint, package recommendations, and directory breakdown to get sneak off the ground in Go.
📦 1. Key Go Packages You'll Need
Go is the absolute king for CLI binaries because it compiles to a single, zero-dependency binary with an ultra-fast startup time. Here are the premier libraries for building this:
CLI Framework & Interactivity
spf13/cobra — The industry standard CLI framework (used by kubectl, gh, docker). Handles subcommands (sneak start, sneak ship), flags, and auto-generated help menus.
charmbracelet/bubbletea — Terminal User Interface (TUI) engine. Used to build the interactive task selectors, fuzzy filters, and spinners.
charmbracelet/lipgloss — Terminal styling and layout. Makes your terminal output look gorgeous (colors, borders, tables) with zero layout headaches.
Git & API Integration
go-git/go-git — Pure Go implementation of Git. Allows sneak to inspect branches, check status, and create feature branches natively without spawning child processes (though calling exec.Command("git", ...) as a fallback is also common practice).
spf13/viper — Pairs seamlessly with Cobra to load configuration files (YAML/JSON) from both global (~/.config/sneak/) and local (.sneak/) directories.
cli/go-gh — Official GitHub API bindings if you want to automate PR creation natively on GitHub.
📂 2. Deep Dive: The .sneak/ Directory
To make sneak feel instant, it needs context persistence. You shouldn't have to pass --epic=EPIC-101 or --board=42every time you run a command.
The strategy uses a two-tier configuration model:
~/.config/sneak/config.yaml <-- Global Config (Auth Tokens, Defaults)
your-project/.sneak/ <-- Local Repository Config (Project Context)
├── config.yaml <-- Epic bindings, board IDs, transition mappings
└── state.json <-- Active task state (temporary workspace session)

A. .sneak/config.yaml (Checked into Git)
This file is shared across the dev team. When a developer clones the repo and runs sneak, it already knows which Jira Project / Azure Board / GitHub Repo it belongs to.
YAML

# .sneak/config.yaml

version: 1
project:
provider: "jira" # azure | jira | github
key: "PROJ" # Board or Project Key
board_id: "1042"

bindings:
epic_key: "PROJ-89" # Default parent Epic for this codebase
area_path: "Core/Backend" # Azure Boards Area Path

workflow:
default_branch_prefix: "feat/"
transitions: # Maps sneak actions to remote status names
start: "In Progress"
ship: "In Review"
close: "Done"

B. .sneak/state.json (Added to .gitignore)
This holds the local machine's active ticket session. It’s how sneak remembers what you are currently working on without querying Jira over HTTP every time you type sneak status.
JSON
{
"active_task": {
"key": "PROJ-123",
"summary": "Implement rate limiting middleware",
"branch": "feat/PROJ-123-rate-limiting",
"started_at": "2026-07-22T15:00:00Z"
}
}

🏗️ 3. Go Project Skeleton Architecture
Here is the clean, modular layout following the standard Go project layout pattern:
Plaintext
sneak/
├── cmd/
│ ├── root.go # Base 'sneak' command setup
│ ├── init.go # 'sneak init' (creates .sneak/config.yaml)
│ ├── list.go # 'sneak list' (fast-track assigned tasks)
│ ├── start.go # 'sneak start' (discovery or pick task)
│ ├── status.go # 'sneak status' (inspect active context)
│ └── ship.go # 'sneak ship' (PR + transition ticket)
├── pkg/
│ ├── config/ # Config loader (handles global vs .sneak/ merge)
│ ├── git/ # Git wrapper (branch checkout, commit, status)
│ ├── tracker/ # Interface for issue tracking services
│ │ ├── tracker.go # Unified interface (Client)
│ │ ├── jira.go # Jira API client
│ │ └── azure.go # Azure Boards API client
│ └── ui/ # Bubbletea TUI components & Lipgloss styles
│ ├── selector.go # Interactive fuzzy list for tasks
│ └── styles.go # Neon cyan/green color palette
├── .sneak.example.yaml # Template file
├── go.mod
└── main.go # Entrypoint

💻 4. Code Blueprints
A. Core Tracker Interface (pkg/tracker/tracker.go)
Define a single interface so the rest of your app doesn't care whether you're using Jira, Azure Boards, or GitHub.
Go
package tracker

type Task struct {
ID string
Key string // e.g., "PROJ-123"
Summary string
Status string
Assignee string
}

type Client interface {
// GetTasksAssignedToMe returns sprint tasks already assigned
GetTasksAssignedToMe() ([]Task, error)
// SearchBacklog returns unassigned tasks under the current Epic
SearchBacklog(epicKey string) ([]Task, error)
// AssignAndStart assigns the ticket and moves status to "In Progress"
AssignAndStart(taskKey string) error
// TransitionTask updates ticket status (e.g. to "Done" or "In Review")
TransitionTask(taskKey string, targetStatus string) error
}

B. Command Skeleton (cmd/start.go)
Using Cobra and the tracker interface to orchestrate the start workflow:
Go
package cmd

import (
"fmt"
"github.com/spf13/cobra"
"sneak/pkg/config"
"sneak/pkg/git"
"sneak/pkg/tracker"
"sneak/pkg/ui"
)

var startCmd = &cobra.Command{
Use: "start [task-key]",
Short: "Pick up a task, checkout a branch, and set ticket to In Progress",
RunE: func(cmd \*cobra.Command, args []string) error {
cfg, err := config.LoadLocal()
if err != nil {
return fmt.Errorf("repo not initialized, run 'sneak init' first: %w", err)
}

    	tr, err := tracker.NewClient(cfg)
    	if err != nil {
    		return err
    	}

    	var selectedTask tracker.Task

    	if len(args) > 0 {
    		// Fast-track: Passed ticket explicitly e.g. `sneak start PROJ-123`
    		selectedTask = tracker.Task{Key: args[0]}
    	} else {
    		// Interactive Mode: Load backlog or assigned tasks into Bubbletea TUI
    		tasks, _ := tr.GetTasksAssignedToMe()
    		selectedTask, err = ui.RunTaskSelector(tasks)
    		if err != nil {
    			return err
    		}
    	}

    	// 1. Assign & update status remotely
    	fmt.Printf("🦝 Moving %s to In Progress...\n", selectedTask.Key)
    	if err := tr.AssignAndStart(selectedTask.Key); err != nil {
    		return err
    	}

    	// 2. Create local Git branch
    	branchName := fmt.Sprintf("%s%s", cfg.Workflow.DefaultBranchPrefix, selectedTask.Key)
    	if err := git.CreateAndCheckoutBranch(branchName); err != nil {
    		return err
    	}

    	// 3. Save active state to .sneak/state.json
    	return config.SaveState(selectedTask, branchName)
    },

}

C. The Main Entrypoint (main.go)
Go
package main

import (
"fmt"
"os"
"sneak/cmd"
)

func main() {
if err := cmd.Execute(); err != nil {
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
os.Exit(1)
}
}

🚀 Recommended First Implementation Milestone
Phase 1 (Mock API): Implement cmd/root.go, cmd/init.go, and write a mock.go implementation for tracker.Client. Get the Bubbletea TUI list selector working locally so you can feel the command-line speed.
Phase 2 (Git Integration): Connect go-git or local exec.Command("git", ...) to automate branch creation and local state saving (.sneak/state.json).
Phase 3 (Jira/Azure Hookup): Implement authentic HTTP REST client calls for your team's issue tracker of choice!
