package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetEnv_Exists(t *testing.T) {
	t.Setenv("TEST_GETENV_EXISTS", "hello")
	if got := getEnv("TEST_GETENV_EXISTS", "fallback"); got != "hello" {
		t.Errorf("getEnv() = %q, want %q", got, "hello")
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("TEST_GETENV_MISSING")
	if got := getEnv("TEST_GETENV_MISSING", "fallback"); got != "fallback" {
		t.Errorf("getEnv() = %q, want %q", got, "fallback")
	}
}

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `# Comment line
FOO=bar
BAZ=qux
  SPACED  =  value  `
	os.WriteFile(envFile, []byte(content), 0644)

	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	loadEnvFile()

	tests := map[string]string{
		"FOO":     "bar",
		"BAZ":     "qux",
		"SPACED":  "value",
	}
	for key, want := range tests {
		if got := os.Getenv(key); got != want {
			t.Errorf("env %q = %q, want %q", key, got, want)
		}
	}
}

func TestLoadEnvFile_NoFile(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	loadEnvFile()
}

func TestLoadEnvFile_EmptyLines(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("\n\n  \n# comment\nKEY=val"), 0644)

	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	loadEnvFile()

	if got := os.Getenv("KEY"); got != "val" {
		t.Errorf("env KEY = %q, want %q", got, "val")
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("MIN_AGE_DAYS")
	os.Unsetenv("NEXUS_URL")
	os.Unsetenv("NPM_REGISTRY")
	os.Unsetenv("LISTEN_PORT")

	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	cfg := Load()

	if cfg.ListenPort != ":8080" {
		t.Errorf("ListenPort = %q, want %q", cfg.ListenPort, ":8080")
	}
	if cfg.MinAgeDays != 7 {
		t.Errorf("MinAgeDays = %v, want %v", cfg.MinAgeDays, 7)
	}
	if cfg.NexusURL != "http://localhost:8081/repository/npm-proxy/" {
		t.Errorf("NexusURL = %q, want %q", cfg.NexusURL, "http://localhost:8081/repository/npm-proxy/")
	}
	if cfg.NPMRegistry != "https://registry.npmjs.org/" {
		t.Errorf("NPMRegistry = %q, want %q", cfg.NPMRegistry, "https://registry.npmjs.org/")
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("LISTEN_PORT", ":3000")
	t.Setenv("MIN_AGE_DAYS", "14")
	t.Setenv("NEXUS_URL", "http://custom:8081/")
	t.Setenv("NPM_REGISTRY", "http://custom-registry/")

	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(orig)

	cfg := Load()

	if cfg.ListenPort != ":3000" {
		t.Errorf("ListenPort = %q, want %q", cfg.ListenPort, ":3000")
	}
	if cfg.MinAgeDays != 14 {
		t.Errorf("MinAgeDays = %v, want %v", cfg.MinAgeDays, 14)
	}
	if cfg.NexusURL != "http://custom:8081/" {
		t.Errorf("NexusURL = %q, want %q", cfg.NexusURL, "http://custom:8081/")
	}
	if cfg.NPMRegistry != "http://custom-registry/" {
		t.Errorf("NPMRegistry = %q, want %q", cfg.NPMRegistry, "http://custom-registry/")
	}
}

func TestLoad_InvalidMinAgeDays(t *testing.T) {
	if os.Getenv("TEST_LOAD_INVALID") == "1" {
		t.Setenv("MIN_AGE_DAYS", "notanumber")

		dir := t.TempDir()
		os.Chdir(dir)

		Load()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoad_InvalidMinAgeDays")
	cmd.Env = append(os.Environ(), "TEST_LOAD_INVALID=1")
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else if err != nil {
		t.Errorf("expected exit error, got %v", err)
	}
}
