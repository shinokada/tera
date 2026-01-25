package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a selectable menu item
type MenuItem struct {
	title    string
	desc     string
	shortcut string
}

func NewMenuItem(title, desc, shortcut string) MenuItem {
	return MenuItem{
		title:    title,
		desc:     desc,
		shortcut: shortcut,
	}
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Description() string { return i.desc }
func (i MenuItem) FilterValue() string { return i.title }
func (i MenuItem) Shortcut() string    { return i.shortcut }

// MenuDelegate is a custom delegate for menu items
type MenuDelegate struct {
	list.DefaultDelegate
}

func NewMenuDelegate() MenuDelegate {
	d := list.NewDefaultDelegate()
	d.SetHeight(1)  // Single line per item, no spacing
	d.SetSpacing(0) // No spacing between items
	return MenuDelegate{DefaultDelegate: d}
}

var (
	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	shortcutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

func (d MenuDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	menuItem, ok := item.(MenuItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s. %s", menuItem.Shortcut(), menuItem.Title())

	fn := normalItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + str)
		}
	} else {
		fn = func(s ...string) string {
			return normalItemStyle.Render("  " + str)
		}
	}

	fmt.Fprint(w, fn())
}

// CreateMenu creates a new menu list with the given items
func CreateMenu(items []MenuItem, title string, width, height int) list.Model {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	delegate := NewMenuDelegate()
	// Set height to accommodate all items without pagination
	// Each item takes 1 line (we removed the spacing)
	itemHeight := len(items)
	if height < itemHeight {
		height = itemHeight + 4 // Add space for title
	}
	
	l := list.New(listItems, delegate, width, height)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false) // Disable pagination dots
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		PaddingLeft(2)

	return l
}

// HandleMenuKey handles standard menu navigation keys
// Returns the selected index if Enter is pressed, -1 otherwise
func HandleMenuKey(msg tea.KeyMsg, m list.Model) (list.Model, int) {
	switch msg.String() {
	case "up", "k":
		m.CursorUp()
		return m, -1
	case "down", "j":
		m.CursorDown()
		return m, -1
	case "g", "home":
		m.Select(0)
		return m, -1
	case "G", "end":
		m.Select(len(m.Items()) - 1)
		return m, -1
	case "enter":
		return m, m.Index()
	}

	// Check for number shortcuts
	for i, item := range m.Items() {
		if menuItem, ok := item.(MenuItem); ok {
			if msg.String() == menuItem.Shortcut() {
				m.Select(i)
				return m, i
			}
		}
	}

	return m, -1
}
