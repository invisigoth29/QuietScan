package ui

import (
    "quietscan/storage"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func RenderHistoryModal(win fyne.Window, table *ResultsTable) {
    history := storage.LoadHistory()

    list := widget.NewList(
        func() int { return len(history) },
        func() fyne.CanvasObject { return widget.NewButton("", nil) },
        func(i int, o fyne.CanvasObject) {
            btn := o.(*widget.Button)
            entry := history[i]

            btn.SetText(entry.Timestamp.Format("Jan 02 2006 3:04 PM"))
            btn.OnTapped = func() {
                table.UpdateData(&entry)
            }
        },
    )

    modal := widget.NewModalPopUp(
        container.NewVBox(
            widget.NewLabel("QuietScan History"),
            list,
            widget.NewButton("Close", func() {
                modal.Hide()
            }),
        ),
        win.Canvas(),
    )

    modal.Show()
}