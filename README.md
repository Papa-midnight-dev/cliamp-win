```
 ██████ ██      ██  █████  ███    ███ ██████        ██     ██ ██ ███    ██
██      ██      ██ ██   ██ ████  ████ ██   ██       ██     ██ ██ ████   ██
██      ██      ██ ███████ ██ ████ ██ ██████  █████ ██  █  ██ ██ ██ ██  ██
██      ██      ██ ██   ██ ██  ██  ██ ██             ██ ███ ██ ██ ██  ██ ██
 ██████ ███████ ██ ██   ██ ██      ██ ██              ███ ███  ██ ██   ████
```

A Windows-first fork of [cliamp](https://github.com/bjarneo/cliamp) — a retro terminal music player inspired by Winamp. Play local files, streams, YouTube, SoundCloud, and Navidrome with a spectrum visualizer, parametric EQ, and playlist management.

## What's Different

This is a **deep fork** focused on Windows:

- Native Windows paths (`%APPDATA%\cliamp-win\`)
- Removed Linux-only code (MPRIS/D-Bus, PulseAudio, CoreAudio)
- Removed Spotify (requires CGO/Linux libs)
- No upstream telemetry
- Bundled FFmpeg support (planned)
- System tray, media controls, toast notifications (planned)

## Install

### Prerequisites

- **Windows 10/11** with [Windows Terminal](https://aka.ms/terminal)
- **yt-dlp** for YouTube/SoundCloud: `winget install yt-dlp`
- **ffmpeg** for AAC/OPUS/WMA: `winget install Gyan.FFmpeg`

### From Source

```powershell
# Requires Go 1.22+
go install github.com/Papa-midnight-dev/cliamp-win@latest

# Or build locally:
git clone https://github.com/Papa-midnight-dev/cliamp-win
cd cliamp-win
go build -o cliamp-win.exe .
```

## Usage

```powershell
# Play local files
cliamp-win track.mp3 C:\Users\You\Music

# Play with options
cliamp-win --shuffle --auto-play C:\Users\You\Music

# Stream radio (default)
cliamp-win

# YouTube search
cliamp-win search "never gonna give you up"

# SoundCloud search
cliamp-win search-sc "lofi beats"

# HTTP stream
cliamp-win https://example.com/stream.m3u
```

## Supported Formats

| Format | Status |
|--------|--------|
| MP3, WAV, FLAC, OGG | Built-in |
| AAC, OPUS, WMA, WebM | Requires ffmpeg |
| YouTube, SoundCloud, Bandcamp | Requires yt-dlp |

## Config

Configuration is stored in `%APPDATA%\cliamp-win\config.toml`.

See `config.toml.example` for all options.

## Credits

Forked from [bjarneo/cliamp](https://github.com/bjarneo/cliamp) — all credit for the original player goes to the upstream maintainers.

## License

MIT — see [LICENSE](LICENSE).
