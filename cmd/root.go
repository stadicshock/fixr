package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fixr",
	Short: "fixr — Private, context-aware grammar assistant",
	Long: `fixr fixes grammar anywhere you write using your own AI API keys.
Copy text, run fixr, paste the corrected version.

Privacy-first: your text is only sent to the AI provider YOU configure.
No free-tier services, no data collection, no third-party access.

Usage:
  fixr fix          Fix text from clipboard (with auto-screenshot context)
  fixr fix -f file  Fix text from a file
  fixr config init  Create default config at ~/.fixr/config.yaml`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
