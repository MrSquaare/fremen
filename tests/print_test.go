package tests

import (
	"fmt"
	"path/filepath"
)

func (s *FremenTestSuite) Test_Print_Default() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_infected")
	output, exitCode := s.runFremen(target)
	s.Equal(1, exitCode)
	s.Contains(output, "\x1b[")
	s.Contains(output, "🚫")
}

func (s *FremenTestSuite) Test_Print_Infected() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_infected")
	output, exitCode := s.runFremen(target, "--no-color")
	s.Equal(1, exitCode)

	expected := fmt.Sprintf(`Loaded 2 infected package versions from database.txt.

🔍 Scan Configuration
─────────────────────
Paths                 : %s
Database              : %s
Recursive             : No
Include .git          : No
Include node_modules  : No
Exclude Regex         : -

🚀 Project Reports
──────────────────

🚫 [INFECTED] %s
   📄 Lockfiles: package-lock.json
   🦠 Infected Packages: 1
      - test-package@1.0.0

📊 Global Summary
─────────────────
Total Projects: 1
Infected:       1
Clean:          0
Total Issues:   1

❌ Found 1 infected projects!
`, target, s.databasePath, target)

	s.Equal(expected, output)
}

func (s *FremenTestSuite) Test_Print_Clean() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_clean")
	output, exitCode := s.runFremen(target, "--no-color")
	s.Equal(0, exitCode)

	expected := fmt.Sprintf(`Loaded 2 infected package versions from database.txt.

🔍 Scan Configuration
─────────────────────
Paths                 : %s
Database              : %s
Recursive             : No
Include .git          : No
Include node_modules  : No
Exclude Regex         : -

🚀 Project Reports
──────────────────

📊 Global Summary
─────────────────
Total Projects: 1
Infected:       0
Clean:          1
Total Issues:   0

🎉 No project infected. You are safe!
`, target, s.databasePath)

	s.Equal(expected, output)
}

func (s *FremenTestSuite) Test_Print_Clean_FullReport() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_clean")
	output, exitCode := s.runFremen(target, "--no-color", "--full-report")
	s.Equal(0, exitCode)

	expected := fmt.Sprintf(`Loaded 2 infected package versions from database.txt.

🔍 Scan Configuration
─────────────────────
Paths                 : %s
Database              : %s
Recursive             : No
Include .git          : No
Include node_modules  : No
Exclude Regex         : -

🚀 Project Reports
──────────────────

✅ [CLEAN]    %s

📊 Global Summary
─────────────────
Total Projects: 1
Infected:       0
Clean:          1
Total Issues:   0

🎉 No project infected. You are safe!
`, target, s.databasePath, target)

	s.Equal(expected, output)
}

func (s *FremenTestSuite) Test_Print_NoLockfile() {
	target := filepath.Join(s.fixturesDir, "cases", "edges", "no_lockfiles")
	output, exitCode := s.runFremen(target, "--no-color")
	s.Equal(1, exitCode)

	expected := fmt.Sprintf(`Loaded 2 infected package versions from database.txt.

🔍 Scan Configuration
─────────────────────
Paths                 : %s
Database              : %s
Recursive             : No
Include .git          : No
Include node_modules  : No
Exclude Regex         : -

🚀 Project Reports
──────────────────

📊 Global Summary
─────────────────
Total Projects: 0
Infected:       0
Clean:          0
Total Issues:   0

⚠️ No lockfile found
`, target, s.databasePath)

	s.Equal(expected, output)
}

func (s *FremenTestSuite) Test_Print_NoColor() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_infected")
	output, _ := s.runFremen(target, "--no-color")
	s.NotContains(output, "\x1b[")
}

func (s *FremenTestSuite) Test_Print_NoColorEnv() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_infected")

	s.T().Setenv("NO_COLOR", "1")
	output, _ := s.runFremen(target)
	s.NotContains(output, "\x1b[")
}

func (s *FremenTestSuite) Test_Print_NoEmoji() {
	target := filepath.Join(s.fixturesDir, "cases", "npm", "v1_infected")
	output, _ := s.runFremen(target, "--no-emoji")
	s.NotContains(output, "🚫")
}
