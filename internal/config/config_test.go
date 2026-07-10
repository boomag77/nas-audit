package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := writeConfig(t, `
server:
  listen_address: "127.0.0.1:8080"
database:
  driver: sqlite
  path: ./var/nas-audit.db
scanner:
  batch_size: 1000
  follow_symlinks: false
roots:
  - name: local-test
    path: ./testdata/sample-root
    read_only_required: false
    exclusions:
      - .DS_Store
reports:
  directory: ./reports
  git_commit_enabled: false
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Driver != "sqlite" {
		t.Fatalf("database driver = %q, want sqlite", cfg.Database.Driver)
	}
	if len(cfg.Roots) != 1 {
		t.Fatalf("roots len = %d, want 1", len(cfg.Roots))
	}
	if cfg.Roots[0].Name != "local-test" {
		t.Fatalf("root name = %q, want local-test", cfg.Roots[0].Name)
	}
}

func TestValidateRejectsUnsupportedDatabaseDriver(t *testing.T) {
	cfg := validConfig()
	cfg.Database.Driver = "postgres"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "unsupported database.driver") {
		t.Fatalf("Validate() error = %v, want unsupported database.driver", err)
	}
}

func TestValidateRejectsEmptyDatabasePath(t *testing.T) {
	cfg := validConfig()
	cfg.Database.Path = ""

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "database.path") {
		t.Fatalf("Validate() error = %v, want database.path error", err)
	}
}

func TestValidateRejectsDuplicateRootNames(t *testing.T) {
	cfg := validConfig()
	cfg.Roots = append(cfg.Roots, cfg.Roots[0])

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "duplicate root name") {
		t.Fatalf("Validate() error = %v, want duplicate root name", err)
	}
}

func TestValidateRejectsEmptyRootPath(t *testing.T) {
	cfg := validConfig()
	cfg.Roots[0].Path = ""

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "roots[0].path") {
		t.Fatalf("Validate() error = %v, want roots[0].path error", err)
	}
}

func TestValidateRejectsInvalidBatchSize(t *testing.T) {
	cfg := validConfig()
	cfg.Scanner.BatchSize = 0

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "scanner.batch_size") {
		t.Fatalf("Validate() error = %v, want scanner.batch_size error", err)
	}
}

func validConfig() Config {
	return Config{
		Server: ServerConfig{
			ListenAddress: "127.0.0.1:8080",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			Path:   "./var/nas-audit.db",
		},
		Scanner: ScannerConfig{
			BatchSize:      1000,
			FollowSymlinks: false,
		},
		Roots: []RootConfig{
			{
				Name:             "local-test",
				Path:             "./testdata/sample-root",
				ReadOnlyRequired: false,
				Exclusions:       []string{".DS_Store"},
			},
		},
		Reports: ReportsConfig{
			Directory:        "./reports",
			GitCommitEnabled: false,
		},
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}
