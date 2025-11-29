package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/MrSquaare/fremen/internal/scanner"
	"github.com/MrSquaare/fremen/internal/style"
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
	fmt.Println(style.StyledText(style.EmojiText("🔍", "Scan Configuration"), style.ColorBlue))
	fmt.Println(style.StyledText("─────────────────────", style.ColorBlue))

	fmt.Printf("%-22s: %s\n", "Paths", listOrDash(displayCfg.Paths))
	fmt.Printf("%-22s: %s\n", "Database", stringOrDash(displayCfg.Database))
	fmt.Printf("%-22s: %s\n", "Recursive", yesNo(displayCfg.Recursive))
	fmt.Printf("%-22s: %s\n", "Include .git", yesNo(displayCfg.IncludeGit))
	fmt.Printf("%-22s: %s\n", "Include node_modules", yesNo(displayCfg.IncludeNodeModules))
	fmt.Printf("%-22s: %s\n", "Exclude Regex", stringOrDash(displayCfg.ExcludeRegex))
	fmt.Println()

	fmt.Println(style.StyledText(style.EmojiText("🚀", "Project Reports"), style.ColorBlue))
	fmt.Println(style.StyledText("──────────────────", style.ColorBlue))

	for _, r := range displayResults {
		count := r.InfectedCount()
		if count > 0 {
			fmt.Println()
			fmt.Println(style.StyledText(
				style.EmojiText("🚫", fmt.Sprintf("[INFECTED] %s", r.ProjectPath)),
				style.ColorRed,
			))
			fmt.Printf("   %s %s\n", style.EmojiText("📄", "Lockfiles:"), strings.Join(r.Lockfiles, ", "))
			fmt.Printf("   %s %d\n", style.EmojiText("🦠", "Infected Packages:"), count)
			for _, v := range r.InfectedPackages {
				fmt.Printf("      - %s@%s\n", v.PackageName, v.Version)
			}
		} else {
			fmt.Println()
			fmt.Println(style.StyledText(
				style.EmojiText("✅", fmt.Sprintf("[CLEAN]    %s", r.ProjectPath)),
				style.ColorGreen,
			))
		}
	}

	fmt.Println()

	fmt.Println(style.StyledText(style.EmojiText("📊", "Global Summary"), style.ColorBlue))
	fmt.Println(style.StyledText("─────────────────", style.ColorBlue))
	fmt.Printf("Total Projects: %d\n", summary.TotalProjects)
	fmt.Printf("Infected:       %d\n", summary.InfectedProjects)
	fmt.Printf("Clean:          %d\n", summary.TotalProjects-summary.InfectedProjects)
	fmt.Printf("Total Issues:   %d\n", summary.TotalInfectedPackages)
	fmt.Println()

	if summary.TotalProjects == 0 {
		fmt.Println(style.StyledText(style.EmojiText("⚠️", "No lockfile found"), style.ColorYellow))
		os.Exit(1)
	} else if summary.InfectedProjects == 0 {
		fmt.Println(style.StyledText(
			style.EmojiText("🎉", "No project infected. You are safe!"),
			style.ColorGreen,
		))
		os.Exit(0)
	} else {
		fmt.Println(style.StyledText(
			style.EmojiText("❌", fmt.Sprintf("Found %d infected projects!", summary.InfectedProjects)),
			style.ColorRed,
		))
		os.Exit(1)
	}
}
