// Package winterm ensures Windows console VT processing is enabled
// for proper ANSI color and cursor control support.
package winterm

import (
	"os"

	"golang.org/x/sys/windows"
)

// EnableVT enables virtual terminal processing on stdout and stdin.
// This is required for ANSI escape sequences (colors, cursor movement)
// on older Windows 10 builds and cmd.exe. Windows Terminal enables
// this by default, but cmd.exe and PowerShell 5 may not.
func EnableVT() {
	enableVTOnHandle(os.Stdout)
	enableVTOnHandle(os.Stderr)
	enableVTInputOnHandle(os.Stdin)
}

func enableVTOnHandle(f *os.File) {
	h := windows.Handle(f.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.ENABLE_PROCESSED_OUTPUT
	_ = windows.SetConsoleMode(h, mode)
}

func enableVTInputOnHandle(f *os.File) {
	h := windows.Handle(f.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	_ = windows.SetConsoleMode(h, mode)
}
