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

type Item struct {
	TitleText string
	Desc      string
	Path      string
	Modified  time.Time
}

func (i Item) FilterValue() string {
	return i.TitleText
}

func (i Item) Title() string {
	return i.TitleText
}

func (i Item) Description() string {
	return i.Desc
}

func NewEntry(content string) *Entry {
	return &Entry{
		Content: content,
		Date:    time.Now(),
	}
}

func GetEntriesDir(config *Config) string {
	return config.EntriesDir
}

func SaveEntry(entry *Entry, config *Config) error {
	var path string
	if entry.Path != "" {
		path = entry.Path
	} else {
		filename := fmt.Sprintf("%d.json", entry.Date.UnixNano())
		path = filepath.Join(config.EntriesDir, filename)
		entry.Path = path
	}
	
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

func LoadEntry(path string) (*Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	
	entry.Path = path
	
	return &entry, nil
}

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
	
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})
	
	return entries, nil
}

func DeleteEntry(path string) error {
	return os.Remove(path)
}
