// Package ffmpeg provides FFmpeg detection, PATH management, and
// auto-installation for cliamp-win.
package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Papa-midnight-dev/cliamp-win/internal/appdir"
)

// BinDir returns the app-local binary directory (%APPDATA%\cliamp-win\bin).
func BinDir() (string, error) {
	dir, err := appdir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "bin"), nil
}

// Available reports whether ffmpeg is on PATH or in the app bin directory.
func Available() bool {
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return true
	}
	// Check app-local bin directory.
	binDir, err := BinDir()
	if err != nil {
		return false
	}
	info, err := os.Stat(filepath.Join(binDir, "ffmpeg.exe"))
	return err == nil && !info.IsDir()
}

// EnsurePath checks the app-local bin directory for ffmpeg and adds it
// to the process PATH if found. Returns true if ffmpeg is available
// after this call.
func EnsurePath() bool {
	// Already on PATH.
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return true
	}
	// Check app-local bin directory.
	binDir, err := BinDir()
	if err != nil {
		return false
	}
	ffmpegPath := filepath.Join(binDir, "ffmpeg.exe")
	if info, err := os.Stat(ffmpegPath); err != nil || info.IsDir() {
		return false
	}
	// Add bin dir to process PATH.
	path := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+path)
	return true
}

// Install attempts to install FFmpeg using winget.
// Output is written to stderr so the user can follow progress.
func Install() error {
	if _, err := exec.LookPath("winget"); err == nil {
		cmd := exec.Command("winget", "install", "Gyan.FFmpeg",
			"--accept-package-agreements", "--accept-source-agreements")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("winget install failed: %w", err)
		}
		return nil
	}
	return fmt.Errorf("winget not found — install FFmpeg manually from https://www.gyan.dev/ffmpeg/builds/")
}

// InstallHint returns a user-friendly install command.
func InstallHint() string {
	return "winget install Gyan.FFmpeg"
}

// StatusMessage returns a human-readable status of FFmpeg availability.
func StatusMessage() string {
	path, err := exec.LookPath("ffmpeg")
	if err == nil {
		return fmt.Sprintf("ffmpeg found: %s", path)
	}
	binDir, err := BinDir()
	if err == nil {
		local := filepath.Join(binDir, "ffmpeg.exe")
		if info, err := os.Stat(local); err == nil && !info.IsDir() {
			return fmt.Sprintf("ffmpeg found: %s (app-local)", local)
		}
	}
	return "ffmpeg not found — some audio formats (AAC, OPUS, WMA) will not play"
}

// CheckResult describes the outcome of a startup FFmpeg check.
type CheckResult int

const (
	// Found means ffmpeg is on PATH.
	Found CheckResult = iota
	// FoundLocal means ffmpeg was found in app bin dir and added to PATH.
	FoundLocal
	// NotFound means ffmpeg is not available.
	NotFound
)

// Check performs the startup FFmpeg detection.
func Check() CheckResult {
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return Found
	}
	binDir, err := BinDir()
	if err != nil {
		return NotFound
	}
	ffmpegPath := filepath.Join(binDir, "ffmpeg.exe")
	if info, err := os.Stat(ffmpegPath); err == nil && !info.IsDir() {
		// Add to process PATH.
		path := os.Getenv("PATH")
		os.Setenv("PATH", binDir+string(os.PathListSeparator)+path)
		return FoundLocal
	}
	return NotFound
}

// UpdateErrorMessage replaces the generic "install with your package manager"
// error with a Windows-specific message including the app bin dir path.
func UpdateErrorMessage(ext string) string {
	binDir, _ := BinDir()
	msg := fmt.Sprintf("ffmpeg is required to play %s files", ext)
	if binDir != "" {
		msg += fmt.Sprintf(" — install: %s\n  or place ffmpeg.exe in %s", InstallHint(), binDir)
	} else {
		msg += " — install: " + InstallHint()
	}
	return msg
}

// NeedsFFmpeg returns true if the file extension requires FFmpeg for decoding.
func NeedsFFmpeg(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".m4a", ".aac", ".m4b", ".alac", ".wma", ".opus", ".webm":
		return true
	}
	return false
}
