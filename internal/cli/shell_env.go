package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newShellEnvCmd() *cobra.Command {
	var shellFlag string
	cmd := &cobra.Command{
		Use:    "shell-env <versão|latest>",
		Short:  "Imprime exports para aplicar uma versão no shell (uso interno)",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := app.ShellEnv(app.ShellEnvParams{
				Version:   args[0],
				ShellName: shellFlag,
			})
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), res.Script)
			return nil
		},
	}
	cmd.Flags().StringVar(&shellFlag, "shell", "bash", "sintaxe a emitir (bash|zsh|fish)")
	return cmd
}
