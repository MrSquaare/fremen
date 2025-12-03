package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/MrSquaare/fremen/internal/database"
	"github.com/MrSquaare/fremen/internal/report"
	"github.com/MrSquaare/fremen/internal/scanner"
	"github.com/MrSquaare/fremen/internal/style"
)

var (
	flagRecursive          bool
	flagIncludeGit         bool
	flagIncludeNodeModules bool
	flagExclude            string
	flagFullReport         bool
	flagJSON               bool
	flagNoColor            bool
	flagNoEmoji            bool
	flagDatabase           string
)

var rootCmd = &cobra.Command{
	Use:   "fremen",
	Short: "Fast Lockfile Scanner for Infected Packages",
	RunE:  runRoot,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagRecursive, "recursive", "r", false, "Scan directories recursively")
	rootCmd.Flags().BoolVarP(&flagIncludeGit, "include-git", "g", false, "Include .git directories during recursion")
	rootCmd.Flags().BoolVarP(&flagIncludeNodeModules, "include-node-modules", "n", false, "Include node_modules directories during recursion")
	rootCmd.Flags().StringVarP(&flagExclude, "exclude", "e", "", "Exclude paths matching this regex")

	rootCmd.Flags().BoolVarP(&flagFullReport, "full-report", "f", false, "Display projects that are not infected")
	rootCmd.Flags().BoolVarP(&flagJSON, "json", "j", false, "Output results in JSON format")
	rootCmd.Flags().BoolVarP(&flagNoColor, "no-color", "C", false, "Disable ANSI colors in the CLI report")
	rootCmd.Flags().BoolVarP(&flagNoEmoji, "no-emoji", "E", false, "Disable emoji icons in the CLI report")

	rootCmd.Flags().StringVarP(&flagDatabase, "database", "d", "", "Path to database.txt database file")
}

func runRoot(cmd *cobra.Command, args []string) error {
	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}
	var color, emoji = true, true
	if noColorEnv, ok := os.LookupEnv("NO_COLOR"); (ok && noColorEnv != "0") || flagNoColor {
		color = false
	}
	if flagNoEmoji {
		emoji = false
	}

	termStyler := style.NewTermStyler(color, emoji)

	var excludeRegex *regexp.Regexp
	if flagExclude != "" {
		r, err := regexp.Compile(flagExclude)
		if err != nil {
			termStyler.PrintLnColor(
				fmt.Sprintf("Invalid exclude regex: %v", err),
				style.ColorRed,
			)
			os.Exit(1)
		}
		excludeRegex = r
	}

	cfg := scanner.ScanConfig{
		TargetPaths:        paths,
		DatabasePath:       flagDatabase,
		Recursive:          flagRecursive,
		IncludeGit:         flagIncludeGit,
		IncludeNodeModules: flagIncludeNodeModules,
		ExcludeRegex:       excludeRegex,
	}

	db := database.New()
	if err := db.Load(cfg.DatabasePath); err != nil {
		if !flagJSON {
			termStyler.PrintLnColor(
				fmt.Sprintf("Error: %s", err),
				style.ColorRed,
			)
		}
		os.Exit(1)
	}
	if !flagJSON {
		termStyler.PrintLnColor(
			fmt.Sprintf("Loaded %d infected package versions from %s.", db.EntryCount(), filepath.Base(db.LoadedPath())),
			style.ColorBlue,
		)

	}

	results, err := scanner.ExecuteScan(cfg, db)
	if err != nil {
		termStyler.PrintLnColor(
			fmt.Sprintf("Scan error: %v", err),
			style.ColorRed,
		)
		os.Exit(1)
	}

	if flagJSON {
		report.PrintJSONReport(results, cfg, flagFullReport)
	} else {
		report.PrintCLIReport(*termStyler, results, cfg, flagFullReport)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
