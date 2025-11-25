package ui

import (
    "quietscan/storage"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
)

type ResultsTable struct {
    Table   *widget.Table
    Devices []storage.DeviceEntry
}

func NewResultsTable(latest *storage.ScanResult) *ResultsTable {
    rt := &ResultsTable{}
    if latest != nil {
        rt.Devices = latest.Devices
    }

    rt.Table = widget.NewTable(
        func() (int, int) { return len(rt.Devices) + 1, 4 },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(i widget.TableCellID, o fyne.CanvasObject) {
            label := o.(*widget.Label)
            if i.Row == 0 {
                // headers
                switch i.Col {
                case 0:
                    label.SetText("IP Address")
                case 1:
                    label.SetText("MAC")
                case 2:
                    label.SetText("Vendor")
                case 3:
                    label.SetText("Hostname")
                }
            } else {
                dev := rt.Devices[i.Row-1]
                switch i.Col {
                case 0:
                    label.SetText(dev.IP)
                case 1:
                    label.SetText(dev.MAC)
                case 2:
                    label.SetText(dev.Vendor)
                case 3:
                    label.SetText(dev.Hostname)
                }
            }
        },
    )
    return rt
}

func (rt *ResultsTable) UpdateData(r *storage.ScanResult) {
    if r != nil {
        rt.Devices = r.Devices
    } else {
        rt.Devices = nil
    }
    rt.Table.Refresh()
}