// Package store manages on-disk layout: $XDG_DATA_HOME/gos (or
// ~/.local/share/gos) for Go installations and ~/go/gopaths for per-version
// GOPATHs. All operations are user-owned — no sudo required.
package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// GosDir returns the absolute path of the directory that holds Go
// installations. Honors $XDG_DATA_HOME when set; otherwise ~/.local/share/gos.
func GosDir() (string, error) {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "gos"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "gos"), nil
}

func GopathsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "go", "gopaths"), nil
}

func VersionDir(version string) (string, error) {
	root, err := GosDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, version), nil
}

func GoBinDir(version string) (string, error) {
	dir, err := VersionDir(version)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "go", "bin"), nil
}

func GopathFor(version string) (string, error) {
	root, err := GopathsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, version), nil
}

// ListInstalled returns the sorted list of version names present in GosDir.
// A directory counts as a valid install when it contains go/bin/go.
func ListInstalled() ([]string, error) {
	gos, err := GosDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(gos)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, statErr := os.Stat(filepath.Join(gos, e.Name(), "go", "bin", "go")); statErr == nil {
			versions = append(versions, e.Name())
		}
	}
	sort.Strings(versions)
	return versions, nil
}

func Exists(version string) bool {
	bin, err := GoBinDir(version)
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(bin, "go"))
	return err == nil
}

// EnsureGosDir creates GosDir if missing.
func EnsureGosDir() error {
	gos, err := GosDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(gos, 0o755)
}

// RemoveVersion removes <GosDir>/<version>.
func RemoveVersion(version string) error {
	dir, err := VersionDir(version)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

// EnsureGopath creates ~/go/gopaths/<version> if missing.
func EnsureGopath(version string) error {
	p, err := GopathFor(version)
	if err != nil {
		return err
	}
	return os.MkdirAll(p, 0o755)
}

// RemoveGopath removes ~/go/gopaths/<version>.
func RemoveGopath(version string) error {
	p, err := GopathFor(version)
	if err != nil {
		return err
	}
	return os.RemoveAll(p)
}
