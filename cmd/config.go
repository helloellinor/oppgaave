package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage application configuration and settings",
	Long: `Configuration management commands for setting up user preferences,
API keys, work hours, contact preferences, and other application settings.

Features:
- User preference management
- API key configuration (OpenAI, etc.)
- Work hours and availability settings
- Default task and contact preferences
- Configuration file management
- Environment variable overrides

Examples:
  oppgaave config set work-hours "09:00-17:00"
  oppgaave config set api-key-openai "sk-..."
  oppgaave config get work-hours
  oppgaave config list
  oppgaave config reset --section "tasks"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configuration management - use subcommands: set, get, list, reset, init")
		cmd.Help()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set a configuration value for the application.
Supports nested keys using dot notation (e.g., tasks.default-priority).`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		global, _ := cmd.Flags().GetBool("global")
		
		fmt.Printf("Setting configuration: %s = %s\n", key, value)
		if global {
			fmt.Println("Setting as global configuration")
		}
		// TODO: Implement actual configuration setting
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Get a configuration value from the application settings.
If no key is provided, shows all configuration values.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Showing all configuration values...")
		} else {
			key := args[0]
			fmt.Printf("Getting configuration value for: %s\n", key)
		}
		// TODO: Implement actual configuration retrieval
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Long: `List all configuration settings with their current values.
Supports filtering by section and output formatting.`,
	Run: func(cmd *cobra.Command, args []string) {
		section, _ := cmd.Flags().GetString("section")
		format, _ := cmd.Flags().GetString("format")
		showDefaults, _ := cmd.Flags().GetBool("show-defaults")
		
		fmt.Println("Listing configuration settings...")
		if section != "" {
			fmt.Printf("Section filter: %s\n", section)
		}
		fmt.Printf("Output format: %s\n", format)
		if showDefaults {
			fmt.Println("Including default values")
		}
		// TODO: Implement actual configuration listing
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long: `Reset configuration settings to their default values.
Can reset specific sections or all settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		section, _ := cmd.Flags().GetString("section")
		all, _ := cmd.Flags().GetBool("all")
		confirm, _ := cmd.Flags().GetBool("confirm")
		
		fmt.Println("Resetting configuration...")
		if section != "" {
			fmt.Printf("Resetting section: %s\n", section)
		}
		if all {
			fmt.Println("Resetting all configuration")
		}
		if !confirm {
			fmt.Println("Use --confirm to actually reset configuration")
		}
		// TODO: Implement actual configuration reset
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration with interactive setup",
	Long: `Initialize application configuration with an interactive setup wizard.
Guides through setting up essential configuration like work hours, API keys, and preferences.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		template, _ := cmd.Flags().GetString("template")
		
		fmt.Println("Initializing configuration...")
		if force {
			fmt.Println("Force initialization (overwriting existing config)")
		}
		if template != "" {
			fmt.Printf("Using template: %s\n", template)
		}
		// TODO: Implement interactive configuration initialization
	},
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configInitCmd)

	// Config set flags
	configSetCmd.Flags().Bool("global", false, "Set as global configuration")
	configSetCmd.Flags().String("type", "auto", "Value type (string, int, bool, float)")

	// Config list flags
	configListCmd.Flags().String("section", "", "Filter by configuration section")
	configListCmd.Flags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	configListCmd.Flags().Bool("show-defaults", false, "Show default values for unset options")

	// Config reset flags
	configResetCmd.Flags().String("section", "", "Reset specific section only")
	configResetCmd.Flags().Bool("all", false, "Reset all configuration")
	configResetCmd.Flags().Bool("confirm", false, "Confirm the reset operation")

	// Config init flags
	configInitCmd.Flags().Bool("force", false, "Force initialization, overwriting existing config")
	configInitCmd.Flags().String("template", "", "Use configuration template (basic, advanced, developer)")
}
