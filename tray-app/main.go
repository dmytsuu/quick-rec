package main

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/registry"
)

const (
	appName = "QuickRec"
	port    = "9999"
	regRun  = `Software\Microsoft\Windows\CurrentVersion\Run`
)

var (
	srv       *http.Server
	listening bool
	outputDir string
)

func initOutputDir() {
	home, _ := os.UserHomeDir()
	outputDir = filepath.Join(home, "Videos", "TwitchVODs")
	os.MkdirAll(outputDir, 0755)
}

func isStartupEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, regRun, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	val, _, err := k.GetStringValue(appName)
	return err == nil && val != ""
}

func addStartup() {
	exe, _ := os.Executable()
	k, err := registry.OpenKey(registry.CURRENT_USER, regRun, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()
	k.SetStringValue(appName, exe)
}

func removeStartup() {
	k, err := registry.OpenKey(registry.CURRENT_USER, regRun, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()
	k.DeleteValue(appName)
}

func startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/run", handleRun)
	srv = &http.Server{Addr: "127.0.0.1:" + port, Handler: mux}
	listening = true
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		listening = false
	}
}

func stopServer() {
	if srv != nil {
		srv.Shutdown(context.Background())
		srv = nil
	}
	listening = false
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	rawURL := r.URL.Query().Get("url")
	if !strings.Contains(rawURL, "twitch.tv/") {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}

	cleanURL := strings.SplitN(rawURL, "?", 2)[0]
	exec.Command("cmd", "/c", "start", "", "cmd", "/k", "yt-dlp", "-P", outputDir, cleanURL).Start()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// makeTrayIcon generates a Twitch Glitch pixel-art icon as PNG-in-ICO.
func makeTrayIcon() []byte {
	const (
		T = iota // transparent
		P        // purple #9146FF
		W        // white
	)
	pixels := [16][16]uint8{
		{T, T, P, P, P, P, P, P, P, P, P, P, T, T, T, T},
		{T, P, P, P, P, P, P, P, P, P, P, P, P, T, T, T},
		{P, P, P, P, P, P, P, P, P, P, P, P, P, P, T, T},
		{P, P, P, P, P, P, P, P, P, P, P, P, P, P, T, T},
		{P, P, W, W, P, P, P, P, W, W, P, P, P, P, T, T},
		{P, P, W, W, P, P, P, P, W, W, P, P, P, P, T, T},
		{P, P, W, W, P, P, P, P, W, W, P, P, P, P, T, T},
		{P, P, W, W, P, P, P, P, W, W, P, P, P, P, T, T},
		{P, P, P, P, P, P, P, P, P, P, P, P, P, P, T, T},
		{P, P, P, P, P, P, P, P, P, P, P, P, P, P, T, T},
		{T, P, P, P, P, P, P, P, P, P, P, P, P, T, T, T},
		{T, T, P, P, P, P, P, P, P, P, P, P, T, T, T, T},
		{T, T, T, P, P, T, T, T, T, P, P, T, T, T, T, T},
		{T, T, T, P, P, T, T, T, T, P, P, T, T, T, T, T},
		{T, T, T, T, T, T, T, T, T, T, T, T, T, T, T, T},
		{T, T, T, T, T, T, T, T, T, T, T, T, T, T, T, T},
	}

	purple := color.NRGBA{R: 145, G: 70, B: 255, A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			switch pixels[y][x] {
			case P:
				img.Set(x, y, purple)
			case W:
				img.Set(x, y, white)
			}
		}
	}

	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	pngData := pngBuf.Bytes()

	var ico bytes.Buffer
	pngSize := uint32(len(pngData))
	ico.Write([]byte{0, 0, 1, 0, 1, 0})
	ico.Write([]byte{16, 16, 0, 0, 1, 0, 32, 0})
	ico.Write([]byte{byte(pngSize), byte(pngSize >> 8), byte(pngSize >> 16), byte(pngSize >> 24)})
	ico.Write([]byte{22, 0, 0, 0})
	ico.Write(pngData)
	return ico.Bytes()
}

var trayIcon = makeTrayIcon()

func main() {
	initOutputDir()
	systray.Run(onReady, func() { stopServer() })
}

func onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTooltip(appName + " — listening on :" + port)

	mStatus := systray.AddMenuItem("● Listening on :"+port, "")
	mStatus.Disable()
	systray.AddSeparator()
	mToggle := systray.AddMenuItem("Stop Listening", "")
	mStartup := systray.AddMenuItemCheckbox("Start at Startup", "", isStartupEnabled())
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "")

	go startServer()

	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				if listening {
					stopServer()
					mStatus.SetTitle("○ Stopped")
					mToggle.SetTitle("Start Listening")
					systray.SetIcon(trayIcon)
					systray.SetTooltip(appName + " — stopped")
				} else {
					go startServer()
					mStatus.SetTitle("● Listening on :" + port)
					mToggle.SetTitle("Stop Listening")
					systray.SetIcon(trayIcon)
					systray.SetTooltip(appName + " — listening on :" + port)
				}
			case <-mStartup.ClickedCh:
				if mStartup.Checked() {
					mStartup.Uncheck()
					removeStartup()
				} else {
					mStartup.Check()
					addStartup()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}
