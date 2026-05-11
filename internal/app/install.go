package app

import (
	"context"
	"fmt"

	"github.com/zewillyan007/gouse/internal/installer"
	"github.com/zewillyan007/gouse/internal/platform"
	"github.com/zewillyan007/gouse/internal/releases"
	"github.com/zewillyan007/gouse/internal/state"
	"github.com/zewillyan007/gouse/internal/store"
)

type InstallParams struct {
	Version      string // "latest" resolves to the newest stable release
	OnSourceURL  func(url string)
	Progress     func(read, total int64)
	SetAsDefault bool // when true, persist the installed version as default
}

type InstallResult struct {
	Version      string // resolved (concrete) version that was installed
	SetAsDefault bool   // true when this run also persisted the version as default
}

func Install(ctx context.Context, p InstallParams) (InstallResult, error) {
	plat := platform.Detect()
	if err := plat.Check(); err != nil {
		return InstallResult{}, err
	}
	rs, err := releases.Fetch(ctx)
	if err != nil {
		return InstallResult{}, err
	}

	// Best-effort: refresh the cached latest stable while we have fresh data.
	if latest, latestErr := releases.LatestStable(rs); latestErr == nil {
		_ = state.UpdateLatest(latest.Version)
	}

	var rel releases.Release
	if p.Version == "latest" {
		rel, err = releases.LatestStable(rs)
	} else {
		rel, err = releases.Find(rs, p.Version)
	}
	if err != nil {
		return InstallResult{}, err
	}

	if store.Exists(rel.Version) {
		return InstallResult{Version: rel.Version}, fmt.Errorf("versão %s já está instalada", rel.Version)
	}

	file, err := platform.SelectFile(rel.Files, plat)
	if err != nil {
		return InstallResult{}, err
	}

	if err := installer.Install(ctx, file, p.OnSourceURL, p.Progress); err != nil {
		return InstallResult{}, err
	}
	if err := store.EnsureGopath(rel.Version); err != nil {
		return InstallResult{Version: rel.Version}, fmt.Errorf("Go instalado mas falha ao criar GOPATH: %w", err)
	}
	result := InstallResult{Version: rel.Version}
	if p.SetAsDefault {
		if err := SetDefault(rel.Version); err != nil {
			return result, fmt.Errorf("instalado, mas falha ao gravar default: %w", err)
		}
		result.SetAsDefault = true
	}
	return result, nil
}
