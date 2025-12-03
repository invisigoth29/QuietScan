package main

import (
	"quietscan/assets"
	"quietscan/storage"
	"quietscan/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	quietApp := app.New()
	quietApp.SetIcon(assets.ResourceIconPng) // set icon from bundled resources

	// load latest results (if file exists)
	latest := storage.LoadLatestResults()

	window := quietApp.NewWindow("QuietScan")
	window.Resize(fyne.NewSize(900, 600))

	ui.RenderDashboard(window, latest)

	window.ShowAndRun()
}
