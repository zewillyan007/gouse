package app

import "github.com/zewillyan007/gouse/internal/shell"

type InitParams struct {
	// ShellName overrides $SHELL autodetection. Empty means "detect".
	ShellName string
}

type InitResult struct {
	ShellName     string // canonical name of the shell selected
	IntegrationFile string // path written (e.g. ~/.gouse/shell.sh)
	Shell         shell.Shell // resolved shell, used by the CLI layer to render completion
}

// Init generates the shell integration file for the chosen shell. The CLI
// layer is responsible for additionally writing the completion file (via
// cobra) next to the integration script.
func Init(p InitParams) (InitResult, error) {
	var sh shell.Shell
	if p.ShellName == "" {
		sh = shell.Detect()
	} else {
		s, err := shell.Get(p.ShellName)
		if err != nil {
			return InitResult{}, err
		}
		sh = s
	}
	path, err := sh.WriteIntegration()
	if err != nil {
		return InitResult{}, err
	}
	return InitResult{
		ShellName:       sh.Name(),
		IntegrationFile: path,
		Shell:           sh,
	}, nil
}
