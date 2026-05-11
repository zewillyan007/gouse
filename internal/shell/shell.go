// Package shell generates the bash integration file (~/.gouse/shell.sh) and
// renders the `export` block consumed by `gouse shell-env <version>`.
package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zewillyan007/gouse/internal/store"
)

// shellTemplate is written verbatim to ~/.gouse/shell.sh by `gouse init`.
// It defines the gouse() function and activates the default version on load.
const shellTemplate = `# ~/.gouse/shell.sh — gerado por ` + "`gouse init`" + `. Não editar manualmente.
export PATH="$HOME/.gouse/bin:$PATH"

gouse() {
  case "$1" in
    use)
      local out
      out=$(command gouse shell-env "$2") || return $?
      eval "$out"
      ;;
    *)
      command gouse "$@"
      ;;
  esac
}

# Tab-complete (gerado por ` + "`gouse init`" + `):
[ -f ~/.gouse/completion.bash ] && source ~/.gouse/completion.bash

# Ativa a versão default ao carregar (se houver):
if _default=$(command gouse default --print 2>/dev/null); then
  [ -n "$_default" ] && eval "$(command gouse shell-env "$_default" 2>/dev/null)"
fi
unset _default
`

func ShellPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gouse", "shell.sh"), nil
}

// WriteShellSh creates ~/.gouse/shell.sh (idempotent).
func WriteShellSh() (string, error) {
	p, err := ShellPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(shellTemplate), 0o644); err != nil {
		return "", err
	}
	return p, nil
}

// RenderEnv returns the `export` block that switches the current shell to
// the given version. It strips the previous version's entries from PATH
// using $GO_CURRENT, then prepends the new go bin + gopath bin.
//
// Paths (GosDir, GOPATH base) are resolved here, at render time, against the
// caller's environment. The resulting script only relies on $GO_CURRENT and
// $PATH being set.
func RenderEnv(version string) (string, error) {
	gosDir, err := store.GosDir()
	if err != nil {
		return "", err
	}
	gopathsRoot, err := store.GopathsDir()
	if err != nil {
		return "", err
	}
	goBin := filepath.Join(gosDir, version, "go", "bin")
	gopath := filepath.Join(gopathsRoot, version)
	gopathBin := filepath.Join(gopath, "bin")
	prevGoBinPrefix := gosDir + "/"
	prevGopathPrefix := gopathsRoot + "/"
	var b strings.Builder
	fmt.Fprintf(&b, "if [ -n \"$GO_CURRENT\" ] && [ \"$GO_CURRENT\" != %q ]; then\n", version)
	b.WriteString("  PATH=$(printf '%s' \"$PATH\" | tr ':' '\\n' \\\n")
	fmt.Fprintf(&b, "    | grep -vFx %q \\\n", prevGoBinPrefix+"$GO_CURRENT/go/bin")
	fmt.Fprintf(&b, "    | grep -vFx %q \\\n", prevGopathPrefix+"$GO_CURRENT/bin")
	b.WriteString("    | paste -sd:)\n")
	b.WriteString("fi\n")
	fmt.Fprintf(&b, "export GOPATH=%q\n", gopath)
	fmt.Fprintf(&b, "export PATH=%q:%q:\"$PATH\"\n", goBin, gopathBin)
	fmt.Fprintf(&b, "export GO_CURRENT=%q\n", version)
	return b.String(), nil
}
