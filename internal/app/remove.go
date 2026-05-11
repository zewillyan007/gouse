package app

import (
	"fmt"

	"github.com/zewillyan007/gouse/internal/state"
	"github.com/zewillyan007/gouse/internal/store"
)

type RemoveParams struct {
	Version string
	Force   bool // allow removing the configured default
	Purge   bool // also remove ~/go/gopaths/<version>
}

type RemoveResult struct {
	RemovedGopath bool
}

func Remove(p RemoveParams) (RemoveResult, error) {
	if !store.Exists(p.Version) {
		return RemoveResult{}, fmt.Errorf("versão %s não está instalada", p.Version)
	}
	st, err := state.Load()
	if err != nil {
		return RemoveResult{}, err
	}
	if st.Default == p.Version && !p.Force {
		return RemoveResult{}, fmt.Errorf("versão %s é a default; rode novamente com --force ou troque o default primeiro", p.Version)
	}
	if err := store.RemoveVersion(p.Version); err != nil {
		return RemoveResult{}, err
	}
	if st.Default == p.Version {
		st.Default = ""
		if err := state.Save(st); err != nil {
			return RemoveResult{}, fmt.Errorf("versão removida, mas falha ao limpar default: %w", err)
		}
	}
	result := RemoveResult{}
	if p.Purge {
		if err := store.RemoveGopath(p.Version); err != nil {
			return result, fmt.Errorf("versão removida, mas falha ao apagar GOPATH: %w", err)
		}
		result.RemovedGopath = true
	}
	return result, nil
}
