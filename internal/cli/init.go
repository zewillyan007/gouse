package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "init",
		Short:  "Gera ~/.gouse/shell.sh e ~/.gouse/completion.bash (executado pelo install.sh)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := app.Init()
			if err != nil {
				return err
			}

			// Generate ~/.gouse/completion.bash next to shell.sh, using this
			// binary's command tree. Cobra is already imported here, so the
			// shell package stays framework-agnostic.
			completionPath := filepath.Join(filepath.Dir(res.ShellPath), "completion.bash")
			f, err := os.Create(completionPath)
			if err != nil {
				return fmt.Errorf("falha ao criar %s: %w", completionPath, err)
			}
			if err := cmd.Root().GenBashCompletionV2(f, true); err != nil {
				f.Close()
				return fmt.Errorf("falha ao gerar completion: %w", err)
			}
			if err := f.Close(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s gerado.\n", res.ShellPath)
			fmt.Fprintf(cmd.OutOrStdout(), "%s gerado.\n", completionPath)
			fmt.Fprintln(cmd.OutOrStdout(), "Adicione ao seu ~/.bashrc (se ainda não estiver lá):")
			fmt.Fprintln(cmd.OutOrStdout(), "  source ~/.gouse/shell.sh")
			return nil
		},
	}
}
