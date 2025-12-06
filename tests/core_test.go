package tests

import (
	"encoding/json"
	"path/filepath"
)

func (s *FremenTestSuite) Test_Core_CustomDatabase() {
	target := filepath.Join(s.fixturesDir, "cases", "features", "custom_database")
	customDb := filepath.Join(target, "custom-database.txt")

	report, exitCode := s.runFremenJSON(target, "--database", customDb)
	s.Equal(1, exitCode)

	expected := Report{
		Configuration: Configuration{
			Paths:              []string{target},
			Database:           customDb,
			Recursive:          false,
			IncludeGit:         false,
			IncludeNodeModules: false,
			ExcludeRegex:       "",
			FullReport:         false,
		},
		Results: []Result{
			{
				InfectedCount: 1,
				InfectedPackages: []InfectedPackage{
					{Name: "custom-package", Version: "1.0.0"},
				},
				Lockfiles: []string{"package-lock.json"},
				Project:   target,
			},
		},
		Summary: Summary{
			TotalProjects:         1,
			InfectedProjects:      1,
			TotalInfectedPackages: 1,
		},
	}

	s.Equal(expected, report)
}

func (s *FremenTestSuite) Test_Core_FullReport() {
	target := filepath.Join(s.fixturesDir, "cases", "features", "full_report")
	report, exitCode := s.runFremenJSON(target, "-f")
	s.Equal(0, exitCode)

	expected := Report{
		Configuration: Configuration{
			Paths:              []string{target},
			Database:           s.databasePath,
			Recursive:          false,
			IncludeGit:         false,
			IncludeNodeModules: false,
			ExcludeRegex:       "",
			FullReport:         true,
		},
		Results: []Result{
			{
				InfectedCount:    0,
				InfectedPackages: []InfectedPackage{},
				Lockfiles:        []string{"package-lock.json"},
				Project:          target,
			},
		},
		Summary: Summary{
			TotalProjects:         1,
			InfectedProjects:      0,
			TotalInfectedPackages: 0,
		},
	}

	s.Equal(expected, report)
}

func (s *FremenTestSuite) Test_Core_JSONValidity() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_infected")
	output, _ := s.runFremen(target, "--json")

	var js map[string]interface{}
	err := json.Unmarshal([]byte(output), &js)
	s.NoError(err, "Output should be valid JSON")

	s.NotEmpty(js["configuration"])
	s.NotEmpty(js["results"])
}
