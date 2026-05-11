package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newShellEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "shell-env <versão|latest>",
		Short:  "Imprime exports para aplicar uma versão no shell (uso interno)",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := app.ShellEnv(app.ShellEnvParams{Version: args[0]})
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), res.Script)
			return nil
		},
	}
}
