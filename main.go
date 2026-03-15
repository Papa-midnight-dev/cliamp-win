// Package main is the entry point for the cliamp-win terminal music player.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Papa-midnight-dev/cliamp-win/config"
	"github.com/Papa-midnight-dev/cliamp-win/external/local"
	"github.com/Papa-midnight-dev/cliamp-win/external/navidrome"
	"github.com/Papa-midnight-dev/cliamp-win/external/radio"
	"github.com/Papa-midnight-dev/cliamp-win/external/ytmusic"
	"github.com/Papa-midnight-dev/cliamp-win/internal/ffmpeg"
	"github.com/Papa-midnight-dev/cliamp-win/internal/tray"
	"github.com/Papa-midnight-dev/cliamp-win/internal/winterm"
	"github.com/Papa-midnight-dev/cliamp-win/mpris"
	"github.com/Papa-midnight-dev/cliamp-win/player"
	"github.com/Papa-midnight-dev/cliamp-win/playlist"
	"github.com/Papa-midnight-dev/cliamp-win/resolve"
	"github.com/Papa-midnight-dev/cliamp-win/theme"
	"github.com/Papa-midnight-dev/cliamp-win/ui"
	"github.com/Papa-midnight-dev/cliamp-win/upgrade"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z".
var version string

func run(overrides config.Overrides, positional []string) error {
	// Check FFmpeg availability early; add app-local bin to PATH if needed.
	switch ffmpeg.Check() {
	case ffmpeg.FoundLocal:
		fmt.Fprintf(os.Stderr, "Using app-local ffmpeg.\n")
	case ffmpeg.NotFound:
		fmt.Fprintf(os.Stderr, "Note: ffmpeg not found — some formats (AAC, OPUS, WMA) won't play.\n")
		fmt.Fprintf(os.Stderr, "Install: %s\n\n", ffmpeg.InstallHint())
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	overrides.Apply(&cfg)

	// Build provider list: Radio is always available, Navidrome if configured.
	radioProv := radio.New()
	var providers []ui.ProviderEntry
	providers = append(providers, ui.ProviderEntry{Key: "radio", Name: "Radio", Provider: radioProv})

	var navClient *navidrome.NavidromeClient
	if c := navidrome.NewFromConfig(cfg.Navidrome); c != nil {
		navClient = c
	} else if c := navidrome.NewFromEnv(); c != nil {
		navClient = c
	}
	if navClient != nil {
		providers = append(providers, ui.ProviderEntry{Key: "navidrome", Name: "Navidrome", Provider: navClient})
	}

	var ytProviders ytmusic.Providers
	// Enable YouTube providers if any [yt]/[youtube]/[ytmusic] config exists,
	// or if the --provider flag selects a YouTube provider,
	// or if fallback credentials are available.
	ytWanted := cfg.YouTubeMusic.IsSetOrFallback(ytmusic.FallbackCredentials)
	if !ytWanted {
		// Also enable if --provider flag selects a YouTube provider.
		switch cfg.Provider {
		case "yt", "youtube", "ytmusic":
			ytWanted = true
		}
	}
	if ytWanted {
		ytClientID, ytClientSecret := cfg.YouTubeMusic.ResolveCredentials(ytmusic.FallbackCredentials)
		// Configure yt-dlp cookie source for YouTube Music uploads/private tracks.
		if cfg.YouTubeMusic.CookiesFrom != "" {
			player.SetYTDLCookiesFrom(cfg.YouTubeMusic.CookiesFrom)
		}
		if ytClientID == "" || ytClientSecret == "" {
			fmt.Fprintf(os.Stderr, "YouTube: no credentials available (configure client_id/client_secret in config.toml)\n")
		} else {
			// YouTube playback requires yt-dlp. Check early and offer to install.
			if !player.YTDLPAvailable() {
				fmt.Fprintf(os.Stderr, "\nYouTube requires yt-dlp for audio playback.\n")
				fmt.Fprintf(os.Stderr, "Install command: %s\n\n", player.YtdlpInstallHint())
				fmt.Fprintf(os.Stderr, "Press Enter to install automatically, or Ctrl+C to skip... ")
				fmt.Scanln()
				fmt.Fprintf(os.Stderr, "Installing yt-dlp...\n")
				if err := player.InstallYTDLP(); err != nil {
					fmt.Fprintf(os.Stderr, "Installation failed: %v\n", err)
					fmt.Fprintf(os.Stderr, "YouTube providers disabled. Install manually and restart.\n\n")
				} else {
					fmt.Fprintf(os.Stderr, "yt-dlp installed successfully!\n\n")
				}
			}
			if player.YTDLPAvailable() {
				ytProviders = ytmusic.New(nil, ytClientID, ytClientSecret, cfg.YouTubeMusic.CookiesFrom != "")
				providers = append(providers,
					ui.ProviderEntry{Key: "yt", Name: "YouTube (All)", Provider: ytProviders.All},
					ui.ProviderEntry{Key: "youtube", Name: "YouTube", Provider: ytProviders.Video},
					ui.ProviderEntry{Key: "ytmusic", Name: "YouTube Music", Provider: ytProviders.Music},
				)
			}
		}
	}

	localProv := local.New()

	defer resolve.CleanupYTDL()
	if ytProviders.Music != nil {
		defer ytProviders.Music.Close()
	}

	if len(positional) > 0 && (positional[0] == "search" || positional[0] == "search-sc") {
		if len(positional) == 1 {
			return fmt.Errorf("search requires a query string (e.g. cliamp-win search \"never gonna give you up\")")
		}
		prefix := "ytsearch1:"
		if positional[0] == "search-sc" {
			prefix = "scsearch1:"
		}
		query := strings.Join(positional[1:], " ")
		positional = []string{prefix + query}
	}

	resolved, err := resolve.Args(positional)
	if err != nil {
		return err
	}

	// Determine default provider key.
	defaultProvider := cfg.Provider
	if defaultProvider == "" {
		defaultProvider = "radio"
	}

	// No args + radio provider: stream the built-in radio directly.
	if len(positional) == 0 && defaultProvider == "radio" {
		resolved.Pending = append(resolved.Pending, "https://radio.cliamp.stream/streams.m3u")
	}

	pl := playlist.New()
	pl.Add(resolved.Tracks...)

	// Resolve sample rate: 0 means auto-detect from the system's default
	// output audio device (e.g. 48 kHz for USB-C headphones). Falls back
	// to 44100 Hz if detection is unavailable or returns an unusable value.
	sampleRate := cfg.SampleRate
	if sampleRate == 0 {
		if detected := player.DeviceSampleRate(); detected > 0 {
			sampleRate = detected
		} else {
			sampleRate = 44100
		}
	}

	p, err := player.New(player.Quality{
		SampleRate:      sampleRate,
		BufferMs:        cfg.BufferMs,
		ResampleQuality: cfg.ResampleQuality,
		BitDepth:        cfg.BitDepth,
	})
	if err != nil {
		return fmt.Errorf("player: %w", err)
	}
	defer p.Close()

	cfg.ApplyPlayer(p)
	cfg.ApplyPlaylist(pl)

	themes := theme.LoadAll()

	m := ui.NewModel(p, pl, providers, defaultProvider, localProv, themes, cfg.Navidrome, navClient)
	m.SetSeekStepLarge(cfg.SeekStepLargeDuration())
	m.SetPendingURLs(resolved.Pending)
	if len(resolved.Tracks) == 0 && len(resolved.Pending) == 0 {
		m.StartInProvider()
	}
	if cfg.EQPreset != "" && cfg.EQPreset != "Custom" {
		m.SetEQPreset(cfg.EQPreset)
	}
	if cfg.Theme != "" {
		m.SetTheme(cfg.Theme)
	}
	if cfg.Visualizer != "" {
		m.SetVisualizer(cfg.Visualizer)
	}
	if overrides.Play != nil && *overrides.Play {
		m.SetAutoPlay(true)
	}

	prog := tea.NewProgram(m, tea.WithAltScreen())

	if svc, err := mpris.New(func(msg interface{}) { prog.Send(msg) }); err == nil && svc != nil {
		defer svc.Close()
		go prog.Send(mpris.InitMsg{Svc: svc})
	}

	// Start system tray in background.
	go tray.Run(tray.Callbacks{
		OnPlayPause: func() { prog.Send(mpris.PlayPauseMsg{}) },
		OnNext:      func() { prog.Send(mpris.NextMsg{}) },
		OnPrev:      func() { prog.Send(mpris.PrevMsg{}) },
		OnQuit:      func() { prog.Send(mpris.QuitMsg{}) },
	})
	defer tray.Quit()

	finalModel, err := prog.Run()
	if err != nil {
		return err
	}

	// Persist theme selection across restarts.
	if fm, ok := finalModel.(ui.Model); ok {
		themeName := fm.ThemeName()
		if themeName == theme.DefaultName {
			themeName = ""
		}
		_ = config.Save("theme", fmt.Sprintf("%q", themeName))
	}

	return nil
}

const helpText = `cliamp-win — Windows-first retro terminal music player

Usage: cliamp-win [flags] <file|folder|url> [...]

Playback:
  --volume <dB>           Volume in dB, range [-30, +6] (e.g. --volume -5)
  --shuffle
  --repeat <off|all|one>
  --mono / --no-mono
  --auto-play             Start playback immediately

Audio engine:
  --sample-rate <Hz>      Output sample rate (0=auto, 22050, 44100, 48000, 96000, 192000)
  --buffer-ms <ms>        Speaker buffer in milliseconds (50–500)
  --resample-quality <n>  Resample quality factor (1–4)
  --bit-depth <n>         PCM bit depth: 16 (default) or 32 (lossless)

Provider:
  --provider <name>       Default provider: radio, navidrome, yt, youtube, ytmusic (default: radio)

Appearance:
  --theme <name>          UI theme name
  --visualizer <mode>     Visualizer mode (Bars, Bricks, Columns, Wave, Scatter, Flame, Retro, Pulse, Matrix, Binary, None)
  --eq-preset <name>      EQ preset name (e.g. "Bass Boost")

General:
  -h, --help              Show this help message
  -v, --version           Show the current version
  --upgrade               Upgrade to the latest release

Examples:
  cliamp-win track.mp3 song.flac C:\Users\You\Music
  cliamp-win --shuffle --volume -5 track.mp3
  cliamp-win track.mp3 --repeat all --mono
  cliamp-win --auto-play --shuffle C:\Users\You\Music
  cliamp-win --eq-preset "Bass Boost" C:\Users\You\Music
  cliamp-win https://example.com/song.mp3
  cliamp-win http://radio.example.com/stream.m3u
  cliamp-win search "rick astley"            # search YouTube
  cliamp-win search-sc "lofi beats"          # search SoundCloud
  cliamp-win https://soundcloud.com/user/sets/playlist
  cliamp-win https://www.youtube.com/watch?v=...

Environment:
  NAVIDROME_URL, NAVIDROME_USER, NAVIDROME_PASS   Navidrome server (env fallback)

Config:    %APPDATA%\cliamp-win\config.toml  (see config.toml.example)
Radios:    %APPDATA%\cliamp-win\radios.toml
Playlists: %APPDATA%\cliamp-win\playlists\*.toml
Formats:   mp3, wav, flac, ogg, m4a, aac, opus, wma (aac/opus/wma need ffmpeg)
SoundCloud/YouTube/Bandcamp require yt-dlp`

func main() {
	// Enable VT processing for ANSI colors in cmd.exe and PowerShell 5.
	winterm.EnableVT()

	action, overrides, positional, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch action {
	case "help":
		fmt.Println(helpText)
		return
	case "version":
		if version == "" {
			fmt.Println("cliamp-win (dev build)")
		} else {
			fmt.Printf("cliamp-win %s\n", version)
		}
		return
	case "upgrade":
		if err := upgrade.Run(version); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := run(overrides, positional); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
