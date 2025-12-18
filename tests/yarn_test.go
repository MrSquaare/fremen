package tests

import (
	"path/filepath"
)

func (s *FremenTestSuite) Test_Yarn_Classic_Infected() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "classic_infected")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.Equal(1, report.Summary.InfectedProjects)

	s.Len(report.Results, 1)
	s.Contains(report.Results[0].InfectedPackages, InfectedPackage{Name: "test-package", Version: "1.0.0"})
}

func (s *FremenTestSuite) Test_Yarn_Classic_Variations() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "classic_variations")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.TotalInfectedPackages, 2)
}

func (s *FremenTestSuite) Test_Yarn_Modern_Infected() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "modern_infected")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)

	s.Len(report.Results, 1)
	s.Contains(report.Results[0].InfectedPackages, InfectedPackage{Name: "spaced-package", Version: "2.0.0"})
}

func (s *FremenTestSuite) Test_Yarn_Modern_Variations() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "modern_variations")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.TotalInfectedPackages, 2)
}

func (s *FremenTestSuite) Test_Yarn_Classic_Clean() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "classic_clean")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_Yarn_Modern_Clean() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "modern_clean")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_Yarn_Classic_Empty() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "empty")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_Yarn_Classic_Malformed() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "classic_malformed")
	_, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
}

func (s *FremenTestSuite) Test_Yarn_Modern_Malformed() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "modern_malformed")
	_, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
}

func (s *FremenTestSuite) Test_Yarn_Recursive() {
	target := filepath.Join(s.fixturesDir, "cases", "yarn", "recursive")
	report, exitCode := s.runFremenJSON(target, "--recursive")
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.InfectedProjects, 1)
}
