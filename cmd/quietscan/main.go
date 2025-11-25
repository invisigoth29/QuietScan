package main
import (
    "quietscan/storage"
    "quietscan/ui"

    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2"
)

func main() {
    quietApp := app.New()
    quietApp.SetIcon(resourceIconPng) // set icon from bundled resources

    // load latest results (if file exists)
    latest := storage.LoadLatestResults()

    window := quietApp.NewWindow("QuietScan")
    window.Resize(fyne.NewSize(900, 600))

    ui.RenderDashboard(window, latest)

    window.ShowAndRun()
}