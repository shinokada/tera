#!/bin/bash
# Quick demo of TERA header customization

echo "TERA Header Customization - Quick Demo"
echo "======================================="
echo ""

CONFIG_DIR="$HOME/.config/tera"
CONFIG_FILE="$CONFIG_DIR/appearance_config.yaml"
BACKUP_FILE="$CONFIG_DIR/appearance_config.yaml.backup"

# Backup existing config if present
if [ -f "$CONFIG_FILE" ]; then
    echo "Backing up existing config to $BACKUP_FILE"
    cp "$CONFIG_FILE" "$BACKUP_FILE"
fi

# Create config directory
mkdir -p "$CONFIG_DIR"

echo "Demo 1: Default TERA header"
echo "---------------------------"
cat > "$CONFIG_FILE" << 'EOF'
appearance:
  header:
    mode: "default"
EOF
echo "Created config with default mode"
echo "Run 'go run .' to see the default TERA header"
echo ""
read -p "Press Enter to continue to next demo..."

echo ""
echo "Demo 2: Custom text header"
echo "--------------------------"
cat > "$CONFIG_FILE" << 'EOF'
appearance:
  header:
    mode: "text"
    custom_text: "🎵 My Personal Radio 🎵"
    alignment: "center"
    width: 50
    color: "auto"
    bold: true
    padding_top: 1
EOF
echo "Created config with custom text"
echo "Run 'go run .' to see custom text instead of TERA"
echo ""
read -p "Press Enter to continue to next demo..."

echo ""
echo "Demo 3: ASCII art header (simple)"
echo "----------------------------------"
cat > "$CONFIG_FILE" << 'EOF'
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
EOF
echo "Created config with ASCII art box"
echo "Run 'go run .' to see ASCII art header"
echo ""
read -p "Press Enter to continue to next demo..."

echo ""
echo "Demo 4: ASCII art header (fancy)"
echo "--------------------------------"
cat > "$CONFIG_FILE" << 'EOF'
appearance:
  header:
    mode: "ascii"
    ascii_art: |
       ____      _    ____ ___ ___  
      |  _ \    / \  |  _ \_ _/ _ \ 
      | |_) |  / _ \ | | | | | | | |
      |  _ <  / ___ \| |_| | | |_| |
      |_| \_\/_/   \_\____/___\___/ 
    alignment: "center"
    width: 60
    color: "auto"
    padding_top: 0
    padding_bottom: 1
EOF
echo "Created config with fancy ASCII art"
echo "Run 'go run .' to see large ASCII art header"
echo ""
read -p "Press Enter to continue to next demo..."

echo ""
echo "Demo 5: No header (maximum screen space)"
echo "-----------------------------------------"
cat > "$CONFIG_FILE" << 'EOF'
appearance:
  header:
    mode: "none"
EOF
echo "Created config with no header"
echo "Run 'go run .' to see more screen space without header"
echo ""
read -p "Press Enter to finish..."

echo ""
echo "Demo complete!"
echo ""
echo "Your config is now set to 'none' mode (no header)."
echo ""
if [ -f "$BACKUP_FILE" ]; then
    echo "To restore your original config:"
    echo "  cp $BACKUP_FILE $CONFIG_FILE"
    echo ""
fi
echo "To reset to default TERA header:"
echo "  rm $CONFIG_FILE"
echo ""
echo "To create your own custom header:"
echo "  1. Edit $CONFIG_FILE"
echo "  2. See appearance_config.example.yaml for examples"
echo "  3. Use https://patorjk.com/software/taag/ to create ASCII art"
echo ""
echo "Enjoy your customized TERA!"
