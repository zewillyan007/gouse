package shell

import (
	"fmt"
	"os"
	"strings"
)

var registry = map[string]Shell{}

func register(s Shell) {
	registry[s.Name()] = s
}

// Names returns the registered shell names in stable order.
func Names() []string {
	out := make([]string, 0, len(registry))
	for _, name := range []string{"bash", "zsh", "fish"} {
		if _, ok := registry[name]; ok {
			out = append(out, name)
		}
	}
	return out
}

// Get returns the shell implementation by canonical name.
func Get(name string) (Shell, error) {
	s, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("shell %q desconhecido (suportados: %v)", name, Names())
	}
	return s, nil
}

// Detect inspects $SHELL and returns the matching shell implementation.
// Unknown shells fall back to bash.
func Detect() Shell {
	sh := os.Getenv("SHELL")
	switch {
	case strings.HasSuffix(sh, "/zsh") || sh == "zsh":
		if s, ok := registry["zsh"]; ok {
			return s
		}
	case strings.HasSuffix(sh, "/fish") || sh == "fish":
		if s, ok := registry["fish"]; ok {
			return s
		}
	}
	return registry["bash"]
}
