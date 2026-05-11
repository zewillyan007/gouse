// Package platform identifies the current OS/arch and selects the matching
// release file from the go.dev/dl payload. New OS/arch combinations only need
// to be added to Supported (and to installer.Extract if their archive format
// differs from tar.gz).
package platform

import (
	"fmt"
	"runtime"
)

type Platform struct {
	OS   string
	Arch string
}

func (p Platform) String() string {
	return p.OS + "/" + p.Arch
}

var Supported = map[string]bool{
	"linux/amd64": true,
}

func Detect() Platform {
	return Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

func (p Platform) Check() error {
	if !Supported[p.String()] {
		return fmt.Errorf("plataforma %s não suportada nesta versão do gouse", p)
	}
	return nil
}

type File interface {
	GetOS() string
	GetArch() string
	GetKind() string
}

func SelectFile[F File](files []F, p Platform) (F, error) {
	var zero F
	for _, f := range files {
		if f.GetKind() == "archive" && f.GetOS() == p.OS && f.GetArch() == p.Arch {
			return f, nil
		}
	}
	return zero, fmt.Errorf("nenhum arquivo archive para %s", p)
}
