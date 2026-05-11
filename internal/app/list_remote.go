package app

import (
	"context"
	"fmt"

	"github.com/zewillyan007/gouse/internal/releases"
	"github.com/zewillyan007/gouse/internal/state"
)

type RemoteVersion struct {
	Version string
	Stable  bool
	Latest  bool
}

type ListRemoteParams struct {
	All     bool // include RC/beta releases
	Page    int  // 1-indexed; page 1 = most recent slice
	PerPage int  // items per page; default 20
}

type ListRemoteResult struct {
	Versions   []RemoteVersion // ordered oldest → newest within this page
	Page       int
	TotalPages int
	TotalCount int
}

const defaultPerPage = 20

func ListRemote(ctx context.Context, p ListRemoteParams) (ListRemoteResult, error) {
	rs, err := releases.Fetch(ctx)
	if err != nil {
		return ListRemoteResult{}, err
	}

	latest, latestErr := releases.LatestStable(rs)
	if latestErr == nil {
		// Best-effort cache refresh — never fail the listing because of it.
		_ = state.UpdateLatest(latest.Version)
	}

	// Filter according to --all.
	filtered := make([]releases.Release, 0, len(rs))
	for _, r := range rs {
		if !p.All && !r.Stable {
			continue
		}
		filtered = append(filtered, r)
	}

	// API order is newest → oldest. Reverse to oldest → newest.
	reversed := make([]releases.Release, len(filtered))
	for i, r := range filtered {
		reversed[len(filtered)-1-i] = r
	}

	perPage := p.PerPage
	if perPage <= 0 {
		perPage = defaultPerPage
	}
	total := len(reversed)
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	if page > totalPages {
		return ListRemoteResult{Page: page, TotalPages: totalPages, TotalCount: total},
			fmt.Errorf("página %d inválida; total de páginas: %d", page, totalPages)
	}

	// page 1 = the newest perPage entries (last slice). page 2 = previous slice. etc.
	end := total - (page-1)*perPage
	start := end - perPage
	if start < 0 {
		start = 0
	}

	out := make([]RemoteVersion, 0, end-start)
	for _, r := range reversed[start:end] {
		out = append(out, RemoteVersion{
			Version: r.Version,
			Stable:  r.Stable,
			Latest:  latestErr == nil && r.Version == latest.Version,
		})
	}
	return ListRemoteResult{
		Versions:   out,
		Page:       page,
		TotalPages: totalPages,
		TotalCount: total,
	}, nil
}
