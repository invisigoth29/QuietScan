package ui

import (
    "fmt"
    "path/filepath"
    "time"

    "quietscan/scanner"
    "quietscan/storage"
    "quietscan/types"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func RenderDashboard(win fyne.Window, latest *types.ScanResult) {
    adapter, subnet := scanner.GetActiveAdapter()

    statusLabel := widget.NewLabel("Idle")
    lastScanLabel := widget.NewLabel("Last Scan: None")

    if latest != nil {
        lastScanLabel.SetText("Last Scan: " + latest.Timestamp.Format(time.RFC1123))
    }

    table := NewResultsTable(latest)

    // Create progress bar
    progressBar := widget.NewProgressBar()
    progressBar.Hide() // Hide initially
    progressLabel := widget.NewLabel("")
    progressLabel.Hide()

    scanBtn := widget.NewButton("Scan Now", func() {
        statusLabel.SetText("Scanning… Polite Mode Enabled")
        progressBar.Show()
        progressLabel.Show()
        progressBar.SetValue(0)

        go func() {
            res := scanner.RunPoliteScanWithProgress(func(current, total int) {
                // Update progress (Fyne widgets are thread-safe)
                progress := float64(current) / float64(total)
                progressBar.SetValue(progress)
                remaining := total - current
                progressLabel.SetText(fmt.Sprintf("%d hosts remaining", remaining))
            })
            storage.SaveScanResult(res)

            // Final update
            progressBar.SetValue(1.0)
            progressLabel.SetText("0 hosts remaining")

            // Brief delay to show completion, then hide progress bar
            time.Sleep(500 * time.Millisecond)
            progressBar.Hide()
            progressLabel.Hide()

            lastScanLabel.SetText("Last Scan: " + res.Timestamp.Format(time.RFC1123))
            table.UpdateData(res)
            statusLabel.SetText("Scan complete.")
        }()
    })

    historyBtn := widget.NewButton("Show History", func() {
        RenderHistoryModal(win, table)
    })

    toolsBtn := widget.NewButton("Tools ▾", nil)
    toolsMenu := fyne.NewMenu("Tools",
        fyne.NewMenuItem("Update MAC Vendor Database", func() {
            RenderUpdateModal(win)
        }),
        fyne.NewMenuItemSeparator(),
        fyne.NewMenuItem("Export to CSV", func() {
            currentResult := storage.LoadLatestResults()
            if currentResult == nil {
                showMessageModal(win, "No Data", "No scan results available to export.\nPlease run a scan first.")
                return
            }
            
            csvPath, err := storage.ExportToCSVWithMetadata(currentResult)
            if err != nil {
                showMessageModal(win, "Export Failed", fmt.Sprintf("Failed to export CSV:\n%v", err))
                return
            }
            
            filename := filepath.Base(csvPath)
            showMessageModal(win, "Export Successful", fmt.Sprintf("Scan results exported to:\n%s", filename))
            statusLabel.SetText(fmt.Sprintf("Exported to %s", filename))
        }),
        fyne.NewMenuItemSeparator(),
        fyne.NewMenuItem("Reset Scan History", func() {
            storage.SaveScanResult(nil)
            table.UpdateData(nil)
            lastScanLabel.SetText("Last Scan: None")
            statusLabel.SetText("History reset.")
        }),
    )

    toolsBtn.OnTapped = func() {
        pop := widget.NewPopUpMenu(toolsMenu, win.Canvas())
        pop.ShowAtPosition(toolsBtn.Position().AddXY(0, toolsBtn.Size().Height))
    }

    header := container.NewVBox(
        widget.NewLabel("Adapter: "+adapter),
        widget.NewLabel("Subnet: "+subnet),
        lastScanLabel,
        container.NewHBox(scanBtn, historyBtn, toolsBtn),
    )

    // Bottom section with status, progress bar, and progress label
    bottomSection := container.NewVBox(
        statusLabel,
        progressLabel,
        progressBar,
    )

    content := container.NewBorder(header, bottomSection, nil, nil, table.Table)
    win.SetContent(content)
}

// showMessageModal displays a simple message modal dialog
func showMessageModal(win fyne.Window, title, message string) {
    closeBtn := widget.NewButton("OK", nil)
    
    content := container.NewVBox(
        widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
        widget.NewLabel(message),
        closeBtn,
    )
    
    modal := widget.NewModalPopUp(content, win.Canvas())
    
    closeBtn.OnTapped = func() {
        modal.Hide()
    }
    
    modal.Show()
}