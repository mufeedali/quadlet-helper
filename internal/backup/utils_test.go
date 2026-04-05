package backup

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestResticEnvIncludesExpectedVariables(t *testing.T) {
	config := &Config{
		Options: Options{PasswordFile: "/tmp/password"},
		Destination: Destination{
			Repository: "s3:bucket/repo",
		},
		Environment: []string{"CUSTOM=1"},
	}

	env := ResticEnv(config)
	checks := []string{
		"CUSTOM=1",
		"RESTIC_PASSWORD_FILE=/tmp/password",
		"RESTIC_REPOSITORY=s3:bucket/repo",
	}

	for _, check := range checks {
		if !slices.Contains(env, check) {
			t.Fatalf("ResticEnv() missing %q in %v", check, env)
		}
	}
}

func TestRcloneDestPath(t *testing.T) {
	if got := RcloneDestPath("remote:backup", "/srv/data", 1); got != "remote:backup" {
		t.Fatalf("RcloneDestPath(single) = %q", got)
	}
	if got := RcloneDestPath("remote:backup", "/srv/data", 2); got != "remote:backup/data" {
		t.Fatalf("RcloneDestPath(multi) = %q", got)
	}
}

func TestRsyncArgs(t *testing.T) {
	config := &Config{
		Source: []string{"/src"},
		Destination: Destination{
			Path: "/dest",
		},
		Options: Options{
			Archive:  true,
			Compress: true,
			Delete:   true,
			Exclude:  []string{"*.tmp"},
		},
	}

	got := RsyncArgs(config, true)
	want := []string{"--dry-run", "-a", "-z", "--delete", "-v", "--progress", "--exclude", "*.tmp", "/src", "/dest"}
	if !slices.Equal(got, want) {
		t.Fatalf("RsyncArgs() = %v, want %v", got, want)
	}
}

func TestRsyncDestPath(t *testing.T) {
	t.Run("single source without trailing slash", func(t *testing.T) {
		destDir := t.TempDir()
		got := RsyncDestPath(destDir, "/srv/data", 1)
		want := filepath.Join(destDir, "data")
		if got != want {
			t.Fatalf("RsyncDestPath() = %q, want %q", got, want)
		}
	})

	t.Run("single source with trailing slash", func(t *testing.T) {
		destDir := t.TempDir()
		got := RsyncDestPath(destDir, "/srv/data/", 1)
		if got != destDir {
			t.Fatalf("RsyncDestPath() = %q, want %q", got, destDir)
		}
	})

	t.Run("single destination file", func(t *testing.T) {
		destDir := t.TempDir()
		destFile := filepath.Join(destDir, "backup.tar")
		if err := os.WriteFile(destFile, []byte("demo"), 0600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got := RsyncDestPath(destFile, "/srv/data", 1)
		if got != destFile {
			t.Fatalf("RsyncDestPath() = %q, want %q", got, destFile)
		}
	})

	t.Run("multiple sources", func(t *testing.T) {
		destDir := t.TempDir()
		got := RsyncDestPath(destDir, "/srv/data", 2)
		want := filepath.Join(destDir, "data")
		if got != want {
			t.Fatalf("RsyncDestPath() = %q, want %q", got, want)
		}
	})
}

func TestRcloneBaseArgs(t *testing.T) {
	config := &Config{
		Options: Options{
			Transfers:      4,
			Checkers:       8,
			BandwidthLimit: "10M",
			Exclude:        []string{"*.bak"},
		},
	}

	got := RcloneBaseArgs(config, true)
	want := []string{"sync", "--dry-run", "--transfers", "4", "--checkers", "8", "--bwlimit", "10M", "--exclude", "*.bak", "-v", "--progress"}
	if !slices.Equal(got, want) {
		t.Fatalf("RcloneBaseArgs() = %v, want %v", got, want)
	}
}
