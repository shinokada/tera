#!/bin/bash

# Phase 2 Implementation Helper Script
# This script helps implement the Appearance Settings UI

set -e

TERA_DIR="/Users/shinichiokada/Terminal-Tools/tera"
cd "$TERA_DIR"

echo "========================================"
echo "TERA Phase 2: Appearance Settings UI"
echo "========================================"
echo ""

# Check if running from correct directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Not in TERA project root"
    echo "Please run from: $TERA_DIR"
    exit 1
fi

echo "Step 1: Copying appearance_settings.go..."
if [ -f "/home/claude/appearance_settings.go" ]; then
    cp /home/claude/appearance_settings.go internal/ui/appearance_settings.go
    echo "✅ appearance_settings.go copied"
else
    echo "⚠️  Warning: /home/claude/appearance_settings.go not found"
    echo "   Please copy it manually to internal/ui/appearance_settings.go"
fi

echo ""
echo "Step 2: Adding help function..."
if [ -f "help_patch.go" ]; then
    echo ""
    echo "Add this function to internal/ui/components/help.go:"
    echo "---"
    cat help_patch.go
    echo "---"
    echo ""
    read -p "Press Enter after you've added the function to help.go..."
else
    echo "⚠️  help_patch.go not found"
fi

echo ""
echo "Step 3: Checking modifications needed..."
echo ""
echo "Files to modify:"
echo "1. internal/ui/app.go - Add screen constant, field, handlers"
echo "2. internal/ui/settings.go - Add menu item and handler"
echo ""
echo "See PHASE2-IMPLEMENTATION.md for detailed instructions"
echo ""

read -p "Have you made all the modifications? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "Please complete the modifications and run this script again."
    echo "See: PHASE2-IMPLEMENTATION.md for step-by-step instructions"
    exit 0
fi

echo ""
echo "Step 4: Building TERA..."
if go build; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed. Please check errors above."
    exit 1
fi

echo ""
echo "Step 5: Testing configuration..."
CONFIG_FILE="$HOME/.config/tera/appearance_config.yaml"
if [ -f "$CONFIG_FILE" ]; then
    echo "✅ Found existing config at: $CONFIG_FILE"
    echo ""
    echo "Current configuration:"
    echo "---"
    cat "$CONFIG_FILE"
    echo "---"
else
    echo "ℹ️  No config file found at: $CONFIG_FILE"
    echo "   This is normal - you can create one through the UI"
fi

echo ""
echo "========================================"
echo "✅ Phase 2 Implementation Complete!"
echo "========================================"
echo ""
echo "Next steps:"
echo "1. Run: ./tera"
echo "2. Press 6 for Settings"
echo "3. Select 'Appearance'"
echo "4. Configure your header"
echo "5. Press 'Save'"
echo "6. Go back to see your changes!"
echo ""
echo "Your existing config file should now work correctly!"
echo ""
