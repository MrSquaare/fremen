package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/MrSquaare/fremen/internal/database"
)

type npmLockfile struct {
	Dependencies map[string]npmPkg `json:"dependencies"`
	Packages     map[string]npmPkg `json:"packages"`
}

type npmPkg struct {
	Version string `json:"version"`
}

// Parses npm lockfiles (both legacy and v2/v3 formats) and returns all packages flagged as vulnerable by the DB.
func parseNpmLockfile(path string, db *database.VulnerabilityDatabase) ([]Vulnerability, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var lf npmLockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}

	var vulns []Vulnerability

	// Legacy `dependencies`
	for name, pkg := range lf.Dependencies {
		if isVulnerable(db, name, pkg.Version) {
			vulns = append(vulns, Vulnerability{
				PackageName: name,
				Version:     pkg.Version,
			})
		}
	}

	// Modern npm v2/v3 `packages` section
	for pkgPath, pkg := range lf.Packages {
		if pkgPath == "" {
			continue
		}

		name := npmPackageName(pkgPath)
		if isVulnerable(db, name, pkg.Version) {
			vulns = append(vulns, Vulnerability{
				PackageName: name,
				Version:     pkg.Version,
			})
		}
	}

	return vulns, nil
}

func npmPackageName(path string) string {
	path = filepath.ToSlash(path)

	const nm = "node_modules/"
	if idx := strings.LastIndex(path, nm); idx != -1 {
		return path[idx+len(nm):]
	}
	return path
}

func isVulnerable(db *database.VulnerabilityDatabase, name, version string) bool {
	if name == "" || version == "" {
		return false
	}
	return db.IsInfected(name, version)
}
