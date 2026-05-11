package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zewillyan007/gouse/internal/store"
)

func init() {
	register(bash{})
}

type bash struct{}

func (bash) Name() string { return "bash" }

func (bash) RCFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".bashrc"), nil
}

func (bash) SourceLine() string { return "source ~/.gouse/shell.sh" }

func (bash) IntegrationPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gouse", "shell.sh"), nil
}

func (bash) CompletionFilename() string { return "completion.bash" }

const bashTemplate = `# ~/.gouse/shell.sh — gerado por ` + "`gouse init`" + `. Não editar manualmente.
export PATH="$HOME/.gouse/bin:$PATH"

gouse() {
  case "$1" in
    use)
      local out
      out=$(command gouse shell-env --shell bash "$2") || return $?
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
  [ -n "$_default" ] && eval "$(command gouse shell-env --shell bash "$_default" 2>/dev/null)"
fi
unset _default
`

func (b bash) WriteIntegration() (string, error) {
	p, err := b.IntegrationPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(bashTemplate), 0o644); err != nil {
		return "", err
	}
	return p, nil
}

// RenderEnv emits POSIX-shell exports. The output is also valid in zsh.
func (bash) RenderEnv(version string) (string, error) {
	return renderPOSIX(version)
}

// renderPOSIX is shared by bash and zsh.
func renderPOSIX(version string) (string, error) {
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
