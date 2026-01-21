# List Menu User Guide

## Quick Navigation Reference

When working with lists in TERA, you now have easy ways to navigate back:

### Navigation Commands

| What to Type             | What Happens                         |
| ------------------------ | ------------------------------------ |
| `0`                      | Go back to List Menu                 |
| `00`                     | Go to Main Menu                      |
| `back`                   | Go back to List Menu (alternative)   |
| `main`                   | Go to Main Menu (alternative)        |
| Just press Enter (empty) | Shows error, stays in current screen |

### Where These Work

✅ **Create a list** - Type `0` when asked for list name  
✅ **Delete a list** - Type `0` when asked which list to delete  
✅ **Edit a list** - Type `0` when selecting list OR when entering new name  

## Examples

### Example 1: Canceling List Creation

```text
TERA - Create New List

My lists: 
jazz
rock

Type '0' to go back, '00' for main menu
Type a new list name: 0

[Returns to List Menu]
```

### Example 2: Going to Main Menu from Delete

```text
TERA - Delete List

My lists: 
jazz
rock

Type '0' to go back, '00' for main menu
Type a list name to delete: 00

[Returns to Main Menu]
```

### Example 3: Canceling Edit Operation

```text
TERA - Edit List Name

My lists: 
jazz
rock

Type '0' to go back, '00' for main menu
Type a list name to edit: jazz
Old name: jazz

Type '0' to go back, '00' for main menu
Type a new name: 0

[Returns to List Menu]
```

## Protected Operations

Some operations are protected for your safety:

### ❌ Cannot Delete My-favorites
```text
Type a list name to delete: My-favorites
Cannot delete My-favorites list!
```

### ❌ Cannot Rename My-favorites
```text
Type a list name to edit: My-favorites
Cannot rename My-favorites list!
```

### ❌ Cannot Create Duplicate Lists
```text
Type a new list name: jazz
List 'jazz' already exists!
```

### ❌ Cannot Use Empty Names
```text
Type a new list name: 
List name cannot be empty.
```

## Tips

1. **Use `0` for quick back** - Fastest way to return to List Menu
2. **Use `00` to escape entirely** - Jump straight to Main Menu
3. **Remember the yellow hints** - They remind you of your navigation options
4. **ESC in fzf menus** - Also works to go back when selecting from menus

## Common Workflows

### Creating Multiple Lists
```text
Main Menu → List Menu → Create → Type name → (created)
                ↑                                    |
                └────────────────────────────────────┘
Automatically returns to List Menu, ready to create another
```

### Quick Cancel
```text
Main Menu → List Menu → Create → Type '00' → Back to Main Menu
```

### Edit with Cancel
```text
List Menu → Edit → Type list name → Type '0' → Back to List Menu
                                              (no changes made)
```

## Keyboard Shortcuts Summary

- **In fzf menus**: Arrow keys to navigate, Enter to select, ESC to cancel
- **In text prompts**: 
  - `0` or `back` = Previous menu
  - `00` or `main` = Main menu
  - Enter (empty) = Error message, stay on current screen

Happy list management! 📝
