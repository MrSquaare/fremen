package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/MrSquaare/fremen/internal/scanner"
	"github.com/MrSquaare/fremen/internal/style"
	"github.com/fatih/color"
)

type DisplayConfig struct {
	Paths              []string
	Database           string
	Recursive          bool
	IncludeGit         bool
	IncludeNodeModules bool
	ExcludeRegex       string
	FullReport         bool
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func listOrDash(list []string) string {
	if len(list) == 0 {
		return "-"
	}
	return strings.Join(list, ", ")
}

func stringOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func PrintCLIReport(results []scanner.ScanResult, cfg scanner.ScanConfig, showFull bool) {
	dbPath := cfg.DatabasePath
	if dbPath == "" {
		dbPath = "Default"
	}
	exclude := ""
	if cfg.ExcludeRegex != nil {
		exclude = cfg.ExcludeRegex.String()
	}

	displayCfg := DisplayConfig{
		Paths:              cfg.TargetPaths,
		Database:           dbPath,
		Recursive:          cfg.Recursive,
		IncludeGit:         cfg.IncludeGit,
		IncludeNodeModules: cfg.IncludeNodeModules,
		ExcludeRegex:       exclude,
		FullReport:         showFull,
	}

	displayResults, summary := summarizeScanResults(results, displayCfg.FullReport)

	fmt.Println()
	color.Blue(style.EmojiText("🔍", "Scan Configuration"))
	color.Blue("─────────────────────")

	fmt.Printf("%-22s: %s\n", "Paths", listOrDash(displayCfg.Paths))
	fmt.Printf("%-22s: %s\n", "Database", stringOrDash(displayCfg.Database))
	fmt.Printf("%-22s: %s\n", "Recursive", yesNo(displayCfg.Recursive))
	fmt.Printf("%-22s: %s\n", "Include .git", yesNo(displayCfg.IncludeGit))
	fmt.Printf("%-22s: %s\n", "Include node_modules", yesNo(displayCfg.IncludeNodeModules))
	fmt.Printf("%-22s: %s\n", "Exclude Regex", stringOrDash(displayCfg.ExcludeRegex))
	fmt.Println()

	color.Blue(style.EmojiText("🚀", "Project Reports"))
	color.Blue("──────────────────")

	for _, r := range displayResults {
		count := r.InfectedCount()
		if count > 0 {
			fmt.Println()
			color.Red(
				style.EmojiText("🚫", fmt.Sprintf("[INFECTED] %s", r.ProjectPath)),
			)
			fmt.Printf("   %s %s\n", style.EmojiText("📄", "Lockfiles:"), strings.Join(r.Lockfiles, ", "))
			fmt.Printf("   %s %d\n", style.EmojiText("🦠", "Infected Packages:"), count)
			for _, v := range r.InfectedPackages {
				fmt.Printf("      - %s@%s\n", v.PackageName, v.Version)
			}
		} else {
			fmt.Println()
			color.Blue(
				style.EmojiText("✅", fmt.Sprintf("[CLEAN]    %s", r.ProjectPath)),
			)
		}
	}

	fmt.Println()

	color.Blue(style.EmojiText("📊", "Global Summary"))
	color.Blue("─────────────────")
	fmt.Printf("Total Projects: %d\n", summary.TotalProjects)
	fmt.Printf("Infected:       %d\n", summary.InfectedProjects)
	fmt.Printf("Clean:          %d\n", summary.TotalProjects-summary.InfectedProjects)
	fmt.Printf("Total Issues:   %d\n", summary.TotalInfectedPackages)
	fmt.Println()

	if summary.TotalProjects == 0 {
		color.Yellow(style.EmojiText("⚠️", "No lockfile found"))
		os.Exit(1)
	} else if summary.InfectedProjects == 0 {
		color.Green(
			style.EmojiText("🎉", "No project infected. You are safe!"),
		)
		os.Exit(0)
	} else {
		color.Red(
			style.EmojiText("❌", fmt.Sprintf("Found %d infected projects!", summary.InfectedProjects)),
		)
		os.Exit(1)
	}
}
