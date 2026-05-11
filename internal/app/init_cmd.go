package app

import "github.com/zewillyan007/gouse/internal/shell"

type InitResult struct {
	ShellPath string
}

// Init generates ~/.gouse/shell.sh. Idempotent — safe to call multiple times.
func Init() (InitResult, error) {
	p, err := shell.WriteShellSh()
	if err != nil {
		return InitResult{}, err
	}
	return InitResult{ShellPath: p}, nil
}
