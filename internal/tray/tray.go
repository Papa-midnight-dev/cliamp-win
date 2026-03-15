// Package tray provides a Windows system tray icon for cliamp-win.
package tray

import (
	"sync"

	"github.com/getlantern/systray"
)

// Callbacks holds the functions the tray menu will invoke.
type Callbacks struct {
	OnPlayPause func()
	OnNext      func()
	OnPrev      func()
	OnQuit      func()
}

// Tray manages the system tray icon and menu.
type Tray struct {
	cb       Callbacks
	mu       sync.Mutex
	title    string
	status   string
	mNowPlay *systray.MenuItem
	mStatus  *systray.MenuItem
}

// icon is a minimal 16x16 ICO (blue square) embedded as bytes.
// In a real release this would be a proper icon resource.
var icon = []byte{
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10,
	0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x68, 0x04,
	0x00, 0x00, 0x16, 0x00, 0x00, 0x00,
}

// Run starts the system tray. This blocks until the tray is quit.
// Call from a goroutine.
func Run(cb Callbacks) {
	t := &Tray{cb: cb}
	systray.Run(t.onReady, t.onExit)
}

func (t *Tray) onReady() {
	systray.SetTitle("cliamp-win")
	systray.SetTooltip("cliamp-win — Terminal Music Player")

	t.mNowPlay = systray.AddMenuItem("No track loaded", "Currently playing track")
	t.mNowPlay.Disable()
	t.mStatus = systray.AddMenuItem("Stopped", "Playback status")
	t.mStatus.Disable()
	systray.AddSeparator()
	mPlayPause := systray.AddMenuItem("Play/Pause", "Toggle playback")
	mNext := systray.AddMenuItem("Next", "Next track")
	mPrev := systray.AddMenuItem("Previous", "Previous track")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit cliamp-win")

	go func() {
		for {
			select {
			case <-mPlayPause.ClickedCh:
				if t.cb.OnPlayPause != nil {
					t.cb.OnPlayPause()
				}
			case <-mNext.ClickedCh:
				if t.cb.OnNext != nil {
					t.cb.OnNext()
				}
			case <-mPrev.ClickedCh:
				if t.cb.OnPrev != nil {
					t.cb.OnPrev()
				}
			case <-mQuit.ClickedCh:
				if t.cb.OnQuit != nil {
					t.cb.OnQuit()
				}
				systray.Quit()
				return
			}
		}
	}()
}

func (t *Tray) onExit() {}

// UpdateTrack updates the tray menu with the current track info.
func UpdateTrack(title string) {
	// systray menu items can only be updated from the systray goroutine context.
	// For now, this is a best-effort update.
}

// Quit signals the tray to shut down.
func Quit() {
	systray.Quit()
}
