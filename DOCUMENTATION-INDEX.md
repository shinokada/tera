# TERA Customizable Headers - Documentation Index

## 🚀 Quick Start

**Start here:** [START-HERE.md](START-HERE.md)

## 📚 Documentation

### For Users
- **[HEADER-CUSTOMIZATION-README.md](HEADER-CUSTOMIZATION-README.md)** - Complete user guide with examples
- **[appearance_config.example.yaml](appearance_config.example.yaml)** - Sample configurations to copy

### For Testing
- **[PHASE1-COMPLETE.md](PHASE1-COMPLETE.md)** - Detailed testing instructions
- **[test-phase1.sh](test-phase1.sh)** - Automated test script
- **[demo-headers.sh](demo-headers.sh)** - Interactive demo

### For Developers
- **[IMPLEMENTATION-SUMMARY.md](IMPLEMENTATION-SUMMARY.md)** - Technical architecture overview
- **[customizable-header-design.md](customizable-header-design.md)** - Original design document

## 🎯 What Was Implemented

### Phase 1: Foundation ✅ COMPLETE

**New Files:**
- `internal/storage/appearance_config.go` - Configuration system
- `internal/ui/header.go` - Header rendering engine

**Modified Files:**
- `internal/ui/styles.go` - Integrated header renderer
- `internal/ui/app.go` - Initializes renderer

**Features:**
- 4 header modes: default, text, ASCII art, none
- Customizable alignment, colors, padding
- Backwards compatible
- Input validation
- Global application

## 🧪 Testing

### Quick Test
```bash
# Run the demo
./demo-headers.sh

# Or test manually
go run .
```

### Comprehensive Test
```bash
# Run automated tests
./test-phase1.sh
```

## 📖 Usage Examples

### Example 1: Custom Text
```yaml
appearance:
  header:
    mode: "text"
    custom_text: "🎵 My Radio 🎵"
```

### Example 2: ASCII Art
```yaml
appearance:
  header:
    mode: "ascii"
    ascii_art: |
      ╔══════════════╗
      ║  MY RADIO  ║
      ╚══════════════╝
```

### Example 3: No Header
```yaml
appearance:
  header:
    mode: "none"
```

## 🔧 Configuration

**Location:** `~/.config/tera/appearance_config.yaml`

**See examples:** [appearance_config.example.yaml](appearance_config.example.yaml)

## 🎨 Creating ASCII Art

**Online tools:**
- https://patorjk.com/software/taag/

**Command line:**
```bash
figlet -f slant "RADIO"
```

## 📋 Documentation Quick Links

| Document | Purpose |
|----------|---------|
| [START-HERE.md](START-HERE.md) | Quick start and overview |
| [HEADER-CUSTOMIZATION-README.md](HEADER-CUSTOMIZATION-README.md) | Complete user guide |
| [PHASE1-COMPLETE.md](PHASE1-COMPLETE.md) | Testing guide |
| [IMPLEMENTATION-SUMMARY.md](IMPLEMENTATION-SUMMARY.md) | Technical details |
| [appearance_config.example.yaml](appearance_config.example.yaml) | Config examples |
| [test-phase1.sh](test-phase1.sh) | Test automation |
| [demo-headers.sh](demo-headers.sh) | Interactive demo |

## 🚧 Roadmap

### Phase 1: Foundation ✅ COMPLETE
- Configuration system
- Header renderer
- 4 modes supported
- Documentation

### Phase 2: Settings UI 🔜 COMING NEXT
- Appearance settings menu
- Interactive editor
- Text input widget
- ASCII art input
- Live preview
- Save/Reset buttons

### Phase 3: Polish 📅 FUTURE
- Color picker
- Import from file
- Export config
- Header templates

## 💡 Tips

1. **Backup your config** before experimenting:
   ```bash
   cp ~/.config/tera/appearance_config.yaml{,.backup}
   ```

2. **Try the demo script** to see all modes:
   ```bash
   ./demo-headers.sh
   ```

3. **Use the example file** - it has many pre-made configs

4. **Reset to default** - just delete the config:
   ```bash
   rm ~/.config/tera/appearance_config.yaml
   ```

## 🐛 Troubleshooting

**Config not working?**
- Check YAML syntax
- Restart TERA (config loads at startup)
- See [HEADER-CUSTOMIZATION-README.md](HEADER-CUSTOMIZATION-README.md) troubleshooting section

**Want to reset?**
```bash
rm ~/.config/tera/appearance_config.yaml
```

## 📞 Support

1. Read [HEADER-CUSTOMIZATION-README.md](HEADER-CUSTOMIZATION-README.md)
2. Check [appearance_config.example.yaml](appearance_config.example.yaml)
3. Run [demo-headers.sh](demo-headers.sh)
4. See [PHASE1-COMPLETE.md](PHASE1-COMPLETE.md) testing guide

## 🎉 Status

**Phase 1: COMPLETE ✅**

All features working and ready to use!

---

**Version:** 1.0  
**Status:** Production Ready  
**Last Updated:** February 2026
