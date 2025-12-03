# QuietScan

QuietScan is a lightweight, polite network discovery tool with a cross-platform GUI. It passively enumerates devices on your subnet and displays their IP, MAC, vendor, and hostname without triggering EDR or IDS alerts.

## Features

- **Polite Scanning**: Uses randomized delays (400-700ms) between pings to avoid triggering security alerts
- **Cross-Platform**: Supports Windows, macOS (Intel & Apple Silicon), and Linux
- **Modern GUI**: Built with Fyne framework for a native look and feel
- **MAC Vendor Lookup**: Automatic vendor identification using IEEE OUI database
- **Hostname Resolution**: Resolves device hostnames when available
- **Scan History**: Keeps track of the last 5 scans for comparison
- **CSV Export**: Export scan results with metadata to CSV files
- **OUI Database Updates**: Update MAC vendor database directly from IEEE sources
- **Progress Tracking**: Real-time progress bar during scans
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

1. Launch QuietScan
2. Click "Scan Now" to start a polite network scan
3. View results in the table showing IP, MAC, Vendor, and Hostname
4. Use "Show History" to view previous scans
5. Export results to CSV via Tools → Export to CSV
6. Update MAC vendor database via Tools → Update MAC Vendor Database

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