package report

import (
	"sort"
	"strings"

	"github.com/MrSquaare/fremen/internal/scanner"
)

type Summary struct {
	TotalProjects         int `json:"total_projects"`
	InfectedProjects      int `json:"infected_projects"`
	TotalInfectedPackages int `json:"total_infected_packages"`
}

func summarizeScanResults(results []scanner.ScanResult, showFull bool) ([]scanner.ScanResult, Summary) {
	totalProjects := len(results)
	infectedProjects := 0
	totalInfected := 0

	displayResults := make([]scanner.ScanResult, 0, len(results))

	for _, r := range results {
		count := r.InfectedCount()
		if count > 0 {
			infectedProjects++
		}
		totalInfected += count

		if showFull || count > 0 {
			displayResults = append(displayResults, r)
		}
	}

	sort.Slice(displayResults, func(i, j int) bool {
		ri, rj := displayResults[i], displayResults[j]

		// Infected first
		ci, cj := ri.InfectedCount(), rj.InfectedCount()
		if (ci > 0) != (cj > 0) {
			return ci > 0
		}

		// Alphabetical (case-insensitive)
		return strings.ToLower(ri.ProjectPath) < strings.ToLower(rj.ProjectPath)
	})

	summary := Summary{
		TotalProjects:         totalProjects,
		InfectedProjects:      infectedProjects,
		TotalInfectedPackages: totalInfected,
	}

	return displayResults, summary
}
