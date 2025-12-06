package tests

import (
	"path/filepath"
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
