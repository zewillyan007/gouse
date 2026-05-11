// Package platform identifies the current OS/arch and selects the matching
// release file from the go.dev/dl payload. Adding a new platform is a
// one-line append to the `supported` slice.
package platform

import (
	"fmt"
	"runtime"
)

// Platform represents a target combination of OS, architecture, and the
// archive format used by Go's download page.
type Platform struct {
	OS            string
	Arch          string
	ArchiveFormat string // "tar.gz" | "zip"
}

func (p Platform) String() string {
	return p.OS + "/" + p.Arch
}

// supported is the canonical list of platforms gouse can install Go onto.
// Add new entries here; installer.Extract must understand the ArchiveFormat.
var supported = []Platform{
	{OS: "linux", Arch: "amd64", ArchiveFormat: "tar.gz"},
	{OS: "linux", Arch: "arm64", ArchiveFormat: "tar.gz"},
}

// Supported returns a copy of the supported platforms list.
func Supported() []Platform {
	out := make([]Platform, len(supported))
	copy(out, supported)
	return out
}

// Detect returns the platform of the running process. Use Resolve to map it
// to the canonical entry (with ArchiveFormat) or get an error.
func Detect() Platform {
	return Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
}

// Resolve looks up the supported entry matching p.OS/p.Arch and returns
// it (which carries the ArchiveFormat). Returns an error if not supported.
func Resolve(p Platform) (Platform, error) {
	for _, s := range supported {
		if s.OS == p.OS && s.Arch == p.Arch {
			return s, nil
		}
	}
	return Platform{}, fmt.Errorf("plataforma %s não suportada nesta versão do gouse", p)
}

// Check is a shortcut for Resolve that discards the canonical entry.
func (p Platform) Check() error {
	_, err := Resolve(p)
	return err
}

type File interface {
	GetOS() string
	GetArch() string
	GetKind() string
}

// SelectFile picks the archive file matching the platform from a list of
// release files (e.g. releases.File).
func SelectFile[F File](files []F, p Platform) (F, error) {
	var zero F
	for _, f := range files {
		if f.GetKind() == "archive" && f.GetOS() == p.OS && f.GetArch() == p.Arch {
			return f, nil
		}
	}
	return zero, fmt.Errorf("nenhum arquivo archive para %s", p)
}
