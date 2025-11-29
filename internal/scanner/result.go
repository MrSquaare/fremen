package scanner

import "github.com/MrSquaare/fremen/internal/parser"

type ScanResult struct {
	ProjectPath      string                 `json:"project"`
	Lockfiles        []string               `json:"lockfiles"`
	InfectedPackages []parser.Vulnerability `json:"infected_packages"`
}

func (sr ScanResult) InfectedCount() int {
	return len(sr.InfectedPackages)
}

func (sr ScanResult) ToMap() map[string]any {
	pkgs := make([]map[string]string, 0, len(sr.InfectedPackages))
	for _, v := range sr.InfectedPackages {
		pkgs = append(pkgs, map[string]string{
			"name":    v.PackageName,
			"version": v.Version,
		})
	}

	return map[string]any{
		"project":           sr.ProjectPath,
		"lockfiles":         sr.Lockfiles,
		"infected_count":    sr.InfectedCount(),
		"infected_packages": pkgs,
	}
}
