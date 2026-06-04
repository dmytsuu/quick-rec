# QuickRec

Windows system-tray app + Chrome extension that adds a **⏺ REC** button to Twitch streams. One click opens a new terminal window running `yt-dlp` for that channel.

## Requirements

- [Go 1.21+](https://go.dev/dl/)
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) — must be in PATH
- Windows 10/11
- Google Chrome

## Setup

### 1. Build the tray app

```powershell
cd tray-app
go mod tidy
go build -ldflags="-H windowsgui" -o QuickRec.exe .
```

The `-H windowsgui` flag suppresses the console window on launch.

### 2. Load the Chrome extension

1. Open `chrome://extensions`
2. Enable **Developer mode** (top right)
3. Click **Load unpacked** → select the `extension/` folder

### 3. Run

Launch `QuickRec.exe` — it appears in the system tray.  
Navigate to any Twitch channel — the **⏺ REC** button appears in the player controls bar.

## Tray menu

| Item | Action |
|---|---|
| ● Listening on :9999 | status indicator (disabled) |
| Stop / Start Listening | toggle the HTTP server |
| Start at Startup | adds/removes from `HKCU\...\Run` registry |
| Quit | exit |

## How it works

```
[Chrome extension] ──click──▶ fetch http://127.0.0.1:9999/run?url=...
                                        │
                              [Go tray app — server]
                                        │
                              cmd /c start "" cmd /k yt-dlp <url>
                                        │
                              [new terminal window]
```

The server binds exclusively to `127.0.0.1` — not reachable from outside the machine.  
Only URLs containing `twitch.tv/` are accepted.

## File structure

```
quick-rec/
├── tray-app/
│   ├── main.go       Go tray app + HTTP server
│   └── go.mod
└── extension/
    ├── manifest.json
    └── content.js    injects REC button, handles SPA navigation
```
