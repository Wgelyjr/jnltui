package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	state        state
	list         list.Model
	textarea     textarea.Model
	currentEntry *Entry
	err          error
	width        int
	height       int
	config       *Config
}

type state int

const (
	listView state = iota
	createView
	viewEntryView
	editEntryView
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadEntriesCmd(m.config),
		tea.EnterAltScreen,
		textarea.Blink,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(m.width-4, m.height-10)
		
		// Update textarea size
		m.textarea.SetWidth(m.width - 20)
		m.textarea.SetHeight(m.height - 15)
		
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
				m.textarea.Focus()
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
				m.textarea.Reset()
				return m, nil
			case "enter":
				if m.textarea.Value() != "" {
					entry := NewEntry(m.textarea.Value())
					err := SaveEntry(entry, m.config)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.textarea.Reset()
					m.state = listView
					return m, loadEntriesCmd(m.config)
				}
			}

			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

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
					m.textarea.SetValue(m.currentEntry.Content)
					m.textarea.Focus()
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
				m.textarea.Reset()
				return m, nil
			case "enter":
				if m.currentEntry != nil && m.textarea.Value() != "" {
					m.currentEntry.Content = m.textarea.Value()
					
					err := SaveEntry(m.currentEntry, m.config)
					if err != nil {
						m.err = err
						return m, nil
					}
					
					m.textarea.Reset()
					m.state = viewEntryView
					return m, loadEntriesCmd(m.config)
				}
			}
			
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
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
		Width(m.width - 2).
		Height(m.height - 2)

	createStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 2).
		Height(m.height - 2)

	viewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 2).
		Height(m.height - 2)

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
		return createStyle.Render(
			fmt.Sprintf(
				"%s\n\n%s\n\n%s",
				titleStyle.Render("New Journal Entry"),
				m.textarea.View(),
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
				dateStyle.Render(m.currentEntry.Date.Format("January 2, 2006 15:04:05")),
				m.currentEntry.Content,
				helpStyle.Render("e: edit • d: delete • esc: back"),
			),
		)
		
	case editEntryView:
		if m.currentEntry == nil {
			return "Error: No entry selected"
		}
		
		return viewStyle.Render(
			fmt.Sprintf(
				"%s\n\n%s\n\n%s\n\n%s",
				titleStyle.Render("Edit Journal Entry"),
				dateStyle.Render(m.currentEntry.Date.Format("January 2, 2006 15:04:05")),
				m.textarea.View(),
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

	// Default dimensions - will be updated when we get a WindowSizeMsg
	defaultWidth, defaultHeight := 80, 24

	// Initialize textarea for creating entries
	ta := textarea.New()
	ta.Placeholder = "Write your journal entry here..."
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.CharLimit = 0
	
	// Configure textarea to wrap text properly
	ta.SetWidth(defaultWidth)
	ta.SetHeight(defaultHeight - 15)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7D56F4"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("#DADADA")).Background(lipgloss.Color("#7D56F4"))
	
	l := list.New([]list.Item{}, delegate, defaultWidth-4, defaultHeight-10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	m := Model{
		state:     listView,
		list:      l,
		textarea:  ta,
		width:     defaultWidth,
		height:    defaultHeight,
		config:    config,
	}

	// Display configuration info - for debug stuff really
	fmt.Printf("Journal entries directory: %s\n", config.EntriesDir)
	if config.DevMode {
		fmt.Println("Running in development mode")
	} else {
		fmt.Println("Running in production mode")
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
