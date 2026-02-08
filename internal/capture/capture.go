package capture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// maxImageSize is the maximum image size in bytes (4MB, under the 5MB API limit).
const maxImageSize = 4 * 1024 * 1024

// ActiveWindow captures a screenshot of the currently active (frontmost) window
// on macOS using the native screencapture command. Returns JPEG image bytes,
// resized to stay under API size limits.
func ActiveWindow() ([]byte, error) {
	// Get the window ID of the frontmost window using AppleScript
	windowID, err := getFrontmostWindowID()
	if err != nil {
		// Fall back to capturing the entire screen if we can't get the window ID
		return captureScreen()
	}

	tmpFile := filepath.Join(os.TempDir(), "fixr-capture.jpg")
	defer os.Remove(tmpFile)

	// -x: no sound, -t jpg: JPEG format (much smaller than PNG), -l: specific window
	cmd := exec.Command("screencapture", "-x", "-t", "jpg", "-l", strconv.Itoa(windowID), tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("capture: screencapture failed: %w", err)
	}

	return readAndResize(tmpFile)
}

// captureScreen captures the entire main screen as a fallback.
func captureScreen() ([]byte, error) {
	tmpFile := filepath.Join(os.TempDir(), "fixr-capture.jpg")
	defer os.Remove(tmpFile)

	// -x: no sound, -t jpg: JPEG format, -m: main monitor only
	cmd := exec.Command("screencapture", "-x", "-t", "jpg", "-m", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("capture: screencapture failed: %w", err)
	}

	return readAndResize(tmpFile)
}

// readAndResize reads the image file and resizes it if it exceeds maxImageSize.
// Uses macOS native `sips` command — no Go image libraries needed.
func readAndResize(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("capture: failed to read screenshot: %w", err)
	}

	// If already under the limit, return as-is
	if len(data) <= maxImageSize {
		return data, nil
	}

	// Resize down using sips (native macOS tool) — halve dimensions until under limit
	for len(data) > maxImageSize {
		cmd := exec.Command("sips", "--resampleWidth", "1920", path)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			// If sips fails, return what we have — the provider will handle the error
			return data, nil
		}

		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("capture: failed to read resized screenshot: %w", err)
		}

		// If still too large after resizing to 1920px, go smaller
		if len(data) > maxImageSize {
			cmd = exec.Command("sips", "--resampleWidth", "1280", path)
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				return data, nil
			}
			data, err = os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("capture: failed to read resized screenshot: %w", err)
			}
		}
		break
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
	script := `tell application "System Events" to return name of first application process whose frontmost is true`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("capture: could not determine frontmost app: %w", err)
	}

	appName := strings.TrimSpace(string(out))
	_ = appName

	return 0, fmt.Errorf("capture: could not determine window ID for %q", appName)
}
