# QuickRec — Claude context

Personal Windows tool: Chrome extension + Go tray app to launch yt-dlp for Twitch streams.

## Architecture

- `tray-app/main.go` — single-file Go app. System tray via `getlantern/systray`. HTTP server on `127.0.0.1:9999`. Startup via Windows Registry `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`. Tray icon generated programmatically as PNG-in-ICO (no asset files).
- `extension/content.js` — injected into `twitch.tv/*`. Finds player controls bar via multiple CSS selectors (Twitch changes DOM often). MutationObserver handles SPA navigation between channels.
- No token auth — server is localhost-only and personal use only.

## Build

```powershell
cd tray-app
go mod tidy
go build -ldflags="-H windowsgui" -o QuickRec.exe .
```

## Key constraints

- Server must bind `127.0.0.1` only, never `0.0.0.0`.
- URL validation: reject anything not containing `twitch.tv/`. Strip query params before passing to yt-dlp.
- exec.Command args passed separately (not shell-interpolated) — no injection risk.
- `-H windowsgui` linker flag is required — omitting it spawns a visible console on startup.
- Twitch CSS selectors in `CONTROL_SELECTORS` break periodically — check `[data-a-target]` attributes first as they are more stable than class names.

## Dependencies

- `github.com/getlantern/systray` — tray
- `golang.org/x/sys/windows/registry` — startup registry
- yt-dlp must be in PATH on the host machine
