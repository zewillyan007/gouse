package app

import (
	"fmt"

	"github.com/zewillyan007/gouse/internal/state"
	"github.com/zewillyan007/gouse/internal/store"
)

func SetDefault(version string) error {
	if !store.Exists(version) {
		return fmt.Errorf("versão %s não está instalada", version)
	}
	st, err := state.Load()
	if err != nil {
		return err
	}
	st.Default = version
	return state.Save(st)
}

func GetDefault() (string, error) {
	st, err := state.Load()
	if err != nil {
		return "", err
	}
	return st.Default, nil
}
