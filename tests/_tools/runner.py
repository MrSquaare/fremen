#!/usr/bin/env python3
"""
Test Runner for Fremen Security Scanner.

This script executes a suite of integration tests against the fremen.py script.
It supports parallel execution, JSON output validation, and output content matching.
"""

import argparse
import concurrent.futures
import json
import os
import shutil
import subprocess
import sys
import time
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional


# --- Configuration & Constants ---


class TerminalColors:
    """ANSI Color codes for terminal output."""

    RED = "\033[91m"
    GREEN = "\033[92m"
    YELLOW = "\033[93m"
    BLUE = "\033[94m"
    CYAN = "\033[96m"
    RESET = "\033[0m"
    BOLD = "\033[1m"


def styled_text(text: str, color_code: str) -> str:
    """Wraps text in ANSI color codes."""
    return f"{color_code}{text}{TerminalColors.RESET}"


@dataclass
class TestCase:
    """Represents a single test case configuration."""

    name: str
    target_paths: List[str]
    expected_exit_code: int
    expected_json_subset: Optional[Dict[str, Any]] = None
    extra_arguments: List[str] = field(default_factory=list)
    use_json_output: bool = True
    stdout_must_contain: List[str] = field(default_factory=list)
    stderr_must_contain: List[str] = field(default_factory=list)
    stdout_must_not_contain: List[str] = field(default_factory=list)


@dataclass
class TestResult:
    """Represents the result of a single test execution."""

    test_case: TestCase
    is_successful: bool
    duration_seconds: float
    error_message: Optional[str] = None
    output_content: str = ""


class TestRunner:
    """Manages the execution of test cases."""

    def __init__(self, test_root_directory: str):
        self.test_root_directory = test_root_directory
        self.cases_directory = os.path.join(test_root_directory, "cases")
        self.tools_directory = os.path.join(test_root_directory, "_tools")

        # Paths for the script under test
        self.source_script_path = os.path.join(
            os.path.dirname(test_root_directory), "fremen.py"
        )
        self.test_script_destination = os.path.join(self.tools_directory, "fremen.py")
        self.database_path = os.path.join(self.tools_directory, "database.txt")

    def setup_environment(self):
        """Prepare the test environment by copying the script to the tools directory."""
        shutil.copy(self.source_script_path, self.test_script_destination)

        # Rename dot_git to .git for testing
        dot_git_path = os.path.join(self.cases_directory, "ignored", "dot_git")
        git_path = os.path.join(self.cases_directory, "ignored", ".git")
        if os.path.exists(dot_git_path):
            os.rename(dot_git_path, git_path)

    def teardown_environment(self):
        """Clean up the test environment."""
        if os.path.exists(self.test_script_destination):
            os.remove(self.test_script_destination)

        # Rename .git back to dot_git
        dot_git_path = os.path.join(self.cases_directory, "ignored", "dot_git")
        git_path = os.path.join(self.cases_directory, "ignored", ".git")
        if os.path.exists(git_path):
            os.rename(git_path, dot_git_path)

    def run_test_case(self, test_case: TestCase) -> TestResult:
        """Execute a single test case and return the result."""
        start_time = time.time()

        # Construct the base command
        command_arguments = (
            [sys.executable, self.test_script_destination]
            + test_case.target_paths
            + ["--database", self.database_path]
        )

        if test_case.use_json_output:
            command_arguments.append("--json")

        # Handle extra arguments and potential database override
        if test_case.extra_arguments:
            self._apply_extra_arguments(command_arguments, test_case.extra_arguments)

        try:
            process_result = subprocess.run(
                command_arguments, capture_output=True, text=True
            )

            duration = time.time() - start_time

            # Validate Exit Code
            if process_result.returncode != test_case.expected_exit_code:
                return TestResult(
                    test_case,
                    False,
                    duration,
                    f"Exit Code: Expected {test_case.expected_exit_code}, Got {process_result.returncode}",
                    process_result.stderr or process_result.stdout,
                )

            # Validate JSON Output
            if test_case.use_json_output:
                json_validation_error = self._validate_json_output(
                    test_case, process_result.stdout
                )
                if json_validation_error:
                    return TestResult(
                        test_case,
                        False,
                        duration,
                        json_validation_error,
                        process_result.stdout,
                    )

            # Validate Stdout Content
            for expected_item in test_case.stdout_must_contain:
                if expected_item not in process_result.stdout:
                    return TestResult(
                        test_case,
                        False,
                        duration,
                        f"Stdout missing: {expected_item}",
                        process_result.stdout,
                    )

            for forbidden_item in test_case.stdout_must_not_contain:
                if forbidden_item in process_result.stdout:
                    return TestResult(
                        test_case,
                        False,
                        duration,
                        f"Stdout contains forbidden: {forbidden_item}",
                        process_result.stdout,
                    )

            # Validate Stderr Content
            for expected_item in test_case.stderr_must_contain:
                if expected_item not in process_result.stderr:
                    return TestResult(
                        test_case,
                        False,
                        duration,
                        f"Stderr missing: {expected_item}",
                        process_result.stderr,
                    )

            return TestResult(test_case, True, duration)

        except Exception as exception:
            return TestResult(
                test_case, False, time.time() - start_time, str(exception)
            )

    def _apply_extra_arguments(
        self, command_arguments: List[str], extra_arguments: List[str]
    ):
        """Applies extra arguments, handling overrides like --database."""
        if "--database" in extra_arguments:
            # Remove the default database arg we added earlier
            try:
                database_index = command_arguments.index("--database")
                command_arguments.pop(database_index)  # remove --database
                command_arguments.pop(database_index)  # remove value
            except ValueError:
                pass

        command_arguments.extend(extra_arguments)

    def _validate_json_output(
        self, test_case: TestCase, stdout_content: str
    ) -> Optional[str]:
        """Parses and validates JSON output against the expected subset."""
        try:
            output_json = json.loads(stdout_content)

            if test_case.expected_json_subset:
                if not self._check_is_subset(
                    test_case.expected_json_subset, output_json
                ):
                    return "JSON Mismatch"

            return None

        except json.JSONDecodeError:
            return "Invalid JSON Output"

    def _check_is_subset(self, expected_structure: Any, actual_structure: Any) -> bool:
        """Recursively check if expected structure is a subset of actual structure."""
        if isinstance(expected_structure, dict):
            return self._check_dict_subset(expected_structure, actual_structure)

        elif isinstance(expected_structure, list):
            return self._check_list_subset(expected_structure, actual_structure)

        else:
            return expected_structure == actual_structure

    def _check_dict_subset(self, expected_dict: Dict, actual_dict: Any) -> bool:
        """Helper to check if a dictionary is a subset of another."""
        if not isinstance(actual_dict, dict):
            return False

        for key, value in expected_dict.items():
            if key not in actual_dict:
                return False

            if not self._check_is_subset(value, actual_dict[key]):
                return False

        return True

    def _check_list_subset(self, expected_list: List, actual_list: Any) -> bool:
        """Helper to check if a list contains expected items."""
        if not isinstance(actual_list, list):
            return False

        if not expected_list:
            return True

        for expected_item in expected_list:
            found_match = False
            for actual_item in actual_list:
                if self._check_is_subset(expected_item, actual_item):
                    found_match = True
                    break

            if not found_match:
                return False

        return True

    def run_all_tests(
        self, test_cases: List[TestCase], run_in_parallel: bool = True
    ) -> bool:
        """Run all provided test cases and print a summary."""
        print(
            f"{styled_text(f'Running {len(test_cases)} tests...', TerminalColors.BOLD)}\n"
        )

        results = []
        start_time_total = time.time()

        if run_in_parallel:
            self._run_tests_parallel(test_cases, results)
        else:
            self._run_tests_sequential(test_cases, results)

        print("\n")
        self._print_execution_summary(results, time.time() - start_time_total)

        return all(result.is_successful for result in results)

    def _run_tests_parallel(
        self, test_cases: List[TestCase], results_list: List[TestResult]
    ):
        """Executes tests using a thread pool."""
        with concurrent.futures.ThreadPoolExecutor() as executor:
            future_to_case_map = {
                executor.submit(self.run_test_case, case): case for case in test_cases
            }

            for future in concurrent.futures.as_completed(future_to_case_map):
                result = future.result()
                results_list.append(result)
                self._print_progress_dot(result)

    def _run_tests_sequential(
        self, test_cases: List[TestCase], results_list: List[TestResult]
    ):
        """Executes tests one by one."""
        for case in test_cases:
            result = self.run_test_case(case)
            results_list.append(result)
            self._print_progress_dot(result)

    def _print_progress_dot(self, result: TestResult):
        """Prints a dot or F to indicate progress."""
        if result.is_successful:
            print(styled_text(".", TerminalColors.GREEN), end="", flush=True)
        else:
            print(styled_text("F", TerminalColors.RED), end="", flush=True)

    def _print_execution_summary(
        self, results: List[TestResult], total_time_seconds: float
    ):
        """Prints the final summary of the test run."""
        failed_results = [r for r in results if not r.is_successful]
        passed_results = [r for r in results if r.is_successful]

        if failed_results:
            print(
                styled_text(
                    "=== FAILURES ===", TerminalColors.BOLD + TerminalColors.RED
                )
            )

            for failure in failed_results:
                print(f"\n{styled_text(failure.test_case.name, TerminalColors.BOLD)}")
                print(
                    f"  {styled_text(f'Error: {failure.error_message}', TerminalColors.RED)}"
                )

                if failure.output_content:
                    # Indent output for readability
                    output_lines = failure.output_content.splitlines()
                    indented_output = "\n".join(
                        f"    {line}" for line in output_lines[:20]
                    )

                    print(f"  Output Snippet:\n{indented_output}")

                    if len(output_lines) > 20:
                        print("    ... (truncated)")

        print(f"\n{styled_text('Test Summary:', TerminalColors.BOLD)}")
        print(f"  Total:  {len(results)}")
        print(
            f"  Passed: {styled_text(str(len(passed_results)), TerminalColors.GREEN)}"
        )
        print(f"  Failed: {styled_text(str(len(failed_results)), TerminalColors.RED)}")
        print(f"  Time:   {total_time_seconds:.2f}s")


def get_test_cases(cases_directory: str, tools_directory: str) -> List[TestCase]:
    """
    Defines and returns the list of all test cases.

    Args:
        cases_directory: Path to the directory containing test case fixtures.
        tools_directory: Path to the directory containing test tools.

    Returns:
        A list of TestCase objects.
    """
    return [
        # --- NPM Tests ---
        TestCase(
            name="NPM v1 Infected",
            target_paths=[os.path.join(cases_directory, "npm/v1_infected")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "test-package", "version": "1.0.0"}
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="NPM v2 Infected",
            target_paths=[os.path.join(cases_directory, "npm/v2_infected")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "spaced-package", "version": "2.0.0"}
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="NPM v1 Clean",
            target_paths=[os.path.join(cases_directory, "npm/v1_clean")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}, "results": []},
        ),
        TestCase(
            name="NPM v2 Clean",
            target_paths=[os.path.join(cases_directory, "npm/v2_clean")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}, "results": []},
        ),
        TestCase(
            name="NPM Empty",
            target_paths=[os.path.join(cases_directory, "npm/empty")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="NPM Malformed",
            target_paths=[os.path.join(cases_directory, "npm/malformed")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="NPM Recursive",
            target_paths=[os.path.join(cases_directory, "npm/recursive")],
            extra_arguments=["-r"],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "project": os.path.join(
                            cases_directory, "npm/recursive/level1/level2"
                        )
                    }
                ],
            },
        ),
        # --- Yarn Tests ---
        TestCase(
            name="Yarn Classic Infected (Double Quote)",
            target_paths=[os.path.join(cases_directory, "yarn/classic_infected")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "test-package", "version": "1.0.0"}
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="Yarn Classic Variations (Single/No Quote)",
            target_paths=[os.path.join(cases_directory, "yarn/classic_variations")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "test-package", "version": "1.0.0"},
                            {"name": "spaced-package", "version": "2.0.0"},
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="Yarn Modern Infected (Double Quote)",
            target_paths=[os.path.join(cases_directory, "yarn/modern_infected")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "spaced-package", "version": "2.0.0"}
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="Yarn Modern Variations (Single/No Quote)",
            target_paths=[os.path.join(cases_directory, "yarn/modern_variations")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "spaced-package", "version": "2.0.0"},
                            {"name": "test-package", "version": "1.0.0"},
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="Yarn Classic Clean",
            target_paths=[os.path.join(cases_directory, "yarn/classic_clean")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="Yarn Modern Clean",
            target_paths=[os.path.join(cases_directory, "yarn/modern_clean")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="Yarn Classic Empty",
            target_paths=[os.path.join(cases_directory, "yarn/empty")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="Yarn Classic Malformed",
            target_paths=[os.path.join(cases_directory, "yarn/classic_malformed")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="Yarn Modern Malformed",
            target_paths=[os.path.join(cases_directory, "yarn/modern_malformed")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="Yarn Recursive",
            target_paths=[os.path.join(cases_directory, "yarn/recursive")],
            extra_arguments=["-r"],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "project": os.path.join(
                            cases_directory, "yarn/recursive/level1/level2"
                        )
                    }
                ],
            },
        ),
        # --- PNPM Tests ---
        TestCase(
            name="PNPM v5 Infected (Single Quote)",
            target_paths=[os.path.join(cases_directory, "pnpm/v5_infected")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "test-package", "version": "1.0.0"}
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="PNPM v5 Variations (No/Double Quote)",
            target_paths=[os.path.join(cases_directory, "pnpm/v5_variations")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "test-package", "version": "1.0.0"},
                            {"name": "spaced-package", "version": "2.0.0"},
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="PNPM v9 Infected (Single Quote)",
            target_paths=[os.path.join(cases_directory, "pnpm/v9_infected")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "test-package", "version": "1.0.0"},
                            {"name": "spaced-package", "version": "2.0.0"},
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="PNPM v9 Variations (No/Double Quote)",
            target_paths=[os.path.join(cases_directory, "pnpm/v9_variations")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "test-package", "version": "1.0.0"},
                            {"name": "spaced-package", "version": "2.0.0"},
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="PNPM v5 Clean",
            target_paths=[os.path.join(cases_directory, "pnpm/v5_clean")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="PNPM v9 Clean",
            target_paths=[os.path.join(cases_directory, "pnpm/v9_clean")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="PNPM Empty",
            target_paths=[os.path.join(cases_directory, "pnpm/empty")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="PNPM v5 Malformed",
            target_paths=[os.path.join(cases_directory, "pnpm/v5_malformed")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="PNPM v9 Malformed",
            target_paths=[os.path.join(cases_directory, "pnpm/v9_malformed")],
            expected_exit_code=0,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="PNPM Recursive",
            target_paths=[os.path.join(cases_directory, "pnpm/recursive")],
            extra_arguments=["-r"],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "project": os.path.join(
                            cases_directory, "pnpm/recursive/level1/level2"
                        )
                    }
                ],
            },
        ),
        # --- Filtering Tests ---
        TestCase(
            name="Ignore Defaults (Recursion)",
            target_paths=[os.path.join(cases_directory, "ignored")],
            extra_arguments=["-r"],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {"project": os.path.join(cases_directory, "ignored/custom_exclude")}
                ],
            },
        ),
        TestCase(
            name="Include .git",
            target_paths=[os.path.join(cases_directory, "ignored")],
            extra_arguments=["--include-git", "-r"],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 2},
                "results": [
                    {
                        "project": os.path.join(
                            cases_directory, "ignored/custom_exclude"
                        )
                    },
                    {"project": os.path.join(cases_directory, "ignored/.git")},
                ],
            },
        ),
        TestCase(
            name="Include node_modules",
            target_paths=[os.path.join(cases_directory, "ignored")],
            extra_arguments=["--include-node-modules", "-r"],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 2},
                "results": [
                    {
                        "project": os.path.join(
                            cases_directory, "ignored/custom_exclude"
                        )
                    },
                    {"project": os.path.join(cases_directory, "ignored/node_modules")},
                ],
            },
        ),
        TestCase(
            name="Exclude Regex",
            target_paths=[os.path.join(cases_directory, "ignored")],
            extra_arguments=["--exclude", "custom_exclude", "-r"],
            expected_exit_code=1,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        # --- Feature Tests ---
        TestCase(
            name="Recursion Flag Check (No Recursion)",
            target_paths=[os.path.join(cases_directory, "pnpm/recursive")],
            expected_exit_code=1,
            expected_json_subset={"summary": {"infected_projects": 0}},
        ),
        TestCase(
            name="Custom Database",
            target_paths=[os.path.join(cases_directory, "custom_db")],
            extra_arguments=[
                "--database",
                os.path.join(tools_directory, "custom-db.txt"),
            ],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "infected_packages": [
                            {"name": "custom-package", "version": "1.0.0"}
                        ]
                    }
                ],
            },
        ),
        TestCase(
            name="Full Report Clean",
            target_paths=[os.path.join(cases_directory, "features/full_report_clean")],
            extra_arguments=["-f"],
            expected_exit_code=0,
            expected_json_subset={
                "summary": {"infected_projects": 0, "total_projects": 1},
                "results": [
                    {
                        "project": os.path.join(
                            cases_directory, "features/full_report_clean"
                        )
                    }
                ],
            },
        ),
        TestCase(
            name="Mixed Lockfiles",
            target_paths=[os.path.join(cases_directory, "features/mixed_lockfiles")],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1},
                "results": [
                    {
                        "lockfiles": ["package-lock.json", "yarn.lock"],
                        "infected_count": 1,
                    }
                ],
            },
        ),
        TestCase(
            name="Case Sensitivity (Node_Modules)",
            target_paths=[os.path.join(cases_directory, "features/case_sensitivity")],
            extra_arguments=["-r"],
            expected_exit_code=0,
            expected_json_subset={
                "summary": {"infected_projects": 0, "total_projects": 1}
            },
        ),
        TestCase(
            name="Case Sensitivity (Include Node_Modules)",
            target_paths=[os.path.join(cases_directory, "features/case_sensitivity")],
            extra_arguments=["-r", "--include-node-modules"],
            expected_exit_code=1,
            expected_json_subset={
                "summary": {"infected_projects": 1, "total_projects": 2},
                "results": [
                    {
                        "project": os.path.join(
                            cases_directory, "features/case_sensitivity/Node_Modules"
                        )
                    }
                ],
            },
        ),
        TestCase(
            name="No Lockfiles",
            target_paths=[os.path.join(cases_directory, "features/no_lockfiles")],
            expected_exit_code=1,
            expected_json_subset={"summary": {"total_projects": 0}},
        ),
        TestCase(
            name="No Color Flag",
            target_paths=[os.path.join(cases_directory, "features/mixed_lockfiles")],
            extra_arguments=["--no-color"],
            use_json_output=False,
            expected_exit_code=1,
            stdout_must_not_contain=["\033["],
            stdout_must_contain=["Infected Packages:"],
        ),
        TestCase(
            name="No Emoji Flag",
            target_paths=[os.path.join(cases_directory, "features/mixed_lockfiles")],
            extra_arguments=["--no-emoji"],
            use_json_output=False,
            expected_exit_code=1,
            stdout_must_not_contain=["🚫", "🦠", "📄"],
            stdout_must_contain=["Infected Packages:"],
        ),
    ]


def main():
    """Main entry point for the test runner."""
    parser = argparse.ArgumentParser(description="Fremen Test Runner")
    parser.add_argument("-k", "--keyword", help="Filter tests by keyword", type=str)
    parser.add_argument(
        "--no-parallel", help="Disable parallel execution", action="store_true"
    )

    parsed_arguments = parser.parse_args()

    test_root_directory = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    test_runner = TestRunner(test_root_directory)

    try:
        test_runner.setup_environment()

        all_test_cases = get_test_cases(
            test_runner.cases_directory, test_runner.tools_directory
        )

        if parsed_arguments.keyword:
            all_test_cases = [
                test_case
                for test_case in all_test_cases
                if parsed_arguments.keyword.lower() in test_case.name.lower()
            ]

            if not all_test_cases:
                print(
                    styled_text(
                        f"No tests matched keyword '{parsed_arguments.keyword}'",
                        TerminalColors.YELLOW,
                    )
                )
                return

        success = test_runner.run_all_tests(
            all_test_cases, run_in_parallel=not parsed_arguments.no_parallel
        )

        if not success:
            sys.exit(1)

    finally:
        test_runner.teardown_environment()


if __name__ == "__main__":
    main()
