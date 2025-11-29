package parser

import (
	"os"
	"regexp"

	"github.com/MrSquaare/fremen/internal/database"
)

var yarnPattern = regexp.MustCompile(
	`(?m)^['"]?(@?[^@"\s']+)@.+?['"]?:\s*(?:\r?\n|\r)\s*version(?:\s+|:\s+)["']?([^"\s']+)["']?`,
)

// Parses a Yarn v1 lockfile and returns all packages flagged as vulnerable by the DB.
func parseYarnLockfile(path string, db *database.VulnerabilityDatabase) ([]Vulnerability, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var issues []Vulnerability

	matches := yarnPattern.FindAllSubmatch(data, -1)
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}

		pkg := string(m[1])
		ver := string(m[2])

		if db.IsInfected(pkg, ver) {
			issues = append(issues, Vulnerability{
				PackageName: pkg,
				Version:     ver,
			})
		}
	}

	return issues, nil
}
