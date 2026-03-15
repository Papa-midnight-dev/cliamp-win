// Package notify provides Windows toast notifications for track changes.
package notify

import (
	"sync"

	"gopkg.in/toast.v1"
)

// Notifier sends Windows toast notifications for track changes.
type Notifier struct {
	enabled bool
	mu      sync.Mutex
	last    string // last notified track to avoid duplicates
}

// New creates a Notifier. If enabled is false, all methods are no-ops.
func New(enabled bool) *Notifier {
	return &Notifier{enabled: enabled}
}

// TrackChanged sends a toast notification for the new track.
func (n *Notifier) TrackChanged(title, artist string) {
	if n == nil || !n.enabled {
		return
	}
	n.mu.Lock()
	key := title + "|" + artist
	if key == n.last {
		n.mu.Unlock()
		return
	}
	n.last = key
	n.mu.Unlock()

	body := title
	if artist != "" {
		body = artist + " — " + title
	}

	// Fire and forget — toast spawns PowerShell, don't block.
	go func() {
		t := toast.Notification{
			AppID:   "cliamp-win",
			Title:   "Now Playing",
			Message: body,
		}
		_ = t.Push()
	}()
}

// SetEnabled toggles notifications on or off.
func (n *Notifier) SetEnabled(enabled bool) {
	if n == nil {
		return
	}
	n.mu.Lock()
	n.enabled = enabled
	n.mu.Unlock()
}
