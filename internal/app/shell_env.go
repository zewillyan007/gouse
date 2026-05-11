package app

import (
	"fmt"

	"github.com/zewillyan007/gouse/internal/shell"
	"github.com/zewillyan007/gouse/internal/store"
)

type ShellEnvParams struct {
	Version   string // "latest" resolves to the newest installed version
	ShellName string // canonical shell name; empty = bash (POSIX exports)
}

type ShellEnvResult struct {
	Version string
	Script  string
}

func ShellEnv(p ShellEnvParams) (ShellEnvResult, error) {
	version := p.Version
	if version == "latest" {
		installed, err := store.ListInstalled()
		if err != nil {
			return ShellEnvResult{}, err
		}
		version = pickLatest(installed)
		if version == "" {
			return ShellEnvResult{}, fmt.Errorf("nenhuma versão instalada")
		}
	}
	if !store.Exists(version) {
		return ShellEnvResult{}, fmt.Errorf("versão %s não está instalada; rode `gouse install %s`", version, version)
	}
	name := p.ShellName
	if name == "" {
		name = "bash"
	}
	sh, err := shell.Get(name)
	if err != nil {
		return ShellEnvResult{}, err
	}
	script, err := sh.RenderEnv(version)
	if err != nil {
		return ShellEnvResult{}, err
	}
	return ShellEnvResult{Version: version, Script: script}, nil
}
