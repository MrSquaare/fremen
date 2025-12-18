package tests

import (
	"path/filepath"
)

func (s *FremenTestSuite) Test_NPM_V1_Infected() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_infected")
	report, exitCode := s.runFremenJSON(target)

	s.Equal(1, exitCode)

	expected := Report{
		Configuration: Configuration{
			Paths:              []string{target},
			Database:           s.databasePath,
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
					{Name: "test-package", Version: "1.0.0"},
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

func (s *FremenTestSuite) Test_NPM_V1_Clean() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_clean")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)

	expected := Report{
		Configuration: Configuration{
			Paths:              []string{target},
			Database:           s.databasePath,
			Recursive:          false,
			IncludeGit:         false,
			IncludeNodeModules: false,
			ExcludeRegex:       "",
			FullReport:         false,
		},
		Results: []Result{},
		Summary: Summary{
			TotalProjects:         1,
			InfectedProjects:      0,
			TotalInfectedPackages: 0,
		},
	}

	s.Equal(expected, report)
}

func (s *FremenTestSuite) Test_NPM_V2_Infected() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v2_infected")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)

	s.Equal(1, report.Summary.InfectedProjects)
	s.Len(report.Results, 1)
	s.Contains(report.Results[0].InfectedPackages, InfectedPackage{Name: "test-package", Version: "1.0.0"})
}

func (s *FremenTestSuite) Test_NPM_V2_Clean() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v2_clean")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_NPM_Empty() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "empty")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.Equal(0, report.Summary.TotalProjects)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_NPM_Malformed() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "malformed")
	_, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
}

func (s *FremenTestSuite) Test_NPM_Recursive() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "recursive")
	report, exitCode := s.runFremenJSON(target, "--recursive")
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.InfectedProjects, 1)

	found := false
	for _, res := range report.Results {
		if filepath.Base(res.Project) == "level2" {
			found = true
			s.Contains(res.InfectedPackages, InfectedPackage{Name: "test-package", Version: "1.0.0"})
			break
		}
	}
	s.True(found)
}
