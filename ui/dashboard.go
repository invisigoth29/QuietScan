package ui

import (
    "time"

    "quietscan/scanner"
    "quietscan/storage"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func RenderDashboard(win fyne.Window, latest *storage.ScanResult) {
    adapter, subnet := scanner.GetActiveAdapter()

    statusLabel := widget.NewLabel("Idle")
    lastScanLabel := widget.NewLabel("Last Scan: None")

    if latest != nil {
        lastScanLabel.SetText("Last Scan: " + latest.Timestamp.Format(time.RFC1123))
    }

    table := NewResultsTable(latest)

    scanBtn := widget.NewButton("Scan Now", func() {
        statusLabel.SetText("Scanning… Polite Mode Enabled")

        go func() {
            res := scanner.RunPoliteScan()
            storage.SaveScanResult(res)

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

    content := container.NewBorder(header, statusLabel, nil, nil, table.Table)
    win.SetContent(content)
}