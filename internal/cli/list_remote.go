package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/app"
)

func newListRemoteCmd() *cobra.Command {
	var all bool
	var page, perPage int
	cmd := &cobra.Command{
		Use:   "list-remote",
		Short: "Lista versões do Go disponíveis para download",
		Long: `Lista versões disponíveis na API oficial.

Mostra 20 versões por página, ordenadas da mais antiga para a mais nova.
A página 1 (default) contém as 20 mais recentes — a (latest) fica na
última linha. Use --page <N> para versões mais antigas.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := app.ListRemote(cmd.Context(), app.ListRemoteParams{
				All:     all,
				Page:    page,
				PerPage: perPage,
			})
			if err != nil {
				return err
			}
			for _, v := range res.Versions {
				line := v.Version
				if v.Latest {
					line += " (latest)"
				} else if !v.Stable {
					line += " (unstable)"
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			if res.TotalPages > 1 {
				fmt.Fprintf(cmd.OutOrStdout(),
					"\nPágina %d/%d. Use --page <N> para versões mais antigas.\n",
					res.Page, res.TotalPages)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "inclui release candidates e betas")
	cmd.Flags().IntVar(&page, "page", 1, "página (1 = mais recente)")
	cmd.Flags().IntVar(&perPage, "per-page", 20, "versões por página")
	_ = cmd.Flags().MarkHidden("per-page")
	return cmd
}
