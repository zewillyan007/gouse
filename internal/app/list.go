// Package app holds use cases for the gouse CLI. Functions in this package
// are framework-agnostic: they accept plain Go types and return data/errors,
// without importing cobra or touching stdout/stderr. The cli/ package adapts
// these to user-facing commands.
package app

import (
	"context"
	"time"

	"github.com/zewillyan007/gouse/internal/releases"
	"github.com/zewillyan007/gouse/internal/state"
	"github.com/zewillyan007/gouse/internal/store"
)

type InstalledVersion struct {
	Version string
	Latest  bool
	Default bool
}

type ListResult struct {
	Versions []InstalledVersion
}

// List returns the installed versions, marking `(default)` and — when a
// cached or freshly fetched latest is available — `(latest)` on the version
// that matches the newest stable Go release published online.
//
// Network access: List tries to refresh the latest cache when it's empty or
// older than the TTL, with a short timeout. Failure is silent — the listing
// still works offline using whatever cache exists.
func List(ctx context.Context) (ListResult, error) {
	versions, err := store.ListInstalled()
	if err != nil {
		return ListResult{}, err
	}
	st, err := state.Load()
	if err != nil {
		return ListResult{}, err
	}

	latestKnown, fresh := state.LatestFresh()
	if !fresh {
		if refreshed, ok := tryRefreshLatest(ctx); ok {
			latestKnown = refreshed
		}
	}

	out := make([]InstalledVersion, 0, len(versions))
	for _, v := range versions {
		out = append(out, InstalledVersion{
			Version: v,
			Latest:  latestKnown != "" && v == latestKnown,
			Default: v == st.Default,
		})
	}
	return ListResult{Versions: out}, nil
}

func tryRefreshLatest(parent context.Context) (string, bool) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	rs, err := releases.Fetch(ctx)
	if err != nil {
		return "", false
	}
	latest, err := releases.LatestStable(rs)
	if err != nil {
		return "", false
	}
	_ = state.UpdateLatest(latest.Version)
	return latest.Version, true
}

// pickLatest returns the highest stable version among the given names. RC/beta
// versions never win the (latest) tag — they sort below their final release.
// Used by `gouse use latest` to resolve against installed versions.
func pickLatest(versions []string) string {
	var best string
	for _, v := range versions {
		if compareVersions(v, best) > 0 {
			best = v
		}
	}
	return best
}
