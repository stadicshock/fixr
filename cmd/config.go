package cmd

import (
	"fmt"
	"os"

	"github.com/fixr-app/fixr/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage fixr configuration",
	Long:  "View or initialize the fixr configuration file at ~/.fixr/config.yaml",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default config file",
	Long:  "Create the default configuration file at ~/.fixr/config.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.WriteDefault()
		if err != nil {
			return err
		}
		fmt.Printf("Config file created at: %s\n", path)
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Edit the config file with your API keys")
		fmt.Println("  2. Run 'fixr fix' to fix clipboard text")
		fmt.Println("  3. Run 'fixr start' to launch the menu bar app")
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current configuration (API keys are masked)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Mask API keys for display
		for name, p := range cfg.Providers {
			if p.APIKey != "" {
				masked := p.APIKey
				if len(masked) > 8 {
					masked = masked[:4] + "..." + masked[len(masked)-4:]
				} else {
					masked = "****"
				}
				p.APIKey = masked
				cfg.Providers[name] = p
			}
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("%s (not found — run 'fixr config init')\n", path)
		} else {
			fmt.Println(path)
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}
