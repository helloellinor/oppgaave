package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// calendarCmd represents the calendar command
var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Manage calendar events and scheduling",
	Long: `Calendar management commands for creating, viewing, and managing events.
Supports recurring events, event conflicts detection, and integration with task scheduling.

Examples:
  oppgaave calendar add "Team meeting" --date "2024-01-15" --time "14:00" --duration "1h"
  oppgaave calendar list --month "2024-01"
  oppgaave calendar remove --id "event-123"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Calendar management - use subcommands: add, list, remove, edit")
		cmd.Help()
	},
}

var calendarAddCmd = &cobra.Command{
	Use:   "add [event title]",
	Short: "Add a new calendar event",
	Long: `Add a new event to the calendar with specified date, time, and duration.
Supports recurring events and automatic conflict detection.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		date, _ := cmd.Flags().GetString("date")
		time, _ := cmd.Flags().GetString("time")
		duration, _ := cmd.Flags().GetString("duration")
		recurring, _ := cmd.Flags().GetString("recurring")
		
		fmt.Printf("Adding calendar event: %s\n", title)
		fmt.Printf("Date: %s, Time: %s, Duration: %s\n", date, time, duration)
		if recurring != "" {
			fmt.Printf("Recurring: %s\n", recurring)
		}
		// TODO: Implement actual calendar event creation
	},
}

var calendarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List calendar events",
	Long: `List calendar events for a specified time period.
Can filter by date range, event type, or search terms.`,
	Run: func(cmd *cobra.Command, args []string) {
		month, _ := cmd.Flags().GetString("month")
		week, _ := cmd.Flags().GetString("week")
		day, _ := cmd.Flags().GetString("day")
		
		fmt.Println("Listing calendar events...")
		if month != "" {
			fmt.Printf("Month filter: %s\n", month)
		}
		if week != "" {
			fmt.Printf("Week filter: %s\n", week)
		}
		if day != "" {
			fmt.Printf("Day filter: %s\n", day)
		}
		// TODO: Implement actual calendar event listing
	},
}

var calendarRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a calendar event",
	Long: `Remove a calendar event by ID or by matching criteria.
Supports removing single instances or entire recurring series.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		title, _ := cmd.Flags().GetString("title")
		
		fmt.Println("Removing calendar event...")
		if id != "" {
			fmt.Printf("Event ID: %s\n", id)
		}
		if title != "" {
			fmt.Printf("Event title: %s\n", title)
		}
		// TODO: Implement actual calendar event removal
	},
}

func init() {
	// Add subcommands
	calendarCmd.AddCommand(calendarAddCmd)
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarRemoveCmd)

	// Calendar add flags
	calendarAddCmd.Flags().StringP("date", "d", "", "Event date (YYYY-MM-DD)")
	calendarAddCmd.Flags().StringP("time", "t", "", "Event time (HH:MM)")
	calendarAddCmd.Flags().String("duration", "1h", "Event duration (e.g., 1h, 30m)")
	calendarAddCmd.Flags().StringP("recurring", "r", "", "Recurring pattern (daily, weekly, monthly)")
	calendarAddCmd.Flags().String("location", "", "Event location")
	calendarAddCmd.Flags().String("description", "", "Event description")

	// Calendar list flags
	calendarListCmd.Flags().String("month", "", "Filter by month (YYYY-MM)")
	calendarListCmd.Flags().String("week", "", "Filter by week (YYYY-WW)")
	calendarListCmd.Flags().String("day", "", "Filter by day (YYYY-MM-DD)")
	calendarListCmd.Flags().StringP("format", "f", "table", "Output format (table, json, csv)")

	// Calendar remove flags
	calendarRemoveCmd.Flags().String("id", "", "Event ID to remove")
	calendarRemoveCmd.Flags().String("title", "", "Event title to match")
	calendarRemoveCmd.Flags().Bool("all-recurring", false, "Remove all instances of recurring event")
}
