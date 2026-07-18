package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"

	"decoreba/internal/core/clipboard"
	"decoreba/internal/core/store"

	"decoreba/desktop/internal/platform/hotkey"
	"decoreba/desktop/internal/platform/tray"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	hotkeyKey := flag.String("hotkey", "space", "hotkey key name (space, d, tab, slash, etc.)")
	flag.Parse()

	showCh := make(chan bool, 8)
	quitCh := make(chan struct{})

	app := NewApp()

	trayAvailable := tray.Available()
	startHidden := trayAvailable

	if clipboard.IsWSL() {
		if !trayAvailable {
			log.Printf("main: WSL detected, no tray available — window will start visible.")
			log.Printf("main: hotkey Alt+Shift+%s may conflict with Windows host shortcuts.", *hotkeyKey)
			startHidden = false
		}
	} else if !trayAvailable {
		log.Printf("main: no tray available — window will start visible")
		startHidden = false
	}

	err := wails.Run(&options.App{
		Title:             "decoreba",
		Width:             app.settings.Width,
		Height:            app.settings.Height,
		MinWidth:          400,
		MinHeight:         280,
		MaxWidth:          1200,
		MaxHeight:         800,
		Frameless:         true,
		AlwaysOnTop:       app.settings.AlwaysOnTop,
		StartHidden:       startHidden,
		HideWindowOnClose: true,

		AssetServer: &assetserver.Options{
			Assets: assets,
		},

		Bind: []any{
			app,
		},

		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Linux: &linux.Options{},

		OnStartup: func(ctx context.Context) {
			app.SetContext(ctx)

			runtime.WindowExecJS(ctx,
				fmt.Sprintf("document.documentElement.style.setProperty('--font-scale','%g')", app.settings.FontScale))

			if trayAvailable {
				t, err := tray.New(showCh, quitCh)
				if err != nil {
					log.Printf("tray: %v", err)
				} else {
					log.Printf("tray: started")
					_ = t
				}
			}

			h, err := hotkey.NewKey(showCh, *hotkeyKey)
			if err != nil {
				log.Printf("hotkey: %v", err)
			}

			go func() {
				for range showCh {
					runtime.WindowShow(ctx)
					runtime.WindowCenter(ctx)
					runtime.WindowExecJS(ctx, "window.decorebaShow && window.decorebaShow()")
				}
			}()

			go func() {
				<-quitCh
				if h != nil {
					h.Close()
				}
				os.Exit(0)
			}()
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	if store.Save(app.store) != nil {
		log.Fatal("failed to save store")
	}
}
