package main

import (
	"context"
	"embed"

	"github.com/devlup-labs/Libr/core/mod_client/peers"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:       "libr",
		Width:       1024,
		Height:      768,
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup:   app.startup,
		OnShutdown: func(ctx context.Context) {
			println("[DEBUG] Shutting down, cleaning up peer...")
			if peers.Cp != nil {
				if err := peers.Cp.Close(); err != nil {
					println("Error closing peer:", err.Error())
				}
			}
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Bind:             []interface{}{app},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
