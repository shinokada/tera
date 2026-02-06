# Block Rules Quick Reference

## ⌨️ Keyboard Shortcuts

### View Blocked Stations
- `↑↓` or `j/k` - Navigate list
- `u` - Unblock selected station
- `c` - Clear all (with confirmation)
- `Esc` - Back to menu

### View Active Rules  
- `↑↓` or `j/k` - Navigate rules
- `d` - Delete selected rule (with confirmation)
- `Esc` - Back to menu

### Add Rule (Country/Language/Tag)
- Type value
- `Enter` - Show confirmation
- `Esc` - Cancel

### Confirmations
- `y` - Yes, proceed
- `n` or `Esc` - No, cancel

## 📋 Quick Tasks

### Add a Blocking Rule
```
Menu → Block List → Manage Block Rules
→ Block by [Country/Language/Tag]
→ Enter value → Enter
→ Confirm → y
```

### Delete a Rule
```
Menu → Block List → Manage Block Rules
→ View Active Rules
→ Navigate to rule → d
→ Confirm → y
```

### Export Blocklist
```
Menu → Block List → Import/Export
→ Export
→ Enter filename (or Enter for auto)
→ File saved to ~/.tera/exports/
```

### Import Blocklist
```
Menu → Block List → Import/Export
→ Import
→ Enter file path
→ Choose m (merge) or r (replace)
```

## 💡 Pro Tips

1. **Before deleting many stations individually**, create a rule instead
2. **Export regularly** to backup your blocklist  
3. **Use merge mode** when importing to keep existing rules
4. **Check confirmation dialogs** - they show exactly what will happen
5. **Cancel anytime** with Esc if you change your mind

## 🎯 Common Scenarios

**Block all French stations:**
```
Block by Language → "french" → Enter → y
```

**Block all news stations:**
```
Block by Tag → "news" → Enter → y
```

**Block stations from USA:**
```
Block by Country → "US" → Enter → y
(or "United States")
```

**Remove a rule:**
```
View Active Rules → Navigate to rule → d → y
```

**Backup before big changes:**
```
Export → Enter → ✓
(Now make your changes)
```

**Share blocklist with friend:**
```
Export → "my-blocklist" → ✓
Send ~/.tera/exports/my-blocklist-*.json
Friend: Import → path → m
```

## ⚠️ Important Notes

- **Deletions are permanent** - export first if unsure
- **Rules affect all matching stations** - be specific
- **Case-insensitive matching** - "English" = "english"
- **Import validates JSON** - corrupted files will be rejected
- **Merge preserves data** - replace clears first

## 🎨 Visual Indicators

| Icon | Meaning |
|------|---------|
| ✓ | Success |
| ✗ | Error |
| ℹ | Information |
| ⚠ | Warning |
| 🚫 | Blocked |
| 📋 | Menu/List |

## 🔄 State Flow

```
Main Menu
  ↓
Block List Management
  ↓
Manage Block Rules ──→ View Active Rules ──→ Delete? ──→ Confirm ──→ ✓
  ↓                                          ↓
Block by Country ──→ Enter value ──→ Confirm ──→ ✓
  ↓                                   ↓
Block by Language                   Cancel → Back
  ↓
Block by Tag
```

## 📞 Help

If you see an error:
1. Check the error message (red ✗)
2. Common issues:
   - Empty value → Enter something
   - Duplicate rule → Rule already exists
   - File not found → Check path
   - Invalid JSON → Use exported format

All operations have confirmations and clear feedback!
