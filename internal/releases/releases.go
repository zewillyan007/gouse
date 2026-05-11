// Package releases fetches and exposes the list of Go releases from the
// official go.dev/dl endpoint.
package releases

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const endpoint = "https://go.dev/dl/?mode=json&include=all"

type File struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

func (f File) GetOS() string   { return f.OS }
func (f File) GetArch() string { return f.Arch }
func (f File) GetKind() string { return f.Kind }

type Release struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []File `json:"files"`
}

func Fetch(ctx context.Context) ([]Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("resposta %d de %s: %s", resp.StatusCode, endpoint, body)
	}
	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON: %w", err)
	}
	return releases, nil
}

// FilterStable returns only stable releases, preserving order.
func FilterStable(rs []Release) []Release {
	out := make([]Release, 0, len(rs))
	for _, r := range rs {
		if r.Stable {
			out = append(out, r)
		}
	}
	return out
}

// LatestStable returns the first stable release (the API returns them
// newest-first).
func LatestStable(rs []Release) (Release, error) {
	for _, r := range rs {
		if r.Stable {
			return r, nil
		}
	}
	return Release{}, fmt.Errorf("nenhuma versão estável encontrada na resposta da API")
}

// Find locates a release by exact version name (e.g. "go1.26.2").
func Find(rs []Release, version string) (Release, error) {
	for _, r := range rs {
		if r.Version == version {
			return r, nil
		}
	}
	return Release{}, fmt.Errorf("versão %s não encontrada na lista de releases", version)
}
