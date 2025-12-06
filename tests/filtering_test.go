package tests

import (
	"path/filepath"
	"strings"
)

func (s *FremenTestSuite) Test_Filtering_IgnoreDefaults() {
	target := filepath.Join(s.fixturesDir, "cases", "filtering")
	report, exitCode := s.runFremenJSON(target, "--recursive")

	s.Equal(1, exitCode)

	for _, res := range report.Results {
		s.False(strings.Contains(res.Project, "/.git"), "Project path should not contain .git: %s", res.Project)
		s.False(strings.Contains(res.Project, "/node_modules"), "Project path should not contain node_modules: %s", res.Project)
	}
}

func (s *FremenTestSuite) Test_Filtering_IncludeGit() {
	target := filepath.Join(s.fixturesDir, "cases", "filtering")
	report, exitCode := s.runFremenJSON(target, "--recursive", "--include-git")
	s.Equal(1, exitCode)

	var basePaths []string
	for _, res := range report.Results {
		basePaths = append(basePaths, filepath.Base(res.Project))
	}

	s.Contains(basePaths, ".git")
	s.GreaterOrEqual(len(report.Results), 2)
}

func (s *FremenTestSuite) Test_Filtering_IncludeNodeModules() {
	target := filepath.Join(s.fixturesDir, "cases", "filtering")
	report, exitCode := s.runFremenJSON(target, "--recursive", "--include-node-modules")
	s.Equal(1, exitCode)
	s.GreaterOrEqual(len(report.Results), 2)
}

func (s *FremenTestSuite) Test_Filtering_ExcludeRegex() {
	target := filepath.Join(s.fixturesDir, "cases", "filtering")
	report, exitCode := s.runFremenJSON(target, "--recursive", "--exclude", "custom_exclude")

	s.Equal(1, exitCode)
	s.Equal(0, report.Summary.TotalProjects)
}
