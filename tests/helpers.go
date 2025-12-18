package tests

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/stretchr/testify/suite"
)

type FremenTestSuite struct {
	suite.Suite
	binaryPath   string
	fixturesDir  string
	databasePath string
}

func (s *FremenTestSuite) runFremen(args ...string) (string, int) {
	s.T().Helper()
	hasDb := false
	for _, arg := range args {
		if arg == "--database" || strings.HasPrefix(arg, "--database=") {
			hasDb = true
			break
		}
	}

	finalArgs := args
	if !hasDb {
		finalArgs = append(finalArgs, "--database", s.databasePath)
	}

	cmd := exec.Command(s.binaryPath, finalArgs...)
	output, err := cmd.Output()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return string(output), exitCode
}

func (s *FremenTestSuite) runFremenJSON(args ...string) (Report, int) {
	s.T().Helper()
	args = append(args, "--json")
	output, exitCode := s.runFremen(args...)

	var report Report
	err := json.Unmarshal([]byte(output), &report)
	s.Require().NoError(err, "Failed to unmarshal JSON output: %s", output)

	return report, exitCode
}
