package tests

import (
	"path/filepath"
)

func (s *FremenTestSuite) Test_PNPM_V5_Infected() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v5_infected")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)

	s.Len(report.Results, 1)
	s.Contains(report.Results[0].InfectedPackages, InfectedPackage{Name: "test-package", Version: "1.0.0"})
}

func (s *FremenTestSuite) Test_PNPM_V5_Variations() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v5_variations")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.TotalInfectedPackages, 2)
}

func (s *FremenTestSuite) Test_PNPM_V9_Infected() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v9_infected")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)

	s.Len(report.Results, 1)
	s.Contains(report.Results[0].InfectedPackages, InfectedPackage{Name: "test-package", Version: "1.0.0"})
}

func (s *FremenTestSuite) Test_PNPM_V9_Variations() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v9_variations")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.TotalInfectedPackages, 2)
}

func (s *FremenTestSuite) Test_PNPM_V5_Clean() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v5_clean")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_PNPM_V9_Clean() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v9_clean")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_PNPM_Empty() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "empty")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
	s.Equal(0, report.Summary.InfectedProjects)
}

func (s *FremenTestSuite) Test_PNPM_V5_Malformed() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v5_malformed")
	_, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
}

func (s *FremenTestSuite) Test_PNPM_V9_Malformed() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "v9_malformed")
	_, exitCode := s.runFremenJSON(target)
	s.Equal(0, exitCode)
}

func (s *FremenTestSuite) Test_PNPM_Recursive() {
	target := filepath.Join(s.fixturesDir, "cases", "pnpm", "recursive")
	report, exitCode := s.runFremenJSON(target, "--recursive")
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.InfectedProjects, 1)
}
