// Package state persists gouse user state in ~/.gouse/state.json.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const latestTTL = 24 * time.Hour

type State struct {
	Default         string `json:"default,omitempty"`
	LatestKnown     string `json:"latest_known,omitempty"`
	LatestCheckedAt int64  `json:"latest_checked_at,omitempty"` // unix seconds
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gouse", "state.json"), nil
}

func Load() (State, error) {
	p, err := Path()
	if err != nil {
		return State{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("falha ao ler %s: %w", p, err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("state.json inválido: %w", err)
	}
	return s, nil
}

// UpdateLatest records the newest stable Go version known to the system.
func UpdateLatest(version string) error {
	s, err := Load()
	if err != nil {
		return err
	}
	s.LatestKnown = version
	s.LatestCheckedAt = time.Now().Unix()
	return Save(s)
}

// LatestFresh returns the cached latest version and whether it was checked
// less than latestTTL ago. An empty string means no cache.
func LatestFresh() (string, bool) {
	s, err := Load()
	if err != nil || s.LatestKnown == "" {
		return "", false
	}
	age := time.Since(time.Unix(s.LatestCheckedAt, 0))
	return s.LatestKnown, age < latestTTL
}

func Save(s State) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
