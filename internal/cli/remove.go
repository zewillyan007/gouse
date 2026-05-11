package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newRemoveCmd() *cobra.Command {
	var force, purge bool
	cmd := &cobra.Command{
		Use:               "remove <versão>",
		Short:             "Remove uma versão instalada",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: installedVersionCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			res, err := app.Remove(app.RemoveParams{Version: version, Force: force, Purge: purge})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s removido.\n", version)
			if res.RemovedGopath {
				fmt.Fprintf(cmd.OutOrStdout(), "GOPATH ~/go/gopaths/%s removido.\n", version)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "permite remover a versão default")
	cmd.Flags().BoolVar(&purge, "purge", false, "também apaga ~/go/gopaths/<versão>")
	return cmd
}
