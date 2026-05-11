package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
	"github.com/zewillyan007/gouse/internal/shell"
)

func newInitCmd() *cobra.Command {
	var shellFlag string
	cmd := &cobra.Command{
		Use:    "init",
		Short:  "Gera ~/.gouse/shell.<sh|zsh|fish> e completion (executado pelo install.sh)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := app.Init(app.InitParams{ShellName: shellFlag})
			if err != nil {
				return err
			}

			completionPath := filepath.Join(filepath.Dir(res.IntegrationFile), res.Shell.CompletionFilename())
			if err := writeCompletion(cmd.Root(), res.Shell, completionPath); err != nil {
				return fmt.Errorf("falha ao gerar completion para %s: %w", res.ShellName, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s gerado.\n", res.IntegrationFile)
			fmt.Fprintf(cmd.OutOrStdout(), "%s gerado.\n", completionPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Shell detectado: %s\n", res.ShellName)
			fmt.Fprintln(cmd.OutOrStdout(), "Adicione ao seu rcfile (se ainda não estiver lá):")
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", res.Shell.SourceLine())
			return nil
		},
	}
	cmd.Flags().StringVar(&shellFlag, "shell", "", "shell-alvo (bash|zsh|fish). Vazio = autodetect via $SHELL")
	return cmd
}

// writeCompletion writes the Cobra-generated completion script for `sh`
// to `dest`. Dispatches between GenBashCompletionV2 / GenZshCompletion /
// GenFishCompletion. Adding a new shell requires updating this switch.
func writeCompletion(root *cobra.Command, sh shell.Shell, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	switch sh.Name() {
	case "bash":
		return root.GenBashCompletionV2(f, true)
	case "zsh":
		return root.GenZshCompletion(f)
	case "fish":
		return root.GenFishCompletion(f, true)
	default:
		return fmt.Errorf("completion não suportada para shell %q", sh.Name())
	}
}
