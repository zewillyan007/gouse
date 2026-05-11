package cli

import (
	"github.com/spf13/cobra"
	"github.com/zewillyan007/gouse/internal/store"
)

// installedVersionCompletion suggests version names from the local install
// directory. It accepts only one positional argument; further completions
// return nothing.
func installedVersionCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	versions, _ := store.ListInstalled()
	return versions, cobra.ShellCompDirectiveNoFileComp
}
