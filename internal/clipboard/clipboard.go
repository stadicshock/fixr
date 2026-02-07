package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

// Read returns the current text content of the system clipboard.
// Uses macOS pbpaste for reliability (no CGo dependency).
func Read() (string, error) {
	out, err := exec.Command("pbpaste").Output()
	if err != nil {
		return "", fmt.Errorf("clipboard: failed to read (pbpaste): %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// Write sets the system clipboard to the given text.
// Uses macOS pbcopy for reliability (no CGo dependency).
func Write(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard: failed to write (pbcopy): %w", err)
	}
	return nil
}
