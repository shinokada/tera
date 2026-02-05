# TERA Customizable Headers

## Overview

TERA now supports customizable headers! You can replace the default "TERA" header with your own text, ASCII art, or remove it entirely for more screen space.

## Quick Start

### Option 1: Keep Default
Do nothing - TERA shows "TERA" by default

### Option 2: Custom Text
```bash
mkdir -p ~/.config/tera
cat > ~/.config/tera/appearance_config.yaml << 'EOF'
appearance:
  header:
    mode: "text"
    custom_text: "🎵 My Radio 🎵"
EOF
```

### Option 3: ASCII Art
```bash
mkdir -p ~/.config/tera
cat > ~/.config/tera/appearance_config.yaml << 'EOF'
appearance:
  header:
    mode: "ascii"
    ascii_art: |
      ╔══════════════════╗
      ║  MY RADIO 🎵  ║
      ╚══════════════════╝
EOF
```

### Option 4: No Header
```bash
mkdir -p ~/.config/tera
cat > ~/.config/tera/appearance_config.yaml << 'EOF'
appearance:
  header:
    mode: "none"
EOF
```

## Configuration File

Location: `~/.config/tera/appearance_config.yaml`

### Full Configuration Options

```yaml
appearance:
  header:
    mode: "default"           # "default", "text", "ascii", or "none"
    custom_text: ""           # Text to display (if mode is "text")
    ascii_art: ""             # ASCII art to display (if mode is "ascii")
    alignment: "center"       # "left", "center", or "right"
    width: 50                 # Width in characters (10-120)
    color: "auto"             # "auto" or color code like "99", "196"
    bold: true                # true or false
    padding_top: 1            # Lines above header (0+)
    padding_bottom: 0         # Lines below header (0+)
```

### Header Modes

#### 1. Default Mode
Shows "TERA" in blue (or theme color)
```yaml
appearance:
  header:
    mode: "default"
```

#### 2. Text Mode
Shows custom text
```yaml
appearance:
  header:
    mode: "text"
    custom_text: "🎵 John's Radio Station 🎵"
```

Limits:
- Maximum 100 characters
- Can use emoji and Unicode

#### 3. ASCII Art Mode
Shows ASCII art banner
```yaml
appearance:
  header:
    mode: "ascii"
    ascii_art: |
       ____      _    ____ ___ ___  
      |  _ \    / \  |  _ \_ _/ _ \ 
      | |_) |  / _ \ | | | | | | | |
      |  _ <  / ___ \| |_| | | |_| |
      |_| \_\/_/   \_\____/___\___/ 
```

Limits:
- Maximum 15 lines
- Use standard ASCII or Unicode box-drawing characters

#### 4. None Mode
No header, maximum screen space
```yaml
appearance:
  header:
    mode: "none"
```

## Creating ASCII Art

### Online Tools

**Best: Text to ASCII Art Generator**
- https://patorjk.com/software/taag/
- Type your text
- Choose a font (try: Standard, Slant, Banner, Big)
- Copy the output

**Other Tools:**
- https://www.ascii-art-generator.org/
- https://fsymbols.com/generators/carty/

### Using figlet (Command Line)

```bash
# Install figlet
# macOS: brew install figlet
# Linux: apt-get install figlet

# Generate ASCII art
figlet -f slant "RADIO"

# Try different fonts
figlet -f banner "RADIO"
figlet -f big "RADIO"

# List all fonts
figlet -f list
```

### Box-Drawing Characters

```
Simple:  ┌─┬─┐
         ├─┼─┤
         └─┴─┘

Double:  ╔═╦═╗
         ╠═╬═╣
         ╚═╩═╝

Rounded: ╭─┬─╮
         ├─┼─┤
         ╰─┴─╯

Heavy:   ┏━┳━┓
         ┣━╋━┫
         ┗━┻━┛
```

Copy these characters and arrange them to create boxes or borders.

## Examples

### Example 1: Simple Text
```yaml
appearance:
  header:
    mode: "text"
    custom_text: "RADIO STATION"
    alignment: "center"
    width: 50
    color: "auto"
    bold: true
    padding_top: 1
```

### Example 2: Emoji Banner
```yaml
appearance:
  header:
    mode: "text"
    custom_text: "🎵 🎧 MUSIC TIME 🎧 🎵"
    alignment: "center"
    width: 60
    color: "99"
    bold: true
    padding_top: 1
```

### Example 3: ASCII Box
```yaml
appearance:
  header:
    mode: "ascii"
    ascii_art: |
      ╔════════════════════════════════╗
      ║                                ║
      ║     🎵  RADIO STATION  🎵      ║
      ║                                ║
      ╚════════════════════════════════╝
    alignment: "center"
    width: 70
    color: "51"
```

### Example 4: Large Banner
```yaml
appearance:
  header:
    mode: "ascii"
    ascii_art: |
      ████████╗███████╗██████╗  █████╗ 
      ╚══██╔══╝██╔════╝██╔══██╗██╔══██╗
         ██║   █████╗  ██████╔╝███████║
         ██║   ██╔══╝  ██╔══██╗██╔══██║
         ██║   ███████╗██║  ██║██║  ██║
         ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝
    alignment: "center"
    width: 80
    color: "auto"
```

### Example 5: Left-Aligned
```yaml
appearance:
  header:
    mode: "text"
    custom_text: "► TERA Radio"
    alignment: "left"
    width: 80
    color: "auto"
    bold: true
    padding_top: 1
```

## Color Codes

When `color` is not "auto", you can use these color codes:

### Basic Colors
- `"0"` - Black
- `"1"` - Red
- `"2"` - Green
- `"3"` - Yellow
- `"4"` - Blue
- `"5"` - Magenta
- `"6"` - Cyan
- `"7"` - White

### Extended Colors (256-color palette)
- `"51"` - Cyan
- `"99"` - Purple
- `"196"` - Bright Red
- `"226"` - Bright Yellow
- `"46"` - Bright Green
- `"21"` - Bright Blue

Or use "auto" to match your current theme.

## Testing Your Configuration

### Interactive Demo
```bash
cd /Users/shinichiokada/Terminal-Tools/tera
chmod +x demo-headers.sh
./demo-headers.sh
```

This script will show you different header modes interactively.

### Manual Testing
1. Edit `~/.config/tera/appearance_config.yaml`
2. Save the file
3. Restart TERA
4. See your changes!

### Quick Test
```bash
# Test with example config
cp appearance_config.example.yaml ~/.config/tera/appearance_config.yaml

# Edit to try different modes
nano ~/.config/tera/appearance_config.yaml

# Run TERA
./tera
```

## Troubleshooting

### Changes Not Appearing
**Problem**: Edited config but header didn't change
**Solution**: Restart TERA (config loads at startup)

### ASCII Art Looks Wrong
**Problem**: ASCII art is misaligned or cut off
**Solutions**:
- Increase `width` setting
- Check that all lines are similar length
- Try different `alignment` values
- Reduce padding

### Config Errors
**Problem**: TERA shows error or uses default
**Solutions**:
- Check YAML syntax (indentation matters!)
- Verify mode is: default, text, ascii, or none
- Ensure ASCII art is max 15 lines
- Ensure custom_text is max 100 characters

### Want to Reset
**Problem**: Want to go back to default
**Solution**: Delete config file
```bash
rm ~/.config/tera/appearance_config.yaml
```

## Tips & Tricks

### Tip 1: Test Before Committing
Keep a backup of your config:
```bash
cp ~/.config/tera/appearance_config.yaml ~/.config/tera/appearance_config.yaml.backup
```

### Tip 2: Create Multiple Configs
Save different headers and swap them:
```bash
# Save favorites
cp ~/.config/tera/appearance_config.yaml ~/.config/tera/minimal-header.yaml
cp ~/.config/tera/appearance_config.yaml ~/.config/tera/fancy-header.yaml

# Switch between them
cp ~/.config/tera/minimal-header.yaml ~/.config/tera/appearance_config.yaml
```

### Tip 3: Use the Example File
The `appearance_config.example.yaml` has many examples commented out. Uncomment one to try it!

### Tip 4: Preview Online
Before adding ASCII art to your config, preview it in your terminal:
```bash
cat << 'EOF'
   ____      _    ____ ___ ___  
  |  _ \    / \  |  _ \_ _/ _ \ 
  | |_) |  / _ \ | | | | | | | |
  |  _ <  / ___ \| |_| | | |_| |
  |_| \_\/_/   \_\____/___\___/ 
EOF
```

## Limitations

- Maximum 15 lines for ASCII art
- Maximum 100 characters for custom text
- Width range: 10-120 characters
- Config loads at startup (restart needed for changes)
- Must manually edit config (Settings UI coming in Phase 2)

## Future Features (Phase 2)

Coming soon:
- Settings UI for easy configuration
- Live preview of changes
- No need to manually edit config files
- Text input widget
- Multi-line ASCII art editor
- Color picker
- Save/Reset buttons

## Getting Help

### Documentation Files
- `IMPLEMENTATION-SUMMARY.md` - Technical details
- `PHASE1-COMPLETE.md` - Testing guide
- `appearance_config.example.yaml` - Example configurations

### Scripts
- `test-phase1.sh` - Run automated tests
- `demo-headers.sh` - Interactive demo

### Support
If you encounter issues:
1. Check this README
2. Look at example configs
3. Run the demo script
4. Check for typos in YAML syntax

## Contributing

Want to share your cool header designs? Create a PR with your config in a new `community-headers/` directory!

---

**Status**: Phase 1 Complete ✓  
**Version**: 1.0  
**Last Updated**: February 2026
