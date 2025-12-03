package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

func (s *FremenTestSuite) SetupSuite() {
	cwd, err := os.Getwd()
	s.Require().NoError(err)

	projectRoot := cwd
	if filepath.Base(cwd) == "tests" {
		projectRoot = filepath.Dir(cwd)
	}

	s.binaryPath = filepath.Join(os.TempDir(), "fremen-test-bin")

	cmd := exec.Command("go", "build", "-o", s.binaryPath, filepath.Join(projectRoot, "cmd", "fremen", "main.go"))
	output, err := cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to build fremen: %s", string(output))

	s.fixturesDir = filepath.Join(projectRoot, "tests", "fixtures")
	s.databasePath = filepath.Join(s.fixturesDir, "database.txt")

	dotGitPath := filepath.Join(s.fixturesDir, "cases", "filtering", "dot_git")
	gitPath := filepath.Join(s.fixturesDir, "cases", "filtering", ".git")
	if _, err := os.Stat(dotGitPath); err == nil {
		err = os.Rename(dotGitPath, gitPath)
		s.Require().NoError(err)
	}
}

func (s *FremenTestSuite) TearDownSuite() {
	_ = os.Remove(s.binaryPath)

	gitPath := filepath.Join(s.fixturesDir, "cases", "filtering", ".git")
	dotGitPath := filepath.Join(s.fixturesDir, "cases", "filtering", "dot_git")
	if _, err := os.Stat(gitPath); err == nil {
		_ = os.Rename(gitPath, dotGitPath)
	}
}

func TestFremenSuite(t *testing.T) {
	suite.Run(t, new(FremenTestSuite))
}
