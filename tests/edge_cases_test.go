package tests

import (
	"path/filepath"
	"runtime"
)

func (s *FremenTestSuite) Test_EdgeCase_MixedLockfiles() {
	target := filepath.Join(s.fixturesDir, "cases", "edges", "mixed_lockfiles")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.GreaterOrEqual(report.Summary.TotalInfectedPackages, 1)
}

func (s *FremenTestSuite) Test_EdgeCase_NoLockfiles() {
	target := filepath.Join(s.fixturesDir, "cases", "edges", "no_lockfiles")
	report, exitCode := s.runFremenJSON(target)
	s.Equal(1, exitCode)
	s.Equal(0, report.Summary.TotalProjects)
}

func (s *FremenTestSuite) Test_EdgeCase_CaseSensitivity() {
	target := filepath.Join(s.fixturesDir, "cases", "edges", "case_sensitivity")
	report, _ := s.runFremenJSON(target, "--recursive")

	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		s.Equal(0, report.Summary.InfectedProjects, "Should NOT find infected projects in Node_Modules on case-insensitive OS")
	} else {
		s.GreaterOrEqual(report.Summary.InfectedProjects, 1, "Should find infected projects in Node_Modules on case-sensitive OS")
	}
}
