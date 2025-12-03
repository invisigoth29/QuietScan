package ui

import (
	"fmt"
	"quietscan/vendordb"
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
	var masterSource string

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

			local, err := vendordb.LoadLocalOUI()
			if err != nil {
				status.SetText("Error loading local OUI file: " + err.Error())
				updateBtn.Enable()
				return
			}

			master, source, err := vendordb.FetchMasterOUI()
			if err != nil {
				status.SetText("Error downloading IEEE OUI list: " + err.Error())
				updateBtn.Enable()
				return
			}
			masterSource = source

			diff := vendordb.DiffOUI(local, master)

			summary := fmt.Sprintf(
				"Source: %s\nNew Vendors: %d\nUpdated Vendors: %d\nRemoved Vendors: %d\nMaster Count: %d\nLocal Count: %d",
				masterSource,
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
			local, _ := vendordb.LoadLocalOUI()
			master, source, _ := vendordb.FetchMasterOUI()

			merged := vendordb.ApplyOUIUpdate(local, master)

			if err := vendordb.SaveLocalOUI(merged); err != nil {
				status.SetText("Error saving updated database: " + err.Error())
				return
			}

			status.SetText(fmt.Sprintf("Vendor database successfully updated! (source: %s)", source))
			progress.SetText("")
		}()
	}

	modal.Show()
}
