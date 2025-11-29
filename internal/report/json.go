package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MrSquaare/fremen/internal/scanner"
)

type JSONConfig struct {
	Paths              []string `json:"paths"`
	Database           string   `json:"database"`
	Recursive          bool     `json:"recursive"`
	IncludeGit         bool     `json:"include_git"`
	IncludeNodeModules bool     `json:"include_node_modules"`
	ExcludeRegex       string   `json:"exclude_regex"`
	FullReport         bool     `json:"full_report"`
}

func PrintJSONReport(results []scanner.ScanResult, cfg scanner.ScanConfig, showFull bool) {
	dbPath := cfg.DatabasePath
	if dbPath == "" {
		dbPath = "Default"
	}
	exclude := ""
	if cfg.ExcludeRegex != nil {
		exclude = cfg.ExcludeRegex.String()
	}

	jsonCfg := JSONConfig{
		Paths:              cfg.TargetPaths,
		Database:           dbPath,
		Recursive:          cfg.Recursive,
		IncludeGit:         cfg.IncludeGit,
		IncludeNodeModules: cfg.IncludeNodeModules,
		ExcludeRegex:       exclude,
		FullReport:         showFull,
	}

	displayResults, summary := summarizeScanResults(results, jsonCfg.FullReport)

	outResults := make([]map[string]any, 0, len(displayResults))
	for _, r := range displayResults {
		outResults = append(outResults, r.ToMap())
	}

	out := map[string]any{
		"configuration": jsonCfg,
		"results":       outResults,
		"summary":       summary,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, "Error encoding JSON:", err)
		os.Exit(1)
	}

	if summary.TotalProjects == 0 {
		os.Exit(1)
	} else if summary.InfectedProjects > 0 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
