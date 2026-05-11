package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
	"github.com/zewillyan007/gouse/internal/store"
)

func newInstallCmd() *cobra.Command {
	var setDefault bool
	cmd := &cobra.Command{
		Use:   "install <versão|latest>",
		Short: "Baixa e instala uma versão do Go",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "Instalando %s...\n", version)
			res, err := app.Install(cmd.Context(), app.InstallParams{
				Version:      version,
				OnSourceURL:  func(url string) { fmt.Fprintf(os.Stderr, "Baixando %s...\n", url) },
				Progress:     makeProgress(),
				SetAsDefault: setDefault,
			})
			if err != nil {
				if res.Version != "" {
					return fmt.Errorf("%s: %w", res.Version, err)
				}
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout())
			if dir, derr := store.VersionDir(res.Version); derr == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%s instalado em %s/go\n", res.Version, dir)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%s instalado.\n", res.Version)
			}
			if res.SetAsDefault {
				fmt.Fprintf(cmd.OutOrStdout(), "Default agora é %s. Próximos shells aplicarão automaticamente.\n", res.Version)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Use com: `gouse use %s`\n", res.Version)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&setDefault, "default", "d", false, "grava esta versão como default (aplicada em novos shells)")
	return cmd
}

func makeProgress() func(read, total int64) {
	if !isTerminal(os.Stderr) {
		return nil
	}
	last := -1
	return func(read, total int64) {
		var pct int
		if total > 0 {
			pct = int(read * 100 / total)
		}
		if pct == last {
			return
		}
		last = pct
		if total > 0 {
			fmt.Fprintf(os.Stderr, "\rDownload: %d%% (%.1f MB / %.1f MB)", pct, mb(read), mb(total))
		} else {
			fmt.Fprintf(os.Stderr, "\rDownload: %.1f MB", mb(read))
		}
	}
}

func mb(b int64) float64 { return float64(b) / 1024 / 1024 }
