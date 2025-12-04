package scanner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/MrSquaare/fremen/internal/database"
	"github.com/MrSquaare/fremen/internal/parser"
)

// ScanConfig describes the configuration options for a scan operation.
type ScanConfig struct {
	TargetPaths  []string
	DatabasePath string

	Recursive          bool
	IncludeGit         bool
	IncludeNodeModules bool

	ExcludeRegex *regexp.Regexp
}

type scanTask struct {
	Dir      string
	Lockfile string
}

type scanResultItem struct {
	Dir      string
	Lockfile string
	Issues   []parser.Vulnerability
}

type projectEntry struct {
	Lockfiles []string
	Issues    []parser.Vulnerability
}

// ExecuteScan scans all target paths concurrently and aggregates results.
func ExecuteScan(cfg ScanConfig, db *database.VulnerabilityDatabase) ([]ScanResult, error) {
	// 1. collect tasks concurrently
	var tasks []scanTask
	var errs []error
	for _, p := range cfg.TargetPaths {
		found, err := collectTasksForPath(cfg, p)
		if err != nil {
			errs = append(errs, err)
		}
		tasks = append(tasks, found...)
	}

	if len(tasks) == 0 {
		return nil, errors.Join(errs...)
	}

	workers := workerCount()
	jobs := make(chan scanTask)
	jobResults := make(chan scanResultItem)

	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()
			for job := range jobs {
				parserFn, ok := parser.LockfileParsers[job.Lockfile]
				if !ok {
					continue
				}

				fullPath := filepath.Join(job.Dir, job.Lockfile)
				issues, err := parserFn(fullPath, db)
				if err != nil {
					errs = append(errs, fmt.Errorf("parse %s: %w", fullPath, err))
				}

				jobResults <- scanResultItem{
					Dir:      job.Dir,
					Lockfile: job.Lockfile,
					Issues:   issues,
				}
			}
		}()
	}

	go func() {
		for _, t := range tasks {
			jobs <- t
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(jobResults)
	}()

	// 2. aggregate scanResultItem per project
	scanResultItemsByProject := make(map[string]*projectEntry)
	var projects []string

	for jobResult := range jobResults {
		entry, exists := scanResultItemsByProject[jobResult.Dir]
		if !exists {
			entry = &projectEntry{}
			scanResultItemsByProject[jobResult.Dir] = entry
			projects = append(projects, jobResult.Dir)
		}

		if !slices.Contains(entry.Lockfiles, jobResult.Lockfile) {
			entry.Lockfiles = append(entry.Lockfiles, jobResult.Lockfile)
		}

		if len(jobResult.Issues) > 0 {
			entry.Issues = append(entry.Issues, jobResult.Issues...)
		}
	}

	// 3. deduplicate to ScanResult
	scanResults := make([]ScanResult, 0, len(projects))
	for _, projectDir := range projects {
		entry := scanResultItemsByProject[projectDir]
		if entry == nil || len(entry.Lockfiles) == 0 {
			continue
		}

		scanResults = append(scanResults, ScanResult{
			ProjectPath:      projectDir,
			Lockfiles:        entry.Lockfiles,
			InfectedPackages: deduplicateVulnerabilities(entry.Issues),
		})
	}

	return scanResults, errors.Join(errs...)
}

func isLockfile(name string) bool {
	_, ok := parser.LockfileParsers[name]
	return ok
}

func isExcludedDir(cfg ScanConfig, dirName string) bool {
	if !cfg.IncludeNodeModules && strings.EqualFold(dirName, "node_modules") {
		return true
	}
	if !cfg.IncludeGit && strings.EqualFold(dirName, ".git") {
		return true
	}
	return false
}

func workerCount() int {
	const minWorkers = 4

	n := runtime.NumCPU()
	if n < minWorkers {
		return minWorkers
	}
	return n
}

// collectTasksForPath finds lockfiles at or under a given input path.
func collectTasksForPath(cfg ScanConfig, path string) ([]scanTask, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path %q: %w", path, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", abs, err)
	}

	if !info.IsDir() {
		name := filepath.Base(abs)
		if isLockfile(name) {
			return []scanTask{{
				Dir:      filepath.Dir(abs),
				Lockfile: name,
			}}, nil
		}
		return nil, nil
	}

	base := abs
	tasks := make([]scanTask, 0)

	walkErrors := []error{}

	err = filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			walkErrors = append(walkErrors, fmt.Errorf("walk %q: %w", base, walkErr))
			return nil
		}

		if cfg.ExcludeRegex != nil && cfg.ExcludeRegex.MatchString(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if path != base && !cfg.Recursive {
				return filepath.SkipDir
			}
			if isExcludedDir(cfg, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if isLockfile(d.Name()) {
			tasks = append(tasks, scanTask{
				Dir:      filepath.Dir(path),
				Lockfile: d.Name(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(walkErrors) > 0 {
		return tasks, errors.Join(walkErrors...)
	}

	return tasks, nil
}

type vulnerabilityKey struct {
	Name    string
	Version string
}

func deduplicateVulnerabilities(items []parser.Vulnerability) []parser.Vulnerability {
	seen := make(map[vulnerabilityKey]struct{}, len(items))
	out := make([]parser.Vulnerability, 0, len(items))

	for _, v := range items {
		key := vulnerabilityKey{
			Name:    v.PackageName,
			Version: v.Version,
		}

		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}

	return out
}
