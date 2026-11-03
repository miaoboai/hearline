package sensevoice

import (
	"archive/tar"
	"compress/bzip2"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Download fetches the official int8 archive into dir and keeps
// model.int8.onnx, tokens.txt, and test_wavs. A directory that is already
// Ready is left unchanged.
func Download(ctx context.Context, dir string) error {
	if dir == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(cache, "hearline", "sensevoice-small-q8")
	}
	if Ready(dir) {
		fmt.Fprintf(os.Stderr, "SenseVoice-Small Q8 already present in %s\n", dir)
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ModelURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "hearline/0.1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download SenseVoice-Small Q8: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download SenseVoice-Small Q8: HTTP %s", resp.Status)
	}
	fmt.Fprintf(os.Stderr, "downloading %s\n", ModelURL)
	body := &progressReader{r: resp.Body, total: resp.ContentLength}
	zr := bzip2.NewReader(body)
	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read model archive: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		rel := stripTop(hdr.Name)
		if rel == "" || !wanted(rel) {
			if _, err := io.Copy(io.Discard, tr); err != nil {
				return err
			}
			continue
		}
		dest, err := safeJoin(dir, rel)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := writeFile(dest, tr, hdr.FileInfo().Mode()); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr)
	if !Ready(dir) {
		return fmt.Errorf("archive extracted but %s is not a complete Q8 model", dir)
	}
	fmt.Fprintf(os.Stderr, "SenseVoice-Small Q8 ready in %s\n", dir)
	return nil
}

func wanted(rel string) bool {
	base := filepath.Base(rel)
	switch base {
	case ModelFile, TokensFile, "README.md", "LICENSE":
		return true
	}
	return strings.HasPrefix(rel, "test_wavs/") && strings.HasSuffix(base, ".wav")
}

func stripTop(name string) string {
	name = strings.TrimPrefix(name, "./")
	parts := strings.Split(name, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[1:], "/")
}

func safeJoin(root, rel string) (string, error) {
	clean := filepath.Clean(rel)
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", fmt.Errorf("unsafe archive path %q", rel)
	}
	return filepath.Join(root, clean), nil
}

func writeFile(dest string, r io.Reader, mode os.FileMode) error {
	if mode == 0 {
		mode = 0o644
	}
	tmp := dest + ".partial"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, r)
	closeErr := f.Close()
	if copyErr != nil {
		os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(tmp)
		return closeErr
	}
	return os.Rename(tmp, dest)
}

type progressReader struct {
	r     io.Reader
	total int64
	n     int64
	next  time.Time
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.n += int64(n)
	if time.Now().After(p.next) {
		p.next = time.Now().Add(time.Second)
		if p.total > 0 {
			fmt.Fprintf(os.Stderr, "\rSenseVoice-Small Q8  %d / %d MB", p.n/1e6, p.total/1e6)
		} else {
			fmt.Fprintf(os.Stderr, "\rSenseVoice-Small Q8  %d MB", p.n/1e6)
		}
	}
	return n, err
}
