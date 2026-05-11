package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zewillyan007/gouse/internal/store"
)

func init() {
	register(fish{})
}

type fish struct{}

func (fish) Name() string { return "fish" }

func (fish) RCFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "fish", "config.fish"), nil
}

func (fish) SourceLine() string { return "source ~/.gouse/shell.fish" }

func (fish) IntegrationPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gouse", "shell.fish"), nil
}

func (fish) CompletionFilename() string { return "completion.fish" }

const fishTemplate = `# ~/.gouse/shell.fish — gerado por ` + "`gouse init`" + `. Não editar manualmente.
fish_add_path -g $HOME/.gouse/bin

# Nota: ` + "`| string collect`" + ` é essencial. Sem ele, a substituição de
# comando quebra a saída em uma lista de linhas e ` + "`eval`" + ` junta os
# argumentos com espaço, destruindo as quebras de linha do script renderizado.

function gouse
  switch $argv[1]
    case use
      set -l out (command gouse shell-env --shell fish $argv[2] | string collect)
      if test $pipestatus[1] -ne 0
        return $pipestatus[1]
      end
      eval $out
    case '*'
      command gouse $argv
  end
end

# Tab-complete (gerado por ` + "`gouse init`" + `):
if test -f ~/.gouse/completion.fish
  source ~/.gouse/completion.fish
end

# Ativa a versão default ao carregar (se houver):
set -l _default (command gouse default --print 2>/dev/null)
if test -n "$_default"
  set -l out (command gouse shell-env --shell fish $_default 2>/dev/null | string collect)
  if test $pipestatus[1] -eq 0
    eval $out
  end
end
`

func (f fish) WriteIntegration() (string, error) {
	p, err := f.IntegrationPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(fishTemplate), 0o644); err != nil {
		return "", err
	}
	return p, nil
}

// RenderEnv emits fish-native assignments via `set -gx`. PATH is cleaned
// of the previous version's entries using fish list filtering.
func (fish) RenderEnv(version string) (string, error) {
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
	prevGoBinTpl := gosDir + "/$GO_CURRENT/go/bin"
	prevGopathTpl := gopathsRoot + "/$GO_CURRENT/bin"

	var b strings.Builder
	// Remove the previous version's bins from $PATH if GO_CURRENT differs.
	fmt.Fprintf(&b, "if set -q GO_CURRENT; and test \"$GO_CURRENT\" != %q\n", version)
	fmt.Fprintf(&b, "  set -l prev_go_bin %s\n", quoteFish(prevGoBinTpl))
	fmt.Fprintf(&b, "  set -l prev_gp_bin %s\n", quoteFish(prevGopathTpl))
	b.WriteString("  set PATH (string match -v -- $prev_go_bin $PATH)\n")
	b.WriteString("  set PATH (string match -v -- $prev_gp_bin $PATH)\n")
	b.WriteString("end\n")
	fmt.Fprintf(&b, "set -gx GOPATH %s\n", quoteFish(gopath))
	fmt.Fprintf(&b, "set -gx PATH %s %s $PATH\n", quoteFish(goBin), quoteFish(gopathBin))
	fmt.Fprintf(&b, "set -gx GO_CURRENT %s\n", quoteFish(version))
	return b.String(), nil
}

// quoteFish wraps a string in double quotes, escaping characters that have
// special meaning in fish double-quoted strings: $ \ " (and backtick is not
// special in fish).
func quoteFish(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
