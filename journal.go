package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a journal entry
type Entry struct {
	Content string    `json:"content"`
	Date    time.Time `json:"date"`
	Path    string    `json:"path,omitempty"`
}

// Item represents a journal entry in the list view
type Item struct {
	TitleText string
	Desc      string
	Path      string
	Modified  time.Time
}

// FilterValue implements list.Item interface
func (i Item) FilterValue() string {
	return i.TitleText
}

// Title implements list.Item interface
func (i Item) Title() string {
	return i.TitleText
}

// Description implements list.Item interface
func (i Item) Description() string {
	return i.Desc
}

// NewEntry creates a new journal entry
func NewEntry(content string) *Entry {
	return &Entry{
		Content: content,
		Date:    time.Now(),
	}
}

// GetEntriesDir returns the directory where entries are stored
func GetEntriesDir(config *Config) string {
	return config.EntriesDir
}

// SaveEntry saves a journal entry to a file
func SaveEntry(entry *Entry, config *Config) error {
	var path string
	
	// If the entry already has a path, use it (for updates)
	// Otherwise, create a new path based on the timestamp (for new entries)
	if entry.Path != "" {
		path = entry.Path
	} else {
		filename := fmt.Sprintf("%d.json", entry.Date.UnixNano())
		path = filepath.Join(config.EntriesDir, filename)
		entry.Path = path
	}
	
	// Marshal the entry to JSON
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	
	// Write the JSON to the file
	return os.WriteFile(path, data, 0644)
}

// LoadEntry loads a journal entry from a file
func LoadEntry(path string) (*Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	
	// Set the path in the entry
	entry.Path = path
	
	return &entry, nil
}

// LoadEntries loads all journal entries
func LoadEntries(config *Config) ([]*Entry, error) {
	dir := config.EntriesDir
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Entry{}, nil
		}
		return nil, err
	}
	
	var entries []*Entry
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		
		path := filepath.Join(dir, file.Name())
		entry, err := LoadEntry(path)
		if err != nil {
			continue
		}
		
		entries = append(entries, entry)
	}
	
	// Sort entries by date (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})
	
	return entries, nil
}

// DeleteEntry deletes a journal entry
func DeleteEntry(path string) error {
	return os.Remove(path)
}
