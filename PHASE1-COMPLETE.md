# Phase 1 Complete - Header Customization Foundation

## What Was Implemented

### New Files Created:
1. **internal/storage/appearance_config.go** - Configuration management
   - Defines header modes: default, text, ascii, none
   - Handles loading/saving configuration
   - Validates user input
   - Provides defaults for backwards compatibility

2. **internal/ui/header.go** - Header rendering engine
   - HeaderRenderer struct
   - Renders headers based on configuration
   - Supports 4 modes: default, text, ascii, none
   - Handles alignment, colors, padding

3. **test-phase1.sh** - Test script for verification
4. **appearance_config.example.yaml** - Sample configuration file

### Modified Files:
1. **internal/ui/styles.go**
   - Added global HeaderRenderer instance
   - Added InitializeHeaderRenderer() function
   - Updated wrapPageWithHeader() to use HeaderRenderer
   - Maintains fallback for safety

2. **internal/ui/app.go**
   - Calls InitializeHeaderRenderer() in NewApp()
   - Initializes header renderer at startup

## How to Test

### Test 1: Default Behavior (No Configuration)
```bash
# Run the application normally - should show "TERA" as before
go run .
```
Expected: Application works exactly as before, showing "TERA" header

### Test 2: Custom Text Header
```bash
# Create config file
mkdir -p ~/.config/tera
cat > ~/.config/tera/appearance_config.yaml << 'EOF'
appearance:
  header:
    mode: "text"
    custom_text: "🎵 My Radio 🎵"
    alignment: "center"
    width: 50
    color: "auto"
    bold: true
    padding_top: 1
    padding_bottom: 0
EOF

# Run application
go run .
```
Expected: Header shows "🎵 My Radio 🎵" instead of "TERA"

### Test 3: ASCII Art Header
```bash
# Create config with ASCII art
cat > ~/.config/tera/appearance_config.yaml << 'EOF'
appearance:
  header:
    mode: "ascii"
    ascii_art: |
      ╔══════════════════════╗
      ║  🎵  MY STATION  🎵  ║
      ╚══════════════════════╝
    alignment: "center"
    width: 50
    color: "99"
    padding_top: 1
    padding_bottom: 0
EOF

# Run application
go run .
```
Expected: Header shows custom ASCII art box

### Test 4: No Header (More Screen Space)
```bash
# Create config with no header
cat > ~/.config/tera/appearance_config.yaml << 'EOF'
appearance:
  header:
    mode: "none"
EOF

# Run application
go run .
```
Expected: No header shown, content starts immediately

### Test 5: Different Alignment
```bash
# Try left-aligned header
cat > ~/.config/tera/appearance_config.yaml << 'EOF'
appearance:
  header:
    mode: "text"
    custom_text: "TERA Radio"
    alignment: "left"
    width: 80
    color: "auto"
    bold: true
    padding_top: 1
EOF

# Run application
go run .
```
Expected: Header appears on the left side

## Quick Tips for Creating ASCII Art

### Online Tools:
- https://patorjk.com/software/taag/ (Text to ASCII Art Generator)
  - Try fonts: Standard, Slant, Banner, Big, Block
  - Copy the output and paste into config

### Using figlet (external tool):
```bash
# Install figlet (if not installed)
# macOS: brew install figlet
# Linux: apt-get install figlet

# Generate ASCII art
figlet -f slant "RADIO"

# Copy the output to your config file
```

### Box Drawing Characters:
```
Simple: ┌─┐ │ └─┘
Double: ╔═╗ ║ ╚═╝
Rounded: ╭─╮ │ ╰─╯
Heavy: ┏━┓ ┃ ┗━┛
```

## Configuration Options Reference

### mode
- `"default"` - Shows "TERA"
- `"text"` - Shows custom_text
- `"ascii"` - Shows ascii_art
- `"none"` - No header

### alignment
- `"left"` - Left-aligned
- `"center"` - Centered (default)
- `"right"` - Right-aligned

### color
- `"auto"` - Uses theme color (default)
- Any lipgloss color code: "99" (purple), "196" (red), "51" (cyan), etc.

### width
- Range: 10-120 characters
- Default: 50

### padding_top / padding_bottom
- Number of blank lines above/below header
- Default: 1 top, 0 bottom

### Limits
- custom_text: max 100 characters
- ascii_art: max 15 lines

## Troubleshooting

### Issue: Config changes not appearing
**Solution**: Restart the application (config is loaded at startup)

### Issue: ASCII art looks misaligned
**Solution**: 
- Check that all lines have similar length
- Adjust the `width` setting
- Try different `alignment` values

### Issue: Invalid config errors
**Solution**: 
- Check YAML syntax (use a YAML validator)
- Ensure mode is one of: default, text, ascii, none
- Verify ASCII art doesn't exceed 15 lines

### Issue: App crashes with custom config
**Solution**: 
- Delete ~/.config/tera/appearance_config.yaml
- App will use defaults
- Check error messages

## What's Next

### Phase 2: Settings UI
- Add appearance settings to Settings menu
- UI for selecting header mode
- Text input for custom_text
- Multi-line input for ascii_art
- Live preview of changes
- Save/Reset buttons

### Phase 3: Polish
- Add color picker for custom colors
- Import ASCII art from file
- Export current config
- Header templates/presets

## Backwards Compatibility

✓ **100% backwards compatible**
- If no config file exists → uses default "TERA"
- If config is invalid → falls back to default
- Existing installations work without changes
- No breaking changes to code or behavior

## Files You Can Safely Modify

You can manually edit these files to test different configurations:
- `~/.config/tera/appearance_config.yaml` - Your active config
- Copy examples from `appearance_config.example.yaml`

## Status

**Phase 1: COMPLETE ✓**
- Foundation implemented
- All core functionality working
- Fully backwards compatible
- Ready for manual testing

**Phase 2: NOT STARTED**
- Settings UI not yet implemented
- Currently must edit config file manually
- Will add UI in next phase
