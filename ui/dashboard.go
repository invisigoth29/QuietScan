package ui

import (
	"fmt"
	"path/filepath"
	"strings"
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

	table := NewResultsTableWithWindow(latest, win)

	// Create progress bar
	progressBar := widget.NewProgressBar()
	progressBar.Hide() // Hide initially
	progressLabel := widget.NewLabel("")
	progressLabel.Hide()

	// Get all adapters for dropdown
	adapters := scanner.GetAllAdapters()
	if len(adapters) == 0 {
		adapters = []string{adapter} // Fallback if no adapters found
	}

	// Ensure the auto-detected adapter is in the list
	adapterInList := false
	for _, a := range adapters {
		if a == adapter {
			adapterInList = true
			break
		}
	}
	if !adapterInList && adapter != "Unknown" {
		adapters = append(adapters, adapter)
	}

	// Create adapter dropdown
	adapterSelect := widget.NewSelect(adapters, nil)
	if adapterInList || adapter != "Unknown" {
		adapterSelect.SetSelected(adapter)
	} else if len(adapters) > 0 {
		adapterSelect.SetSelected(adapters[0])
		adapter = adapters[0]
		subnet = scanner.GetAdapterSubnet(adapter)
	}

	// Create subnet entry field that allows manual input
	subnetEntry := widget.NewEntry()
	subnetEntry.SetPlaceHolder("e.g., 192.168.1.0/24 or 192.168.1.1-254")
	subnetEntry.SetText(subnet)

	// Create validation label for subnet entry
	subnetErrorLabel := widget.NewLabel("")
	subnetErrorLabel.Hide()

	adapterLabel := widget.NewLabel("Adapter: " + adapter)

	// Store selected adapter for use in scan button
	var selectedAdapter string = adapter

	// Update subnet entry when adapter changes
	adapterSelect.OnChanged = func(selected string) {
		selectedAdapter = selected
		if selected != "" {
			selectedSubnet := scanner.GetAdapterSubnet(selected)
			subnetEntry.SetText(selectedSubnet)
			subnetErrorLabel.Hide()
			adapterLabel.SetText("Adapter: " + selected)
		}
	}

	// Validate subnet entry on change
	subnetEntry.OnChanged = func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			subnetErrorLabel.Hide()
			return
		}

		_, err := scanner.ParseSubnetInput(text)
		if err != nil {
			subnetErrorLabel.SetText("Invalid format: " + err.Error())
			subnetErrorLabel.Show()
		} else {
			subnetErrorLabel.Hide()
		}
	}

	scanBtn := widget.NewButton("Scan Now", func() {
		// Get subnet input
		subnetInput := strings.TrimSpace(subnetEntry.Text)

		// Use selected adapter or fallback to auto-detected
		adapterToUse := selectedAdapter
		if adapterToUse == "" {
			adapterToUse = adapter
		}

		// Determine subnet string for display
		subnetDisplay := subnet
		if subnetInput != "" {
			subnetDisplay = subnetInput
		} else {
			// Get subnet from adapter if entry is empty
			if adapterToUse != "" {
				subnetDisplay = scanner.GetAdapterSubnet(adapterToUse)
			}
		}

		// Validate subnet input if provided
		var ips []string
		var err error
		if subnetInput != "" {
			ips, err = scanner.ParseSubnetInput(subnetInput)
			if err != nil {
				// Display clear, user-friendly error message
				showMessageModal(win, "Invalid Subnet", fmt.Sprintf("%s\n\nPlease use:\n• CIDR notation: 192.168.1.0/24\n• Range notation: 192.168.1.1-254\n• Single IP: 192.168.1.1", err.Error()))
				return
			}
		} else {
			// Use auto-detected subnet from adapter
			ips = scanner.EnumerateSubnetForAdapter(adapterToUse)
			if len(ips) == 0 {
				showMessageModal(win, "No Valid Subnet", "Could not detect a valid subnet from the selected adapter. Please specify a subnet manually.")
				return
			}
		}

		if len(ips) == 0 {
			showMessageModal(win, "No Valid Targets", "No valid scan targets found. Please check your subnet input.")
			return
		}

		statusLabel.SetText("Scanning… Polite Mode Enabled")
		progressBar.Show()
		progressLabel.Show()
		progressBar.SetValue(0)

		go func() {
			res := scanner.RunPoliteScanWithIPs(ips, adapterToUse, subnetDisplay, func(current, total int) {
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

	// Create search entry field
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by IP, MAC, Vendor, or Hostname...")

	// Clear button for search
	clearBtn := widget.NewButton("✕", func() {
		searchEntry.SetText("")
		table.Filter("")
	})
	clearBtn.Importance = widget.LowImportance

	// Search entry callback
	searchEntry.OnChanged = func(query string) {
		table.Filter(query)
		if query == "" {
			clearBtn.Hide()
		} else {
			clearBtn.Show()
		}
	}
	clearBtn.Hide() // Hide initially when search is empty

	// Search container with entry and clear button
	searchContainer := container.NewBorder(nil, nil, nil, clearBtn, searchEntry)

	header := container.NewVBox(
		adapterSelect,
		adapterLabel,
		widget.NewLabel("Subnet:"),
		subnetEntry,
		subnetErrorLabel,
		lastScanLabel,
		container.NewHBox(scanBtn, historyBtn, toolsBtn),
	)

	// Bottom section with status, progress bar, and progress label
	bottomSection := container.NewVBox(
		statusLabel,
		progressLabel,
		progressBar,
	)

	// Table container with search above it
	tableWithSearch := container.NewBorder(searchContainer, nil, nil, nil, table.Table)

	content := container.NewBorder(header, bottomSection, nil, nil, tableWithSearch)
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
