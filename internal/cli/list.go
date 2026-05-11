package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista versões do Go instaladas",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := app.List(cmd.Context())
			if err != nil {
				return err
			}
			if len(res.Versions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Nenhuma versão instalada. Rode `gouse install latest`.")
				return nil
			}
			for _, v := range res.Versions {
				line := v.Version
				if v.Latest {
					line += " (latest)"
				}
				if v.Default {
					line += " (default)"
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}
}
