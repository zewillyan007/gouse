package shell

import (
	"os"
	"path/filepath"
)

func init() {
	register(zsh{})
}

type zsh struct{}

func (zsh) Name() string { return "zsh" }

func (zsh) RCFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zshrc"), nil
}

func (zsh) SourceLine() string { return "source ~/.gouse/shell.zsh" }

func (zsh) IntegrationPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gouse", "shell.zsh"), nil
}

func (zsh) CompletionFilename() string { return "completion.zsh" }

const zshTemplate = `# ~/.gouse/shell.zsh — gerado por ` + "`gouse init`" + `. Não editar manualmente.
export PATH="$HOME/.gouse/bin:$PATH"

gouse() {
  case "$1" in
    use)
      local out
      out=$(command gouse shell-env --shell zsh "$2") || return $?
      eval "$out"
      ;;
    *)
      command gouse "$@"
      ;;
  esac
}

# Tab-complete (gerado por ` + "`gouse init`" + `):
[ -f ~/.gouse/completion.zsh ] && source ~/.gouse/completion.zsh

# Ativa a versão default ao carregar (se houver):
if _default=$(command gouse default --print 2>/dev/null); then
  [ -n "$_default" ] && eval "$(command gouse shell-env --shell zsh "$_default" 2>/dev/null)"
fi
unset _default
`

func (z zsh) WriteIntegration() (string, error) {
	p, err := z.IntegrationPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(zshTemplate), 0o644); err != nil {
		return "", err
	}
	return p, nil
}

// RenderEnv emits POSIX exports — zsh understands the same bash-style
// syntax, so we reuse renderPOSIX.
func (zsh) RenderEnv(version string) (string, error) {
	return renderPOSIX(version)
}
