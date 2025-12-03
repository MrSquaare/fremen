package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/MrSquaare/fremen/internal/database"
)

var (
	pnpmIgnoredKeys = map[string]struct{}{
		"resolution":                 {},
		"engines":                    {},
		"os":                         {},
		"cpu":                        {},
		"peerDependencies":           {},
		"dependencies":               {},
		"optionalDependencies":       {},
		"devDependencies":            {},
		"transitivePeerDependencies": {},
		"dev":                        {},
		"hasBin":                     {},
		"requiresBuild":              {},
		"name":                       {},
		"version":                    {},
		"lockfileVersion":            {},
		"settings":                   {},
		"importers":                  {},
		"packages":                   {},
		"specifiers":                 {},
		"patchedDependencies":        {},
	}

	pnpmKeyPattern = regexp.MustCompile(`^\s+['"]?/?([^:'"\s]+)['"]?:`)
)

// Parse pnpm lockfile for vulnerable packages and and returns all packages flagged as vulnerable by the DB.
func parsePnpmLockfile(path string, db *database.VulnerabilityDatabase) ([]Vulnerability, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var issues []Vulnerability
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := sc.Text()

		m := pnpmKeyPattern.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		key := m[1]

		if _, skip := pnpmIgnoredKeys[key]; skip {
			continue
		}

		name, version, ok := parsePnpmKey(key)
		if !ok {
			continue
		}

		if db.IsInfected(name, version) {
			issues = append(issues, Vulnerability{
				PackageName: name,
				Version:     version,
			})
		}
	}

	if err := sc.Err(); err != nil {
		return issues, err
	}
	return issues, nil
}

func parsePnpmKey(key string) (name, version string, ok bool) {
	if idx := strings.Index(key, "("); idx != -1 {
		key = key[:idx]
	}

	sep := strings.LastIndexAny(key, "/@")
	if sep <= 0 || sep >= len(key)-1 {
		return "", "", false
	}

	name = key[:sep]
	version = key[sep+1:]

	if u := strings.Index(version, "_"); u != -1 {
		version = version[:u]
	}

	return name, version, name != "" && version != ""
}
