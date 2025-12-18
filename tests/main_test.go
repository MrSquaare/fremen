package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
)

func (s *FremenTestSuite) SetupSuite() {
	_, testFile, _, ok := runtime.Caller(0)
	s.Require().True(ok, "Failed to determine main test file location")

	testsDir := filepath.Dir(testFile)

	binName := "fremen-test-bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	s.binaryPath = filepath.Join(os.TempDir(), binName)

	buildArgs := []string{"build", "-o", s.binaryPath}
	_, coverageEnabled := os.LookupEnv("GOCOVERDIR")
	if coverageEnabled {
		buildArgs = append(buildArgs, "-cover", "-coverpkg=../...")
	}
	buildArgs = append(buildArgs, "../cmd/fremen")

	cmd := exec.Command("go", buildArgs...)
	output, err := cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to build fremen: %s", string(output))

	s.fixturesDir = filepath.Join(testsDir, "fixtures")

	dotGitPath := filepath.Join(s.fixturesDir, "cases", "filtering", "dot_git")
	gitPath := filepath.Join(s.fixturesDir, "cases", "filtering", ".git")
	if _, err := os.Stat(dotGitPath); err == nil {
		err = os.Rename(dotGitPath, gitPath)
		s.Require().NoError(err)
	}

	s.databasePath = filepath.Join(s.fixturesDir, "database.txt")
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
