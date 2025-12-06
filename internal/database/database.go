package database

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type VulnerabilityDatabase struct {
	entries    map[string]map[string]struct{}
	loadedPath string
	entryCount int
}

func (db *VulnerabilityDatabase) EntryCount() int {
	return db.entryCount
}

func (db *VulnerabilityDatabase) LoadedPath() string {
	return db.loadedPath
}

func New() *VulnerabilityDatabase {
	return &VulnerabilityDatabase{
		entries: make(map[string]map[string]struct{}),
	}
}

// Load reads the vulnerability database file from disk.
func (db *VulnerabilityDatabase) Load(path string) error {
	if path == "" {
		exe, err := os.Executable()
		if err != nil {
			path = "database.txt"
		} else {
			path = filepath.Join(filepath.Dir(exe), "database.txt")
		}
	}

	// If already loaded with the same path, skip.
	if db.loadedPath == path {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("database not found at %q; please provide a valid path using --database or ensure database.txt exists", path)
		}
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = file.Close() }()

	newEntries := make(map[string]map[string]struct{})
	count := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, version, ok := parseEntry(line)
		if !ok {
			continue
		}

		versions := newEntries[name]
		if versions == nil {
			versions = make(map[string]struct{})
			newEntries[name] = versions
		}

		if _, exists := versions[version]; !exists {
			versions[version] = struct{}{}
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error while reading database: %w", err)
	}

	db.entries = newEntries
	db.loadedPath = path
	db.entryCount = count

	return nil
}

func (db *VulnerabilityDatabase) IsInfected(name, version string) bool {
	if name == "" || version == "" {
		return false
	}

	versions := db.entries[name]
	if versions == nil {
		return false
	}

	_, found := versions[version]
	return found
}

// Parses a line of "name:version" and returns name, version, ok.
func parseEntry(line string) (name, version string, ok bool) {
	idx := strings.IndexByte(line, ':')
	if idx <= 0 || idx >= len(line)-1 {
		return "", "", false
	}

	name = strings.TrimSpace(line[:idx])
	version = strings.TrimSpace(line[idx+1:])

	if name == "" || version == "" {
		return "", "", false
	}

	return name, version, true
}
