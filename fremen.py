#!/usr/bin/env python3
"""
Fremen Security Scanner.

A fast, parallelized security scanner for detecting infected packages in lockfiles.
Supports npm (package-lock.json), yarn (yarn.lock), and pnpm (pnpm-lock.yaml).
"""

import argparse
import concurrent.futures
import ctypes
import json
import os
import re
import sys
import time
from collections import defaultdict
from dataclasses import dataclass
from typing import Any, Dict, Iterable, List, Optional, Pattern, Set, TextIO, Tuple


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


def _configure_console_colors() -> bool:
    """Best-effort enable ANSI colors on Windows, else report lack of support."""
    if os.name != "nt":
        return True

    try:
        kernel32 = ctypes.windll.kernel32  # type: ignore[attr-defined]
        handle = kernel32.GetStdHandle(-11)  # STD_OUTPUT_HANDLE
        mode = ctypes.c_uint32()

        if not kernel32.GetConsoleMode(handle, ctypes.byref(mode)):
            return False

        enable_vt = 0x0004  # ENABLE_VIRTUAL_TERMINAL_PROCESSING
        if kernel32.SetConsoleMode(handle, mode.value | enable_vt):
            return True

    except Exception:
        pass

    return False


CONSOLE_SUPPORTS_COLOR = _configure_console_colors()
GLOBAL_OUTPUT_PREFERENCES = {"color": True, "emoji": True}


@dataclass
class ScanConfiguration:
    """Configuration for the scan session."""

    target_paths: List[str]
    database_path: Optional[str]
    is_recursive: bool
    include_git: bool
    include_node_modules: bool
    exclude_pattern: Optional[Pattern]
    show_full_report: bool
    use_json_output: bool
    disable_color: bool
    disable_emoji: bool


@dataclass(frozen=True)
class Vulnerability:
    """Represents a single infected package version."""

    package_name: str
    version: str


@dataclass
class ScanResult:
    """Result of scanning a single project directory."""

    project_path: str
    lockfiles: List[str]
    infected_packages: List[Vulnerability]

    @property
    def infected_count(self) -> int:
        return len(self.infected_packages)

    def to_dictionary(self) -> Dict[str, Any]:
        return {
            "project": self.project_path,
            "lockfiles": self.lockfiles,
            "infected_count": self.infected_count,
            "infected_packages": [
                {"name": v.package_name, "version": v.version}
                for v in self.infected_packages
            ],
        }


# --- Helper Functions ---


def get_color_sequence(color_code: str) -> str:
    """Returns the ANSI color sequence if enabled."""
    if not CONSOLE_SUPPORTS_COLOR or not GLOBAL_OUTPUT_PREFERENCES["color"]:
        return ""

    return color_code


def styled_text(text: str, color_code: str) -> str:
    """Wraps text in ANSI color codes if enabled."""
    prefix = get_color_sequence(color_code)
    if not prefix:
        return text

    suffix = get_color_sequence(TerminalColors.RESET)
    return f"{prefix}{text}{suffix}"


def get_emoji_icon(symbol: str) -> str:
    """Returns the emoji if enabled."""
    if not GLOBAL_OUTPUT_PREFERENCES["emoji"]:
        return ""

    return symbol


def emoji_text(symbol: str, text: str) -> str:
    """Prefixes text with an emoji if enabled."""
    glyph = get_emoji_icon(symbol)

    if glyph:
        return f"{glyph} {text}"

    return text


def deduplicate_vulnerabilities(items: Iterable[Vulnerability]) -> List[Vulnerability]:
    """Return vulnerabilities with duplicates removed while preserving order."""
    seen_items: Set[Tuple[str, str]] = set()
    unique_items: List[Vulnerability] = []

    for vulnerability in items:
        key = (vulnerability.package_name, vulnerability.version)
        if key not in seen_items:
            seen_items.add(key)
            unique_items.append(vulnerability)

    return unique_items


# --- Database ---


class VulnerabilityDatabase:
    """Singleton-like class to manage the database of infected packages."""

    def __init__(self):
        self._database_content: Dict[str, Set[str]] = defaultdict(set)
        self._loaded_path: Optional[str] = None
        self._entry_count: int = 0

    def load_database(
        self, database_path: Optional[str], quiet_mode: bool = False
    ) -> None:
        """Loads the infected packages list into memory for O(1) lookups."""
        if not database_path:
            # Default to looking in the script's directory
            script_directory = os.path.dirname(os.path.abspath(__file__))
            database_path = os.path.join(script_directory, "database.txt")

        if self._loaded_path == database_path:
            return

        if not os.path.exists(database_path):
            if not quiet_mode:
                print(
                    styled_text(
                        f"Error: Infected database not found at: {database_path}",
                        TerminalColors.RED,
                    ),
                    file=sys.stderr,
                )
                print(
                    styled_text(
                        "Please provide a valid path using --database or ensure database.txt exists.",
                        TerminalColors.YELLOW,
                    ),
                    file=sys.stderr,
                )
            raise FileNotFoundError(database_path)

        try:
            count = 0
            new_database: Dict[str, Set[str]] = defaultdict(set)

            with open(database_path, "r", encoding="utf-8") as file_handle:
                for line in file_handle:
                    line = line.strip()
                    if not line or line.startswith("#"):
                        continue

                    if ":" in line:
                        package_name, version = line.split(":", 1)
                        new_database[package_name].add(version.strip())
                        count += 1

            self._database_content = new_database
            self._loaded_path = database_path
            self._entry_count = count

            if not quiet_mode:
                print(
                    styled_text(
                        f"Loaded {count} infected package versions from {os.path.basename(database_path)}.",
                        TerminalColors.BLUE,
                    ),
                    file=sys.stderr,
                )

        except Exception as exception:
            if not quiet_mode:
                print(
                    styled_text(
                        f"Error loading infected packages: {exception}",
                        TerminalColors.RED,
                    ),
                    file=sys.stderr,
                )
            raise

    def is_package_infected(self, package_name: str, version: str) -> bool:
        """Checks if a specific package version is in the infected database."""
        return (
            package_name in self._database_content
            and version in self._database_content[package_name]
        )


# --- Parsers ---


class LockfileParser:
    """Base class for lockfile parsers with shared error handling."""

    @classmethod
    def parse_file(
        cls, file_path: str, database: "VulnerabilityDatabase"
    ) -> List[Vulnerability]:
        try:
            with open(file_path, "r", encoding="utf-8") as file_handle:
                return cls._parse_content(file_handle, database)
        except Exception as exception:
            cls._log_parse_error(file_path, exception)
            return []

    @classmethod
    def _parse_content(
        cls, file_handle: TextIO, database: "VulnerabilityDatabase"
    ) -> List[Vulnerability]:
        raise NotImplementedError

    @staticmethod
    def _log_parse_error(file_path: str, error: Exception) -> None:
        warning_message = f"Warning: Unable to parse {file_path}: {error}"
        print(styled_text(warning_message, TerminalColors.YELLOW), file=sys.stderr)


class NpmParser(LockfileParser):
    """Parser for package-lock.json (JSON)."""

    @classmethod
    def _parse_content(
        cls, file_handle: TextIO, database: "VulnerabilityDatabase"
    ) -> List[Vulnerability]:
        identified_issues: List[Vulnerability] = []
        json_data = json.load(file_handle)

        # Check direct dependencies
        dependencies = json_data.get("dependencies") or {}
        for package_name, package_details in dependencies.items():
            version = package_details.get("version", "")
            if database.is_package_infected(package_name, version):
                identified_issues.append(Vulnerability(package_name, version))

        # Check packages (npm v2/v3)
        packages = json_data.get("packages") or {}
        for package_path, package_details in packages.items():
            if not package_path:
                continue  # Root package entry

            package_name = package_path.rsplit("node_modules/", 1)[-1]
            version = package_details.get("version", "")

            if database.is_package_infected(package_name, version):
                identified_issues.append(Vulnerability(package_name, version))

        return identified_issues


class YarnParser(LockfileParser):
    """Parser for yarn.lock (Custom Format)."""

    _PATTERN = re.compile(
        r'^[\'"]?(@?[^@"\s\']+)@.+?[\'"]?:\s*(?:\r?\n|\r)\s*version(?:\s+|:\s+)["\']?([^"\s\']+)["\']?',
        re.MULTILINE,
    )

    @classmethod
    def _parse_content(
        cls, file_handle: TextIO, database: "VulnerabilityDatabase"
    ) -> List[Vulnerability]:
        identified_issues: List[Vulnerability] = []
        line_buffer: List[str] = []

        for line in file_handle:
            if not line.strip():
                cls._process_buffer(line_buffer, identified_issues, database)
            else:
                line_buffer.append(line)

        cls._process_buffer(line_buffer, identified_issues, database)
        return identified_issues

    @classmethod
    def _process_buffer(
        cls,
        buffer: List[str],
        issues_list: List[Vulnerability],
        database: "VulnerabilityDatabase",
    ) -> None:
        if not buffer:
            return

        block_content = "".join(buffer)
        for match in cls._PATTERN.finditer(block_content):
            package_name = match.group(1)
            version = match.group(2)

            if database.is_package_infected(package_name, version):
                issues_list.append(Vulnerability(package_name, version))

        buffer.clear()


class PnpmParser(LockfileParser):
    """Parser for pnpm-lock.yaml (YAML-like)."""

    _IGNORED_KEYS = {
        "resolution",
        "engines",
        "os",
        "cpu",
        "peerDependencies",
        "dependencies",
        "optionalDependencies",
        "devDependencies",
        "transitivePeerDependencies",
        "dev",
        "hasBin",
        "requiresBuild",
        "name",
        "version",
        "lockfileVersion",
        "settings",
        "importers",
        "packages",
        "specifiers",
        "patchedDependencies",
    }
    _REGEX_KEYS = re.compile(r'^\s+[\'"]?/?([^:\'"\s]+)[\'"]?:', re.MULTILINE)

    @classmethod
    def _parse_content(
        cls, file_handle: TextIO, database: "VulnerabilityDatabase"
    ) -> List[Vulnerability]:
        identified_issues: List[Vulnerability] = []

        for line in file_handle:
            match = cls._REGEX_KEYS.match(line)
            if not match:
                continue

            key = match.group(1)
            if key in cls._IGNORED_KEYS:
                continue

            coordinates = cls._extract_coordinates(key)
            if not coordinates:
                continue

            package_name, version = coordinates
            if database.is_package_infected(package_name, version):
                identified_issues.append(Vulnerability(package_name, version))

        return identified_issues

    @staticmethod
    def _extract_coordinates(key: str) -> Optional[Tuple[str, str]]:
        """Best-effort extraction of name/version pairs from pnpm lock entries."""
        if "(" in key:
            key = key.split("(", 1)[0]

        if "/" not in key and "@" not in key:
            return None

        package_name = None
        version = None

        last_at_index = key.rfind("@")
        if last_at_index > 0:
            package_name = key[:last_at_index]
            version = key[last_at_index + 1 :]
        else:
            last_slash_index = key.rfind("/")
            if last_slash_index > 0:
                package_name = key[:last_slash_index]
                version = key[last_slash_index + 1 :]

        if version and "_" in version:
            version = version.split("_", 1)[0]

        if package_name and version:
            return package_name, version

        return None


# --- Scanner Engine ---


class ScannerEngine:
    """Orchestrates the scanning process using parallel execution."""

    LOCKFILE_PARSER_MAP = {
        "package-lock.json": NpmParser,
        "yarn.lock": YarnParser,
        "pnpm-lock.yaml": PnpmParser,
    }

    def __init__(
        self, configuration: ScanConfiguration, database: "VulnerabilityDatabase"
    ):
        self.configuration = configuration
        self.database = database

        # Concurrency tuning
        base_threads = max(4, os.cpu_count() or 4)
        self._base_future_limit = base_threads
        self._future_limit = base_threads * 2
        self._max_future_limit = base_threads * 8

    def _should_skip_directory(self, root_path: str, base_path: str) -> bool:
        if (
            self.configuration.exclude_pattern
            and self.configuration.exclude_pattern.search(root_path)
        ):
            return True

        if not self.configuration.is_recursive and root_path != base_path:
            return True

        return False

    def _filter_directories(self, directories: List[str]) -> None:
        if not self.configuration.include_node_modules:
            directories[:] = [d for d in directories if d.lower() != "node_modules"]

        if not self.configuration.include_git:
            directories[:] = [d for d in directories if d.lower() != ".git"]

    def _scan_lockfile(
        self, directory_path: str, lockfile_name: str
    ) -> Tuple[str, str, List[Vulnerability]]:
        parser_class = self.LOCKFILE_PARSER_MAP[lockfile_name]
        full_path = os.path.join(directory_path, lockfile_name)
        issues = parser_class.parse_file(full_path, self.database)
        return directory_path, lockfile_name, issues

    def _adjust_future_limit(self, task_duration: float) -> None:
        if task_duration < 0.05 and self._future_limit < self._max_future_limit:
            self._future_limit = min(self._future_limit * 2, self._max_future_limit)
        elif task_duration > 0.25 and self._future_limit > self._base_future_limit:
            self._future_limit = max(self._base_future_limit, self._future_limit // 2)

    def _drain_futures(
        self,
        pending_futures: Dict[concurrent.futures.Future, float],
        project_data: Dict[str, Dict[str, Any]],
        project_order: List[str],
        wait_for_all: bool = False,
    ) -> None:
        if not pending_futures:
            return

        wait_condition = (
            concurrent.futures.ALL_COMPLETED
            if wait_for_all
            else concurrent.futures.FIRST_COMPLETED
        )
        done_futures, _ = concurrent.futures.wait(
            list(pending_futures.keys()), return_when=wait_condition
        )

        for future in done_futures:
            start_time = pending_futures.pop(future, time.monotonic())
            duration = max(0.0, time.monotonic() - start_time)
            self._adjust_future_limit(duration)

            try:
                directory, lockfile, issues = future.result()
            except Exception:
                continue

            if directory not in project_data:
                project_data[directory] = {"lockfiles": [], "issues": []}
                project_order.append(directory)

            entry = project_data[directory]
            if lockfile not in entry["lockfiles"]:
                entry["lockfiles"].append(lockfile)

            if issues:
                entry["issues"].extend(issues)

    def _submit_scan_task(
        self,
        directory_path: str,
        lockfile_name: str,
        executor: concurrent.futures.ThreadPoolExecutor,
        pending_futures: Dict[concurrent.futures.Future, float],
        project_data: Dict[str, Dict[str, Any]],
        project_order: List[str],
    ) -> None:
        future = executor.submit(self._scan_lockfile, directory_path, lockfile_name)
        pending_futures[future] = time.monotonic()

        if len(pending_futures) >= self._future_limit:
            self._drain_futures(
                pending_futures, project_data, project_order, wait_for_all=False
            )

    def execute_scan(self) -> List[ScanResult]:
        """Executes the scan based on configuration."""
        project_data: Dict[str, Dict[str, Any]] = {}
        project_order: List[str] = []
        pending_futures: Dict[concurrent.futures.Future, float] = {}

        with concurrent.futures.ThreadPoolExecutor() as executor:
            for input_path in self.configuration.target_paths:
                absolute_path = os.path.abspath(input_path)
                if not os.path.exists(absolute_path):
                    continue

                # Case 1: Input is a file
                if os.path.isfile(absolute_path):
                    file_name = os.path.basename(absolute_path)
                    directory_path = os.path.dirname(absolute_path) or "."
                    if file_name in self.LOCKFILE_PARSER_MAP:
                        self._submit_scan_task(
                            directory_path,
                            file_name,
                            executor,
                            pending_futures,
                            project_data,
                            project_order,
                        )
                    continue

                # Case 2: Input is a directory (walk it)
                for root, dirs, files in os.walk(absolute_path):
                    if self._should_skip_directory(root, absolute_path):
                        dirs[:] = []
                        continue

                    self._filter_directories(dirs)
                    if not self.configuration.is_recursive:
                        dirs[:] = []

                    for file_name in files:
                        if file_name in self.LOCKFILE_PARSER_MAP:
                            self._submit_scan_task(
                                root,
                                file_name,
                                executor,
                                pending_futures,
                                project_data,
                                project_order,
                            )

        self._drain_futures(
            pending_futures, project_data, project_order, wait_for_all=True
        )

        final_results: List[ScanResult] = []
        for directory in project_order:
            entry = project_data.get(directory)
            if not entry:
                continue

            lockfiles = entry["lockfiles"]
            if not lockfiles:
                continue

            issues = deduplicate_vulnerabilities(entry["issues"])
            final_results.append(ScanResult(directory, lockfiles, issues))

        return final_results


# --- Reporting ---


def print_cli_report(
    results: List[ScanResult],
    configuration: ScanConfiguration,
    arguments_dictionary: Dict[str, Any],
):
    """Prints the human-readable CLI report."""
    display_results, summary = summarize_scan_results(
        results, configuration.show_full_report
    )

    # 1. Args and Options
    print(
        "\n" + styled_text(emoji_text("🔍", "Scan Configuration"), TerminalColors.BLUE)
    )
    print(styled_text("─────────────────────", TerminalColors.BLUE))

    for key, value in arguments_dictionary.items():
        if isinstance(value, bool):
            display_value = "Yes" if value else "No"
        elif isinstance(value, list):
            display_value = ", ".join(str(entry) for entry in value) if value else "-"
        elif value is None:
            display_value = "-"
        else:
            display_value = value
        print(f"{key:<22}: {display_value}")
    print("")

    # 2. Per Project Report
    print(styled_text(emoji_text("🚀", "Project Reports"), TerminalColors.BLUE))
    print(styled_text("──────────────────", TerminalColors.BLUE))

    for result in display_results:
        count = result.infected_count
        if count > 0:
            print(
                "\n"
                + styled_text(
                    emoji_text("🚫", f"[INFECTED] {result.project_path}"),
                    TerminalColors.RED,
                )
            )
            print(f"   {emoji_text('📄', 'Lockfiles:')} {', '.join(result.lockfiles)}")
            print(f"   {emoji_text('🦠', 'Infected Packages:')} {count}")

            for vulnerability in result.infected_packages:
                print(f"      - {vulnerability.package_name}@{vulnerability.version}")
        else:
            print(
                "\n"
                + styled_text(
                    emoji_text("✅", f"[CLEAN]    {result.project_path}"),
                    TerminalColors.GREEN,
                )
            )

    print("")

    # 3. Global Report
    print(styled_text(emoji_text("📊", "Global Summary"), TerminalColors.BLUE))
    print(styled_text("─────────────────", TerminalColors.BLUE))
    print(f"Total Projects: {summary['total_projects']}")
    print(f"Infected:       {summary['infected_projects']}")
    print(f"Clean:          {summary['total_projects'] - summary['infected_projects']}")
    print(f"Total Issues:   {summary['total_infected_packages']}")

    print("")
    if summary["total_projects"] == 0:
        print(styled_text(emoji_text("⚠️", "No lockfile found"), TerminalColors.YELLOW))
        sys.exit(1)
    elif summary["infected_projects"] == 0:
        print(
            styled_text(
                emoji_text("🎉", "No project infected. You are safe!"),
                TerminalColors.GREEN,
            )
        )
        sys.exit(0)
    else:
        print(
            styled_text(
                emoji_text(
                    "❌", f"Found {summary['infected_projects']} infected projects!"
                ),
                TerminalColors.RED,
            )
        )
        sys.exit(1)


def print_json_report(
    results: List[ScanResult],
    configuration: ScanConfiguration,
    arguments_dictionary: Dict[str, Any],
):
    """Prints the JSON report."""
    display_results, summary = summarize_scan_results(
        results, configuration.show_full_report
    )
    output_data = {
        "configuration": arguments_dictionary,
        "results": [r.to_dictionary() for r in display_results],
        "summary": summary,
    }
    print(json.dumps(output_data, indent=2))

    if summary["total_projects"] == 0:
        sys.exit(1)
    elif summary["infected_projects"] > 0:
        sys.exit(1)
    else:
        sys.exit(0)


def summarize_scan_results(
    results: List[ScanResult], show_full_report: bool
) -> Tuple[List[ScanResult], Dict[str, int]]:
    total_projects = len(results)
    infected_projects = sum(1 for r in results if r.infected_count > 0)
    total_infected_packages = sum(r.infected_count for r in results)

    if show_full_report:
        display_results = list(results)
    else:
        display_results = [r for r in results if r.infected_count > 0]

    display_results.sort(key=lambda r: (r.infected_count == 0, r.project_path.lower()))

    summary = {
        "total_projects": total_projects,
        "infected_projects": infected_projects,
        "total_infected_packages": total_infected_packages,
    }
    return display_results, summary


def build_output_configurations(
    parsed_arguments: argparse.Namespace,
) -> Tuple[Dict[str, Any], Dict[str, Any]]:
    """Constructs both CLI display and JSON configuration dictionaries."""
    json_config = {
        "paths": parsed_arguments.paths,
        "database": (
            parsed_arguments.database if parsed_arguments.database else "Default"
        ),
        "recursive": parsed_arguments.recursive,
        "include_git": parsed_arguments.include_git,
        "include_node_modules": parsed_arguments.include_node_modules,
        "exclude_regex": parsed_arguments.exclude,
        "full_report": parsed_arguments.full_report,
        "json_output": parsed_arguments.json,
        "color_output": not parsed_arguments.no_color,
        "emoji_output": not parsed_arguments.no_emoji,
    }

    display_config = {
        "Paths": json_config["paths"],
        "Database": json_config["database"],
        "Recursive": json_config["recursive"],
        "Include .git": json_config["include_git"],
        "Include node_modules": json_config["include_node_modules"],
        "Exclude Regex": json_config["exclude_regex"],
        "Full Report": json_config["full_report"],
        "JSON Output": json_config["json_output"],
        "Color Output": json_config["color_output"],
        "Emoji Output": json_config["emoji_output"],
    }

    return display_config, json_config


def main():
    """Main entry point for the application."""
    parser = argparse.ArgumentParser(
        description="Fast Lockfile Scanner for Infected Packages"
    )
    parser.add_argument(
        "paths", nargs="*", default=["."], help="Directories or files to scan"
    )

    scan_group = parser.add_argument_group("Scan Options")
    scan_group.add_argument(
        "-r", "--recursive", action="store_true", help="Scan directories recursively"
    )
    scan_group.add_argument(
        "-g",
        "--include-git",
        action="store_true",
        help="Include .git directories during recursion",
    )
    scan_group.add_argument(
        "-n",
        "--include-node-modules",
        action="store_true",
        help="Include node_modules directories during recursion",
    )
    scan_group.add_argument(
        "-e", "--exclude", type=str, help="Exclude paths matching this regex"
    )

    output_group = parser.add_argument_group("Output Options")
    output_group.add_argument(
        "-f",
        "--full-report",
        action="store_true",
        help="Display projects that are not infected",
    )
    output_group.add_argument(
        "-j", "--json", action="store_true", help="Output results in JSON format"
    )
    output_group.add_argument(
        "-C",
        "--no-color",
        action="store_true",
        help="Disable ANSI colors in the CLI report",
    )
    output_group.add_argument(
        "-E",
        "--no-emoji",
        action="store_true",
        help="Disable emoji icons in the CLI report",
    )

    advanced_group = parser.add_argument_group("Advanced")
    advanced_group.add_argument(
        "-d", "--database", type=str, help="Path to database.txt database file"
    )

    parsed_arguments = parser.parse_args()

    if parsed_arguments.no_color:
        GLOBAL_OUTPUT_PREFERENCES["color"] = False
    if parsed_arguments.no_emoji:
        GLOBAL_OUTPUT_PREFERENCES["emoji"] = False

    # Create Configuration Object
    configuration = ScanConfiguration(
        target_paths=parsed_arguments.paths,
        database_path=parsed_arguments.database,
        is_recursive=parsed_arguments.recursive,
        include_git=parsed_arguments.include_git,
        include_node_modules=parsed_arguments.include_node_modules,
        exclude_pattern=(
            re.compile(parsed_arguments.exclude) if parsed_arguments.exclude else None
        ),
        show_full_report=parsed_arguments.full_report,
        use_json_output=parsed_arguments.json,
        disable_color=parsed_arguments.no_color,
        disable_emoji=parsed_arguments.no_emoji,
    )

    # Initialize Database
    database = VulnerabilityDatabase()
    try:
        database.load_database(
            parsed_arguments.database, quiet_mode=parsed_arguments.json
        )
    except Exception:
        sys.exit(1)

    # Run Scan
    engine = ScannerEngine(configuration, database)
    scan_results = engine.execute_scan()

    # Output
    display_config, json_config = build_output_configurations(parsed_arguments)
    if parsed_arguments.json:
        print_json_report(scan_results, configuration, json_config)
    else:
        print_cli_report(scan_results, configuration, display_config)


if __name__ == "__main__":
    main()
