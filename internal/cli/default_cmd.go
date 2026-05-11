package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newDefaultCmd() *cobra.Command {
	var printOnly bool
	cmd := &cobra.Command{
		Use:   "default [versão]",
		Short: "Define ou exibe a versão padrão",
		Long: `Sem argumentos, exibe a versão default atual.
Com uma versão, grava como default — aplicada em novos shells.

` + "`--print`" + ` (uso interno): imprime apenas o valor (vazio se não houver).`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: installedVersionCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || printOnly {
				current, err := app.GetDefault()
				if err != nil {
					return err
				}
				if printOnly {
					if current != "" {
						fmt.Fprintln(cmd.OutOrStdout(), current)
					}
					return nil
				}
				if current == "" {
					fmt.Fprintln(cmd.OutOrStdout(), "Nenhuma versão default definida.")
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Default: %s\n", current)
				}
				return nil
			}
			if err := app.SetDefault(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Default agora é %s. Próximos shells aplicarão automaticamente.\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&printOnly, "print", false, "imprime apenas o valor atual (uso interno do shell.sh)")
	return cmd
}
