package main

import (
	"flag"
	"fmt"
	"os"
	"quietscan/assets"
	"quietscan/scanner"
	"quietscan/storage"
	"quietscan/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Parse command-line flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "QuietScan - Polite network discovery tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "QuietScan is designed to be conservative by default. Larger or more aggressive scans\n")
		fmt.Fprintf(os.Stderr, "are opt-in via flags. Default settings prioritize network politeness.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	allowLargeRanges := flag.Bool("allow-large-ranges", false, "Allow scanning ranges larger than /24 (up to 1024 hosts)")
	concurrency := flag.Int("concurrency", scanner.DefaultConcurrency, fmt.Sprintf("Number of concurrent workers (default: %d, max: %d). Higher values scan faster but generate more network traffic.", scanner.DefaultConcurrency, scanner.MaxConcurrency))
	delayMs := flag.Int("delay-ms", 0, "Per-target delay in milliseconds (default: 0). Adds delay between probes to reduce network load. A small random jitter is always applied.")
	timeoutMs := flag.Int("timeout-ms", scanner.DefaultTimeoutMs, fmt.Sprintf("Network operation timeout in milliseconds (default: %d, max: %d). Applies to ping, DNS lookups, and ARP queries.", scanner.DefaultTimeoutMs, scanner.MaxTimeoutMs))
	allowOverwrite := flag.Bool("overwrite", false, "Allow overwriting existing output files. By default, QuietScan will fail if the output file already exists.")
	flag.Parse()

	// Set flags in scanner package
	scanner.SetAllowLargeRanges(*allowLargeRanges)

	// Set concurrency with clamping and warning
	clamped := scanner.SetConcurrency(*concurrency)
	if clamped {
		fmt.Fprintf(os.Stderr, "Requested concurrency %d exceeds the maximum of %d. Clamping to %d to avoid flooding.\n", *concurrency, scanner.MaxConcurrency, scanner.MaxConcurrency)
	}

	// Set delay
	scanner.SetDelayMs(*delayMs)

	// Set timeout with clamping and warning
	timeoutClamped := scanner.SetTimeoutMs(*timeoutMs)
	if timeoutClamped {
		fmt.Fprintf(os.Stderr, "Requested timeout %d ms exceeds the maximum of %d ms. Clamping to %d ms.\n", *timeoutMs, scanner.MaxTimeoutMs, scanner.MaxTimeoutMs)
	}

	// Set file overwrite permission
	storage.SetAllowOverwrite(*allowOverwrite)

	quietApp := app.New()
	quietApp.SetIcon(assets.ResourceIconPng) // set icon from bundled resources

	// load latest results (if file exists)
	latest := storage.LoadLatestResults()

	window := quietApp.NewWindow("QuietScan")
	window.Resize(fyne.NewSize(900, 600))

	ui.RenderDashboard(window, latest)

	window.ShowAndRun()
}
