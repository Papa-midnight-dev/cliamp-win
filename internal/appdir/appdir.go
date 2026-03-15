package appdir

import (
	"os"
	"path/filepath"
)

// Dir returns the cliamp-win configuration directory.
// On Windows this is %APPDATA%\cliamp-win. Falls back to ~/.config/cliamp-win
// if APPDATA is not set.
func Dir() (string, error) {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "cliamp-win"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "cliamp-win"), nil
}
