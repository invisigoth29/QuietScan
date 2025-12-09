# QuietScan

QuietScan is a lightweight, polite network discovery tool with a cross-platform GUI. It passively enumerates devices on your subnet and displays their IP, MAC, vendor, and hostname without triggering EDR or IDS alerts.

## Features

- **Polite Scanning**: Uses randomized delays (400-700ms) between pings to avoid triggering security alerts
- **Cross-Platform**: Supports Windows, macOS (Intel & Apple Silicon), and Linux
- **Modern GUI**: Built with Fyne framework for a native look and feel
- **Adapter Selection**: Choose which network adapter to scan from via dropdown menu
- **Flexible Subnet Configuration**: Manually specify subnet ranges using CIDR (192.168.1.0/24) or range notation (192.168.1.1-254)
- **Real-time Search/Filter**: Instantly filter scan results by IP, MAC, Vendor, or Hostname
- **Context Menu**: Right-click any table cell to copy its value to clipboard
- **MAC Vendor Lookup**: Automatic vendor identification using IEEE OUI database
- **Hostname Resolution**: Resolves device hostnames when available
- **Scan History**: Keeps track of the last 5 scans for comparison
- **CSV Export**: Export scan results with metadata to CSV files
- **OUI Database Updates**: Update MAC vendor database directly from IEEE sources
- **Progress Tracking**: Real-time progress bar during scans with remaining host count
- **Self-Contained**: Bundled resources (icon, OUI database) for easy distribution

## Building

### macOS
```bash
./build-macos.sh
```

### All Platforms
```bash
./build-all.sh
```

This will create executables for:
- Windows: `dist/quietscan-windows.exe`
- macOS Intel: `QuietScan-macos-intel`
- macOS ARM64: `QuietScan-macos-arm64`

## Usage

1. **Launch QuietScan** - Start the application
2. **Select Network Adapter** (optional) - Choose which adapter to scan from using the dropdown
3. **Configure Subnet** (optional) - Manually enter a subnet range:
   - CIDR notation: `192.168.1.0/24`
   - Range notation: `192.168.1.1-254`
   - Single IP: `192.168.1.100`
4. **Scan Network** - Click "Scan Now" to start a polite network scan
5. **Filter Results** - Use the search bar to filter by IP, MAC, Vendor, or Hostname
6. **Copy Data** - Right-click any cell in the results table to copy its value
7. **View History** - Click "Show History" to view previous scans
8. **Export Results** - Use Tools → Export to CSV to save scan results
9. **Update Database** - Use Tools → Update MAC Vendor Database to refresh the OUI database

## How It Works

QuietScan uses a combination of:
- **ARP table enumeration** to discover devices
- **ICMP ping** with randomized delays for polite scanning
- **IEEE OUI database** for MAC vendor identification
- **DNS reverse lookup** for hostname resolution

The scanning process is designed to be slow and polite, avoiding rapid-fire pings that might trigger intrusion detection systems.

## Requirements

- Go 1.21 or later
- CGO enabled (for GUI support)
- Network access to your local subnet