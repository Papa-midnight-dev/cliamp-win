// Package mpris provides stub types for MPRIS media control integration.
// On Windows, D-Bus is not available so all methods are no-ops.
// Future milestones will replace this with Windows SMTC integration.
package mpris

import "math"

// Message types used by the UI event loop.
type (
	PlayPauseMsg   struct{}
	NextMsg        struct{}
	PrevMsg        struct{}
	StopMsg        struct{}
	QuitMsg        struct{}
	SeekMsg        struct{ Offset int64 }   // microseconds (relative)
	SetPositionMsg struct{ Position int64 } // microseconds (absolute)
	SetVolumeMsg   struct{ Volume float64 } // linear 0.0–1.0
	InitMsg        struct{ Svc *Service }
)

// TrackInfo carries metadata for the currently playing track.
type TrackInfo struct {
	Title       string
	Artist      string
	Album       string
	Genre       string
	TrackNumber int
	URL         string
	Length      int64 // microseconds
}

// Service is a no-op stub (SMTC integration planned for M3).
type Service struct{}

// New returns nil — media transport controls are not yet implemented.
func New(send func(interface{})) (*Service, error) {
	return nil, nil
}

// Update is a no-op.
func (s *Service) Update(status string, track TrackInfo, volumeDB float64, positionUs int64, canSeek bool) {
}

// LinearToDb converts a 0.0–1.0 linear volume to dB (range [-30, +6]).
func LinearToDb(v float64) float64 {
	if v <= 0 {
		return -30
	}
	if v >= 1 {
		return 6
	}
	db := 20*math.Log10(v) + 6
	if db < -30 {
		return -30
	}
	return db
}

// EmitSeeked is a no-op.
func (s *Service) EmitSeeked(positionUs int64) {}

// Close is a no-op.
func (s *Service) Close() {}
