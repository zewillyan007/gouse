// Package cli wires Cobra to the framework-agnostic use cases in
// internal/app. It is the only package allowed to import cobra. Swapping
// frameworks means rewriting just this package and cmd/gouse/main.go.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "gouse",
		Version:       version,
		Short:         "Gerenciador de versões do Go",
		Long:          "gouse instala, remove e troca entre versões do Go no Linux.\nUse `gouse --help-all` para ver comandos avançados.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	var helpAll bool
	root.PersistentFlags().BoolVar(&helpAll, "help-all", false, "mostra todos os comandos, incluindo os avançados")
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if helpAll {
			revealHidden(cmd.Root())
		}
		defaultHelp(cmd, args)
	})
	// When `--help-all` is passed without `--help`, Cobra still runs the
	// command. Root has no Run, so display help in that case too.
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if helpAll {
			revealHidden(cmd.Root())
		}
		return cmd.Help()
	}
	root.AddCommand(
		newListCmd(),
		newListRemoteCmd(),
		newInstallCmd(),
		newRemoveCmd(),
		newUseCmd(),
		newDefaultCmd(),
		newInitCmd(),
		newShellEnvCmd(),
	)
	return root
}

func revealHidden(c *cobra.Command) {
	for _, sub := range c.Commands() {
		sub.Hidden = false
		revealHidden(sub)
	}
}

// Execute runs the gouse CLI. Returns an exit code suitable for os.Exit.
func Execute(version string) int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	root := newRootCmd(version)
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, formatError(err))
		return 1
	}
	return 0
}

func formatError(err error) string {
	return "erro: " + err.Error()
}
