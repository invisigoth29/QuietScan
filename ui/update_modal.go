package ui

import (
    "fmt"
    "quietscan/vendor"
    "time"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func RenderUpdateModal(win fyne.Window) {
    status := widget.NewLabel("Ready to update vendor database.")
    progress := widget.NewLabel("")

    updateBtn := widget.NewButton("Download Latest IEEE OUI Data", nil)
    applyBtn := widget.NewButton("Apply Update", nil)
    closeBtn := widget.NewButton("Close", nil)

    applyBtn.Disable()

    content := container.NewVBox(
        widget.NewLabelWithStyle("Update MAC Vendor Database", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
        status,
        progress,
        updateBtn,
        applyBtn,
        closeBtn,
    )

    modal := widget.NewModalPopUp(content, win.Canvas())

    closeBtn.OnTapped = func() {
        modal.Hide()
    }

    updateBtn.OnTapped = func() {
        updateBtn.Disable()
        status.SetText("Downloading OUI database from IEEE…")
        progress.SetText("")

        go func() {
            time.Sleep(300 * time.Millisecond)

            local, err := vendor.LoadLocalOUI()
            if err != nil {
                status.SetText("Error loading local OUI file: " + err.Error())
                updateBtn.Enable()
                return
            }

            master, err := vendor.FetchMasterOUI()
            if err != nil {
                status.SetText("Error downloading IEEE OUI list: " + err.Error())
                updateBtn.Enable()
                return
            }

            diff := vendor.DiffOUI(local, master)

            summary := fmt.Sprintf(
                "New Vendors: %d\nUpdated Vendors: %d\nRemoved Vendors: %d\nMaster Count: %d\nLocal Count: %d",
                len(diff.NewVendors),
                len(diff.UpdatedVendors),
                len(diff.RemovedVendors),
                diff.TotalMaster,
                diff.TotalLocal,
            )

            progress.SetText(summary)
            status.SetText("Update ready. Click 'Apply Update' to finish.")

            applyBtn.Enable()
        }()
    }

    applyBtn.OnTapped = func() {
        applyBtn.Disable()
        status.SetText("Applying update…")

        go func() {
            local, _ := vendor.LoadLocalOUI()
            master, _ := vendor.FetchMasterOUI()

            merged := vendor.ApplyOUIUpdate(local, master)

            if err := vendor.SaveLocalOUI(merged); err != nil {
                status.SetText("Error saving updated database: " + err.Error())
                return
            }

            status.SetText("Vendor database successfully updated!")
            progress.SetText("")
        }()
    }

    modal.Show()
}