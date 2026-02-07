package capture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ActiveWindow captures a screenshot of the currently active (frontmost) window
// on macOS using the native screencapture command. Returns the PNG image bytes.
func ActiveWindow() ([]byte, error) {
	// Get the window ID of the frontmost window using AppleScript
	windowID, err := getFrontmostWindowID()
	if err != nil {
		// Fall back to capturing the entire screen if we can't get the window ID
		return captureScreen()
	}

	tmpFile := filepath.Join(os.TempDir(), "fixr-capture.png")
	defer os.Remove(tmpFile)

	// -x: no sound, -l: capture specific window by ID
	cmd := exec.Command("screencapture", "-x", "-l", strconv.Itoa(windowID), tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("capture: screencapture failed: %w", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("capture: failed to read screenshot: %w", err)
	}

	return data, nil
}

// captureScreen captures the entire main screen as a fallback.
func captureScreen() ([]byte, error) {
	tmpFile := filepath.Join(os.TempDir(), "fixr-capture.png")
	defer os.Remove(tmpFile)

	// -x: no sound, -m: main monitor only
	cmd := exec.Command("screencapture", "-x", "-m", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("capture: screencapture failed: %w", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("capture: failed to read screenshot: %w", err)
	}

	return data, nil
}

// getFrontmostWindowID returns the CGWindowID of the frontmost window.
func getFrontmostWindowID() (int, error) {
	script := `
		tell application "System Events"
			set frontApp to name of first application process whose frontmost is true
		end tell
		tell application "System Events"
			tell process frontApp
				set winID to id of front window
			end tell
		end tell
		return winID
	`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		// Fallback: use CGWindowListInfo via python to get frontmost window ID
		return getFrontmostWindowIDFallback()
	}

	idStr := strings.TrimSpace(string(out))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("capture: invalid window ID %q: %w", idStr, err)
	}
	return id, nil
}

// getFrontmostWindowIDFallback uses a simpler AppleScript approach.
func getFrontmostWindowIDFallback() (int, error) {
	// Get the frontmost app's name and use screencapture with app name approach
	script := `tell application "System Events" to return name of first application process whose frontmost is true`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("capture: could not determine frontmost app: %w", err)
	}

	appName := strings.TrimSpace(string(out))
	_ = appName // We'll fall back to full screen capture if we can't get the window ID

	return 0, fmt.Errorf("capture: could not determine window ID for %q", appName)
}
