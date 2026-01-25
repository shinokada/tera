#!/bin/bash
# Test menu display fixes

echo "🔧 Building with menu fixes..."
make clean
make build

if [ ! -f "./tera" ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""
echo "📋 Menu Display Fixes:"
echo "  1. ✅ Main menu - single spacing, all items visible"
echo "  2. ✅ Search menu - single spacing, all 6 options visible"
echo "  3. ✅ No pagination dots (•••) on menus"
echo ""
echo "🧪 Manual Tests:"
echo ""
echo "Test 1 - Main Menu:"
echo "  • Run: ./tera"
echo "  • Verify: See all 7 menu items without double spacing"
echo "  • Verify: No ••• at bottom"
echo "  Expected items:"
echo "    1. Play from Favorites"
echo "    2. Search Stations"
echo "    3. Manage Lists"
echo "    4. I Feel Lucky"
echo "    5. Delete Station"
echo "    6. Gist Management"
echo "    0. Exit"
echo ""
echo "Test 2 - Search Menu:"
echo "  • From main menu, press 2"
echo "  • Verify: See all 6 search options"
echo "  • Verify: No ••• at bottom"
echo "  Expected items:"
echo "    1. Search by Tag"
echo "    2. Search by Name"
echo "    3. Search by Language"
echo "    4. Search by Country Code"
echo "    5. Search by State"
echo "    6. Advanced Search"
echo ""
echo "Test 3 - No Double Spacing:"
echo "  • Check both menus"
echo "  • Verify: Single line spacing between items"
echo "  • Verify: Compact, readable layout"
echo ""
echo "🚀 Ready to test! Run: ./tera"
