package appdir

import (
	"os"
	"path/filepath"
)

// Dir returns the cliamp-win configuration directory.
// Uses os.UserConfigDir() which returns %APPDATA% on Windows,
// $XDG_CONFIG_HOME or ~/.config on Linux/macOS.
func Dir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cliamp-win"), nil
}
