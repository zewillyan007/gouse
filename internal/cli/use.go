package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newUseCmd() *cobra.Command {
	var shellFlag string
	cmd := &cobra.Command{
		Use:   "use <versão|latest>",
		Short: "Troca a versão ativa no shell atual",
		Long: `Troca a versão ativa no shell atual.

Requer a função shell ` + "`gouse()`" + ` carregada (via ~/.gouse/shell.{sh|zsh|fish}).
Se o install.sh foi executado e o rcfile faz source desse arquivo,
basta rodar ` + "`gouse use <versão>`" + ` em qualquer terminal novo.

Sem a função carregada, este comando imprime os exports no stdout —
use ` + "`eval \"$(gouse use <versão>)\"`" + ` (ou ` + "`eval (gouse use <versão>)`" + ` no fish)
como alternativa manual.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: installedVersionCompletion,
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
