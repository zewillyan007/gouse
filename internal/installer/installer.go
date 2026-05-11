// Package installer downloads a Go release archive, verifies its SHA256
// checksum, and extracts it into the target version directory.
package installer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zewillyan007/gouse/internal/platform"
	"github.com/zewillyan007/gouse/internal/releases"
	"github.com/zewillyan007/gouse/internal/store"
)

const downloadBase = "https://go.dev/dl/"

// Install downloads the given release file, checks SHA256, and extracts it
// into <store.GosDir()>/<version>/. The Go tarball already contains a
// top-level `go/` directory, so the resulting layout is
// <store.GosDir()>/<version>/go/...
//
// onSourceURL, if non-nil, is called once with the full URL before the
// download starts (for transparency).
// progress, if non-nil, is called with (bytesRead, totalBytes) during the
// download.
func Install(ctx context.Context, file releases.File, plat platform.Platform, onSourceURL func(url string), progress func(read, total int64)) error {
	if err := store.EnsureGosDir(); err != nil {
		return err
	}
	target, err := store.VersionDir(file.Version)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("falha ao criar %s: %w", target, err)
	}

	tmp, err := os.CreateTemp(target, "download-*.tar.gz")
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo temporário: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	url := downloadBase + file.Filename
	if onSourceURL != nil {
		onSourceURL(url)
	}
	if err := download(ctx, url, tmp, file.Size, progress); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		return err
	}

	got, err := sha256sum(tmp)
	if err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()
	if !strings.EqualFold(got, file.SHA256) {
		return fmt.Errorf("SHA256 não confere: esperado %s, obtido %s", file.SHA256, got)
	}

	if err := extract(tmpPath, target, plat.ArchiveFormat); err != nil {
		// best-effort cleanup on extraction failure
		_ = os.RemoveAll(target)
		return err
	}
	return nil
}

// extract dispatches by archive format. Only tar.gz is implemented today;
// zip will be added when Windows support lands.
func extract(archivePath, destDir, format string) error {
	switch format {
	case "tar.gz":
		return extractTarGz(archivePath, destDir)
	case "zip":
		return fmt.Errorf("formato zip ainda não suportado (TODO: Windows)")
	default:
		return fmt.Errorf("formato de arquivo desconhecido: %s", format)
	}
}

func download(ctx context.Context, url string, w io.Writer, expectedSize int64, progress func(read, total int64)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("falha no download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d ao baixar %s", resp.StatusCode, url)
	}

	total := expectedSize
	if total <= 0 {
		total = resp.ContentLength
	}
	var reader io.Reader = resp.Body
	if progress != nil {
		reader = &progressReader{r: resp.Body, total: total, cb: progress}
	}
	if _, err := io.Copy(w, reader); err != nil {
		return fmt.Errorf("falha durante a transferência: %w", err)
	}
	return nil
}

type progressReader struct {
	r     io.Reader
	read  int64
	total int64
	cb    func(read, total int64)
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	if n > 0 {
		p.read += int64(n)
		p.cb(p.read, p.total)
	}
	return n, err
}

func sha256sum(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("falha ao calcular SHA256: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("falha ao abrir gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("falha ao ler tar: %w", err)
		}
		// strip nothing: tarball has a top-level "go/" directory which we keep,
		// resulting in <destDir>/go/...
		cleaned := filepath.Clean(hdr.Name)
		if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
			return fmt.Errorf("entrada suspeita no tar: %s", hdr.Name)
		}
		target := filepath.Join(destDir, cleaned)
		// Defensive check against tar path traversal.
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != filepath.Clean(destDir) {
			return fmt.Errorf("entrada fora do destino: %s", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)&0o777); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0o777)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			_ = os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		default:
			// Skip other entries (hardlinks, devices, etc.) — not present in Go tarballs.
		}
	}
	return nil
}
