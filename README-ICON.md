# Windows Icon Setup

## Issue: Icon Not Showing

If your Windows executable is showing the generic icon instead of your custom icon, try these steps:

### 1. Create Multi-Size ICO File

Windows works best with ICO files that contain multiple sizes. The current `icon.ico` only has one size (256x256).

**Recommended: Use an online converter to create a multi-size ICO:**

1. Go to https://convertio.co/png-ico/ or https://cloudconvert.com/png-ico
2. Upload `assets/icon.png`
3. **Important:** Enable "Multiple sizes" option if available
4. Download the converted `icon.ico`
5. Replace `assets/icon.ico` with the new file

**Or use ImageMagick (if installed):**
```bash
magick convert assets/icon.png -define icon:auto-resize=256,128,64,48,32,16 assets/icon.ico
```

### 2. Rebuild the Executable

After updating the icon file:

```bash
build-windows.bat
```

### 3. Clear Windows Icon Cache

Windows caches icons. After rebuilding:

1. Press `Win + R`
2. Type: `ie4uinit.exe -show`
3. Press Enter
4. Or restart Windows Explorer

### 4. Verify Icon is Embedded

Use Resource Hacker (http://www.angusj.com/resourcehacker/) to check if the icon is embedded in `quietscan.exe`:
- Open `quietscan.exe` in Resource Hacker
- Look under "Icon" section
- You should see icon resources (ID 1, etc.)

### 5. Alternative: Use go-winres with explicit output

If the icon still doesn't work, try manually running:

```bash
go-winres simply --icon assets\icon.ico --manifest gui --out resource.syso
go build -mod=mod -ldflags "-H=windowsgui" -o quietscan.exe ./cmd/quietscan
```

## Troubleshooting

- **Icon still generic:** Make sure `resource.syso` exists before building
- **Build fails:** Check that `go-winres` is installed: `go install github.com/tc-hib/go-winres@latest`
- **Icon shows in Resource Hacker but not in Explorer:** Clear icon cache (step 3)


