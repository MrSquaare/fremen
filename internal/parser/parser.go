package parser

import (
	"github.com/MrSquaare/fremen/internal/database"
)

type Vulnerability struct {
	PackageName string `json:"name"`
	Version     string `json:"version"`
}

type LockfileParserFunc func(path string, db *database.VulnerabilityDatabase) ([]Vulnerability, error)

var LockfileParsers = map[string]LockfileParserFunc{
	"package-lock.json": parseNpmLockfile,
	"yarn.lock":         parseYarnLockfile,
	"pnpm-lock.yaml":    parsePnpmLockfile,
}
