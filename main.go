package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state
type Model struct {
	state        state
	list         list.Model
	textInput    textinput.Model
	currentEntry *Entry
	err          error
	width        int
	height       int
	config       *Config
}

// State represents the current view of the application
type state int

const (
	listView state = iota
	createView
	viewEntryView
	editEntryView
)

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadEntriesCmd(m.config),
		tea.EnterAltScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(m.width-4, m.height-10)
		
		return m, nil
		
	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch m.state {
		case listView:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "n":
				m.state = createView
				m.textInput.Focus()
				return m, nil
			case "enter":
				if len(m.list.Items()) > 0 {
					i, ok := m.list.SelectedItem().(Item)
					if ok {
						entry, err := LoadEntry(i.Path)
						if err != nil {
							m.err = err
							return m, nil
						}
						m.currentEntry = entry
						m.state = viewEntryView
						return m, nil
					}
				}
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case createView:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = listView
				m.textInput.Reset()
				return m, nil
			case "enter":
				if m.textInput.Value() != "" {
					entry := NewEntry(m.textInput.Value())
					err := SaveEntry(entry, m.config)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.textInput.Reset()
					m.state = listView
					return m, loadEntriesCmd(m.config)
				}
			}

			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd

		case viewEntryView:
			switch msg.String() {
			case "esc", "q":
				m.state = listView
				m.currentEntry = nil
				return m, nil
			case "d":
				if m.currentEntry != nil {
					err := DeleteEntry(m.currentEntry.Path)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.state = listView
					m.currentEntry = nil
					return m, loadEntriesCmd(m.config)
				}
			case "e":
				if m.currentEntry != nil {
					m.textInput.SetValue(m.currentEntry.Content)
					m.textInput.Focus()
					m.state = editEntryView
					return m, nil
				}
			}
			
		case editEntryView:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = viewEntryView
				m.textInput.Reset()
				return m, nil
			case "enter":
				if m.currentEntry != nil && m.textInput.Value() != "" {
					m.currentEntry.Content = m.textInput.Value()
					
					err := SaveEntry(m.currentEntry, m.config)
					if err != nil {
						m.err = err
						return m, nil
					}
					
					m.textInput.Reset()
					m.state = viewEntryView
					return m, loadEntriesCmd(m.config)
				}
			}
			
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

	case entriesLoadedMsg:
		m.list.SetItems(msg)
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress any key to continue", m.err)
	}

	listStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 2)

	createStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 2)

	viewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 2)

	switch m.state {
	case listView:
		itemCount := len(m.list.Items())
		itemsInfo := fmt.Sprintf("You have %d journal entries", itemCount)
		if itemCount == 0 {
			itemsInfo = "No journal entries yet. Press 'n' to create one."
		}
		
		return listStyle.Render(
			fmt.Sprintf(
				"%s\n\n%s\n%s\n\n%s",
				titleStyle.Render("Journal"),
				itemsInfo,
				m.list.View(),
				helpStyle.Render("n: new entry • enter: view entry • q: quit"),
			),
		)

	case createView:
		m.textInput.Width = m.width - 10
		
		return createStyle.Render(
			fmt.Sprintf(
				"%s\n\n%s\n\n%s",
				titleStyle.Render("New Journal Entry"),
				m.textInput.View(),
				helpStyle.Render("enter: save • esc: cancel"),
			),
		)

	case viewEntryView:
		if m.currentEntry == nil {
			return "Error: No entry selected"
		}
		return viewStyle.Render(
			fmt.Sprintf(
				"%s\n\n%s\n\n%s\n\n%s",
				titleStyle.Render("Journal Entry"),
				dateStyle.Render(m.currentEntry.Date.Format("January 1, 2006 15:04:05")),
				m.currentEntry.Content,
				helpStyle.Render("e: edit • d: delete • esc: back"),
			),
		)
		
	case editEntryView:
		if m.currentEntry == nil {
			return "Error: No entry selected"
		}
		
		m.textInput.Width = m.width - 10
		
		return viewStyle.Render(
			fmt.Sprintf(
				"%s\n\n%s\n\n%s\n\n%s",
				titleStyle.Render("Edit Journal Entry"),
				dateStyle.Render(m.currentEntry.Date.Format("January 2, 2006 15:04:05")),
				m.textInput.View(),
				helpStyle.Render("enter: save • esc: cancel"),
			),
		)

	default:
		return "Unknown state"
	}
}

// Styles for elements that don't need to adapt to window size
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type entriesLoadedMsg []list.Item
type errMsg error

func loadEntriesCmd(config *Config) tea.Cmd {
	return func() tea.Msg {
		entries, err := LoadEntries(config)
		if err != nil {
			return errMsg(err)
		}

		items := make([]list.Item, len(entries))
		for i, entry := range entries {
			items[i] = Item{
				TitleText: entry.Date.Format("January 2, 2006 15:04:05"),
				Desc:      truncate(entry.Content, 50),
				Path:      entry.Path,
				Modified:  entry.Date,
			}
		}

		return entriesLoadedMsg(items)
	}
}

// Helper functions
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create default config file if it doesn't exist
	if err := CreateDefaultConfigFile(); err != nil {
		fmt.Printf("Warning: Could not create default config file: %v\n", err)
	}

	// Ensure entries directory exists
	if err := config.EnsureEntriesDir(); err != nil {
		fmt.Printf("Error creating entries directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize text input for creating entries
	ti := textinput.New()
	ti.Placeholder = "Write your journal entry here..."
	ti.CharLimit = 0
	ti.Width = 50

	// Initialize list for viewing entries
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7D56F4"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("#DADADA")).Background(lipgloss.Color("#7D56F4"))
	
	// Default dimensions - will be updated when we get a WindowSizeMsg
	defaultWidth, defaultHeight := 80, 24
	
	l := list.New([]list.Item{}, delegate, defaultWidth-4, defaultHeight-10)
	l.SetShowTitle(false)  // We'll handle the title separately
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	// Initialize model
	m := Model{
		state:     listView,
		list:      l,
		textInput: ti,
		width:     defaultWidth,
		height:    defaultHeight,
		config:    config,
	}

	// Display configuration info
	fmt.Printf("Journal entries directory: %s\n", config.EntriesDir)
	if config.DevMode {
		fmt.Println("Running in development mode")
	} else {
		fmt.Println("Running in production mode")
	}

	// Load entries
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	// Start the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
