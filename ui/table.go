package ui

import (
	"strings"
	"time"

	"quietscan/types"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ResultsTable struct {
	Table           *tableWithContextMenu
	Devices         []types.DeviceEntry // Deprecated: use FilteredDevices instead
	AllDevices      []types.DeviceEntry // Stores unfiltered data
	FilteredDevices []types.DeviceEntry // Stores filtered results
	window          fyne.Window
}

// tableWithContextMenu wraps widget.Table and implements SecondaryTappable interface
type tableWithContextMenu struct {
	widget.BaseWidget
	table  *widget.Table
	rt     *ResultsTable
	window fyne.Window
}

// tappableOverlay is an invisible overlay that captures secondary taps
// but allows other interactions to pass through to the table below
type tappableOverlay struct {
	widget.BaseWidget
	onSecondaryTap func(*fyne.PointEvent)
}

func (o *tappableOverlay) TappedSecondary(ev *fyne.PointEvent) {
	if o.onSecondaryTap != nil {
		o.onSecondaryTap(ev)
	}
}

func (o *tappableOverlay) CreateRenderer() fyne.WidgetRenderer {
	// Return an empty renderer - this widget is invisible
	return widget.NewSimpleRenderer(container.NewMax())
}

func newTappableOverlay(onSecondaryTap func(*fyne.PointEvent)) *tappableOverlay {
	overlay := &tappableOverlay{
		onSecondaryTap: onSecondaryTap,
	}
	overlay.ExtendBaseWidget(overlay)
	return overlay
}

// Custom renderer that layers the table with an overlay for tap capture
type tableWithContextMenuRenderer struct {
	wrapper *tableWithContextMenu
	overlay *tappableOverlay
	objects []fyne.CanvasObject
}

func (r *tableWithContextMenuRenderer) Layout(size fyne.Size) {
	r.wrapper.table.Resize(size)
	r.overlay.Resize(size)
}

func (r *tableWithContextMenuRenderer) MinSize() fyne.Size {
	return r.wrapper.table.MinSize()
}

func (r *tableWithContextMenuRenderer) Refresh() {
	r.wrapper.table.Refresh()
}

func (r *tableWithContextMenuRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *tableWithContextMenuRenderer) Destroy() {}

func (t *tableWithContextMenu) CreateRenderer() fyne.WidgetRenderer {
	overlay := newTappableOverlay(func(ev *fyne.PointEvent) {
		t.TappedSecondary(ev)
	})

	return &tableWithContextMenuRenderer{
		wrapper: t,
		overlay: overlay,
		objects: []fyne.CanvasObject{t.table, overlay},
	}
}

func (t *tableWithContextMenu) Refresh() {
	t.table.Refresh()
	t.BaseWidget.Refresh()
}

func (t *tableWithContextMenu) SetColumnWidth(colID int, width float32) {
	t.table.SetColumnWidth(colID, width)
}

func (t *tableWithContextMenu) TappedSecondary(ev *fyne.PointEvent) {
	if t.rt == nil || t.window == nil {
		return
	}

	// Calculate row and column from the tap position
	// We need to account for the table's scroll position and cell sizes
	row := t.getRowAtPosition(ev.Position.Y)
	col := t.getColAtPosition(ev.Position.X)

	if row >= 0 && col >= 0 {
		var cellText string
		if row == 0 {
			// headers
			switch col {
			case 0:
				cellText = "IP Address"
			case 1:
				cellText = "MAC"
			case 2:
				cellText = "Vendor"
			case 3:
				cellText = "Hostname"
			}
		} else if row-1 < len(t.rt.FilteredDevices) {
			dev := t.rt.FilteredDevices[row-1]
			switch col {
			case 0:
				cellText = dev.IP
			case 1:
				cellText = dev.MAC
			case 2:
				cellText = dev.Vendor
			case 3:
				cellText = dev.Hostname
			}
		}
		if cellText != "" {
			showCopyMenuAtPosition(t.rt, cellText, ev.AbsolutePosition, t.window)
		}
	}
}

// getRowAtPosition calculates which row is at the given Y position
func (t *tableWithContextMenu) getRowAtPosition(y float32) int {
	// Use a default row height (Fyne tables typically use ~30px per row)
	rowHeight := float32(30)
	row := int(y / rowHeight)
	return row
}

// getColAtPosition calculates which column is at the given X position
func (t *tableWithContextMenu) getColAtPosition(x float32) int {
	colWidths := []float32{140, 160, 200, 200} // IP, MAC, Vendor, Hostname
	accumulatedWidth := float32(0)

	for i, width := range colWidths {
		if x < accumulatedWidth+width {
			return i
		}
		accumulatedWidth += width
	}
	// Default to last column if beyond all columns
	return len(colWidths) - 1
}

func NewResultsTable(latest *types.ScanResult) *ResultsTable {
	return NewResultsTableWithWindow(latest, nil)
}

func NewResultsTableWithWindow(latest *types.ScanResult, window fyne.Window) *ResultsTable {
	rt := &ResultsTable{window: window}
	if latest != nil {
		rt.AllDevices = latest.Devices
		rt.FilteredDevices = latest.Devices
		rt.Devices = latest.Devices // Keep for backward compatibility
	} else {
		rt.AllDevices = []types.DeviceEntry{}
		rt.FilteredDevices = []types.DeviceEntry{}
		rt.Devices = []types.DeviceEntry{}
	}

	innerTable := widget.NewTable(
		func() (int, int) { return len(rt.FilteredDevices) + 1, 4 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			var cellText string
			if i.Row == 0 {
				// headers
				switch i.Col {
				case 0:
					cellText = "IP Address"
				case 1:
					cellText = "MAC"
				case 2:
					cellText = "Vendor"
				case 3:
					cellText = "Hostname"
				}
			} else {
				dev := rt.FilteredDevices[i.Row-1]
				switch i.Col {
				case 0:
					cellText = dev.IP
				case 1:
					cellText = dev.MAC
				case 2:
					cellText = dev.Vendor
				case 3:
					cellText = dev.Hostname
				}
			}
			label.SetText(cellText)
		},
	)

	// Widen columns so headers/content don't overlap.
	innerTable.SetColumnWidth(0, 140) // IP Address
	innerTable.SetColumnWidth(1, 160) // MAC
	innerTable.SetColumnWidth(2, 200) // Vendor
	innerTable.SetColumnWidth(3, 200) // Hostname

	wrapper := &tableWithContextMenu{
		table:  innerTable,
		rt:     rt,
		window: window,
	}

	// Extend base widget to enable TappedSecondary
	wrapper.ExtendBaseWidget(wrapper)

	rt.Table = wrapper

	return rt
}

func (rt *ResultsTable) UpdateData(r *types.ScanResult) {
	if r != nil {
		rt.AllDevices = r.Devices
		rt.Devices = r.Devices // Keep for backward compatibility
	} else {
		rt.AllDevices = []types.DeviceEntry{}
		rt.Devices = []types.DeviceEntry{}
	}
	// Reset filter when new data loads
	rt.FilteredDevices = rt.AllDevices
	rt.Table.Refresh()
}

// Filter filters the table results based on the search query
func (rt *ResultsTable) Filter(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		rt.FilteredDevices = rt.AllDevices
	} else {
		queryLower := strings.ToLower(query)
		rt.FilteredDevices = []types.DeviceEntry{}
		for _, dev := range rt.AllDevices {
			// Check if query matches IP, MAC, Vendor, or Hostname (case-insensitive)
			if strings.Contains(strings.ToLower(dev.IP), queryLower) ||
				strings.Contains(strings.ToLower(dev.MAC), queryLower) ||
				strings.Contains(strings.ToLower(dev.Vendor), queryLower) ||
				strings.Contains(strings.ToLower(dev.Hostname), queryLower) {
				rt.FilteredDevices = append(rt.FilteredDevices, dev)
			}
		}
	}
	rt.Table.Refresh()
}

func (rt *ResultsTable) SetWindow(window fyne.Window) {
	rt.window = window
	if rt.Table != nil {
		rt.Table.window = window
	}
}

// showCopyMenuAtPosition shows a context menu at the specified position
func showCopyMenuAtPosition(rt *ResultsTable, content string, pos fyne.Position, window fyne.Window) {
	if content == "" {
		return
	}

	menu := fyne.NewMenu("",
		fyne.NewMenuItem("Copy", func() {
			clipboard := window.Clipboard()
			clipboard.SetContent(content)
			showCopyNotification(window)
		}),
	)

	popup := widget.NewPopUpMenu(menu, window.Canvas())
	popup.ShowAtPosition(pos)
}

// showCopyNotification shows a brief notification that content was copied
func showCopyNotification(window fyne.Window) {
	notification := widget.NewLabel("Copied to clipboard")
	notification.Alignment = fyne.TextAlignCenter

	popup := widget.NewPopUp(container.NewVBox(notification), window.Canvas())
	popup.Resize(fyne.NewSize(200, 50))

	// Position near cursor or center
	popup.ShowAtPosition(fyne.NewPos(100, 100))

	// Auto-hide after 1.5 seconds
	go func() {
		time.Sleep(1500 * time.Millisecond)
		popup.Hide()
	}()
}
