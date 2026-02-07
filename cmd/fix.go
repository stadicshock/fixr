package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fixr-app/fixr/internal/clipboard"
	"github.com/fixr-app/fixr/internal/config"
	"github.com/fixr-app/fixr/internal/engine"
	"github.com/spf13/cobra"
)

var (
	fixFile     string
	fixProvider string
	fixCopy     bool
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Fix grammar in clipboard text or a file",
	Long: `Fix grammar in text from the clipboard (default) or from a file.

When fixing clipboard text, fixr also captures a screenshot of your
active window to give the AI context about the conversation you're in.

Examples:
  fixr fix                    # Fix clipboard text
  fixr fix -f draft.txt       # Fix text in a file  
  fixr fix -p openai          # Use a specific provider
  fixr fix -c                 # Copy corrected text to clipboard`,
	RunE: runFix,
}

func init() {
	fixCmd.Flags().StringVarP(&fixFile, "file", "f", "", "Fix text from a file instead of clipboard")
	fixCmd.Flags().StringVarP(&fixProvider, "provider", "p", "", "AI provider to use (overrides config default)")
	fixCmd.Flags().BoolVarP(&fixCopy, "copy", "c", false, "Copy corrected text to clipboard")
	rootCmd.AddCommand(fixCmd)
}

func runFix(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	eng, err := engine.NewWithProvider(cfg, fixProvider)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var result *engine.FixResult

	if fixFile != "" {
		fmt.Fprintf(os.Stderr, "Fixing grammar in %s...\n", fixFile)
		result, err = eng.FixFile(ctx, fixFile)
	} else {
		fmt.Fprintf(os.Stderr, "Fixing grammar from clipboard...\n")
		result, err = eng.FixClipboard(ctx)
	}
	if err != nil {
		return err
	}

	// Print the corrected text to stdout
	fmt.Println(result.Corrected)

	// Copy to clipboard if requested
	if fixCopy || cfg.Preferences.AutoPaste {
		if err := clipboard.Write(result.Corrected); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not copy to clipboard: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Corrected text copied to clipboard.\n")
		}
	}

	fmt.Fprintf(os.Stderr, "(provider: %s)\n", result.Provider)
	return nil
}
