package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks with dependencies and AI-powered breakdown",
	Long: `Task management commands for creating, organizing, and tracking tasks.
Supports hierarchical task structures, dependency management, time tracking,
and AI-powered task breakdown for complex projects.

Features:
- Recursive task hierarchies (tasks with subtasks)
- Task dependencies and requirement validation
- Time estimation and tracking
- AI-powered task breakdown
- Contact associations and scheduling
- Recurring task patterns

Examples:
  oppgaave task create "Build website" --priority high --estimate "2w"
  oppgaave task breakdown "Build website" --ai
  oppgaave task list --status pending --priority high
  oppgaave task track start --id "task-123"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Task management - use subcommands: create, list, edit, remove, breakdown, track")
		cmd.Help()
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create [task title]",
	Short: "Create a new task",
	Long: `Create a new task with optional dependencies, requirements, and scheduling.
Supports hierarchical task creation and automatic dependency validation.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetString("priority")
		estimate, _ := cmd.Flags().GetString("estimate")
		parent, _ := cmd.Flags().GetString("parent")
		dependencies, _ := cmd.Flags().GetStringSlice("depends-on")
		contacts, _ := cmd.Flags().GetStringSlice("contacts")
		
		fmt.Printf("Creating task: %s\n", title)
		if description != "" {
			fmt.Printf("Description: %s\n", description)
		}
		fmt.Printf("Priority: %s, Estimate: %s\n", priority, estimate)
		if parent != "" {
			fmt.Printf("Parent task: %s\n", parent)
		}
		if len(dependencies) > 0 {
			fmt.Printf("Dependencies: %v\n", dependencies)
		}
		if len(contacts) > 0 {
			fmt.Printf("Associated contacts: %v\n", contacts)
		}
		// TODO: Implement actual task creation
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks with filtering options",
	Long: `List tasks with various filtering and sorting options.
Supports hierarchical view, dependency visualization, and status filtering.`,
	Run: func(cmd *cobra.Command, args []string) {
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		parent, _ := cmd.Flags().GetString("parent")
		tree, _ := cmd.Flags().GetBool("tree")
		
		fmt.Println("Listing tasks...")
		if status != "" {
			fmt.Printf("Status filter: %s\n", status)
		}
		if priority != "" {
			fmt.Printf("Priority filter: %s\n", priority)
		}
		if parent != "" {
			fmt.Printf("Parent task: %s\n", parent)
		}
		if tree {
			fmt.Println("Tree view enabled")
		}
		// TODO: Implement actual task listing
	},
}

var taskBreakdownCmd = &cobra.Command{
	Use:   "breakdown [task-id]",
	Short: "Break down a task into subtasks using AI",
	Long: `Use AI to automatically break down a complex task into manageable subtasks.
Considers dependencies, requirements, and optimal task sequencing.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		ai, _ := cmd.Flags().GetBool("ai")
		interactive, _ := cmd.Flags().GetBool("interactive")
		
		fmt.Printf("Breaking down task: %s\n", taskID)
		if ai {
			fmt.Println("Using AI-powered breakdown")
		}
		if interactive {
			fmt.Println("Interactive mode enabled")
		}
		// TODO: Implement actual task breakdown
	},
}

var taskTrackCmd = &cobra.Command{
	Use:   "track [action]",
	Short: "Track time for tasks (start, stop, pause, resume)",
	Long: `Time tracking commands for monitoring task progress.
Supports automatic time tracking across task hierarchies and dependencies.`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"start", "stop", "pause", "resume", "status"},
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]
		taskID, _ := cmd.Flags().GetString("id")
		
		fmt.Printf("Time tracking action: %s\n", action)
		if taskID != "" {
			fmt.Printf("Task ID: %s\n", taskID)
		}
		// TODO: Implement actual time tracking
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit [task-id]",
	Short: "Edit an existing task",
	Long: `Edit task properties including title, description, priority, dependencies, and requirements.
Supports interactive editing and dependency validation.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetString("priority")
		
		fmt.Printf("Editing task: %s\n", taskID)
		if title != "" {
			fmt.Printf("New title: %s\n", title)
		}
		if description != "" {
			fmt.Printf("New description: %s\n", description)
		}
		if priority != "" {
			fmt.Printf("New priority: %s\n", priority)
		}
		// TODO: Implement actual task editing
	},
}

var taskRemoveCmd = &cobra.Command{
	Use:   "remove [task-id]",
	Short: "Remove a task and handle dependencies",
	Long: `Remove a task while properly handling dependencies and subtasks.
Supports cascading removal and dependency reassignment.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		cascade, _ := cmd.Flags().GetBool("cascade")
		force, _ := cmd.Flags().GetBool("force")
		
		fmt.Printf("Removing task: %s\n", taskID)
		if cascade {
			fmt.Println("Cascade removal enabled")
		}
		if force {
			fmt.Println("Force removal enabled")
		}
		// TODO: Implement actual task removal
	},
}

func init() {
	// Add subcommands
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskBreakdownCmd)
	taskCmd.AddCommand(taskTrackCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskRemoveCmd)

	// Task create flags
	taskCreateCmd.Flags().StringP("description", "d", "", "Task description")
	taskCreateCmd.Flags().StringP("priority", "p", "medium", "Task priority (low, medium, high, urgent)")
	taskCreateCmd.Flags().StringP("estimate", "e", "", "Time estimate (e.g., 2h, 1d, 1w)")
	taskCreateCmd.Flags().String("parent", "", "Parent task ID")
	taskCreateCmd.Flags().StringSlice("depends-on", []string{}, "Task dependencies (comma-separated IDs)")
	taskCreateCmd.Flags().StringSlice("contacts", []string{}, "Associated contacts (comma-separated)")
	taskCreateCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")
	taskCreateCmd.Flags().String("recurring", "", "Recurring pattern (daily, weekly, monthly)")

	// Task list flags
	taskListCmd.Flags().String("status", "", "Filter by status (pending, in-progress, completed, blocked)")
	taskListCmd.Flags().String("priority", "", "Filter by priority (low, medium, high, urgent)")
	taskListCmd.Flags().String("parent", "", "Filter by parent task ID")
	taskListCmd.Flags().BoolP("tree", "t", false, "Show hierarchical tree view")
	taskListCmd.Flags().StringP("format", "f", "table", "Output format (table, json, tree)")
	taskListCmd.Flags().Bool("dependencies", false, "Show task dependencies")

	// Task breakdown flags
	taskBreakdownCmd.Flags().Bool("ai", false, "Use AI for task breakdown")
	taskBreakdownCmd.Flags().BoolP("interactive", "i", false, "Interactive breakdown mode")
	taskBreakdownCmd.Flags().Int("max-depth", 3, "Maximum breakdown depth")

	// Task track flags
	taskTrackCmd.Flags().String("id", "", "Task ID to track")
	taskTrackCmd.Flags().String("note", "", "Add a note to the time entry")

	// Task edit flags
	taskEditCmd.Flags().String("title", "", "New task title")
	taskEditCmd.Flags().String("description", "", "New task description")
	taskEditCmd.Flags().String("priority", "", "New task priority")
	taskEditCmd.Flags().String("estimate", "", "New time estimate")
	taskEditCmd.Flags().StringSlice("add-deps", []string{}, "Add dependencies")
	taskEditCmd.Flags().StringSlice("remove-deps", []string{}, "Remove dependencies")

	// Task remove flags
	taskRemoveCmd.Flags().Bool("cascade", false, "Remove subtasks as well")
	taskRemoveCmd.Flags().Bool("force", false, "Force removal even with dependencies")
}
