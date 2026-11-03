// Package sensevoice runs SenseVoice-Small Q8 (sherpa-onnx int8 ONNX) on CPU.
// The recognizer is offline, so the subtitle loop feeds it short utterances
// rather than an unbounded stream.
package sensevoice

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// ModelDirName is the official sherpa-onnx int8 package directory.
	ModelDirName = "sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17"
	// ModelFile is the Q8 (int8) ONNX weights, about 229 MB.
	ModelFile = "model.int8.onnx"
	// TokensFile is the matching token table.
	TokensFile = "tokens.txt"
	// ModelURL is the CPU int8 archive (~156 MB compressed) published by k2-fsa.
	ModelURL = "https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17.tar.bz2"

	minModelBytes int64 = 200_000_000
)

// Paths are the files a runner needs.
type Paths struct {
	Dir    string
	Model  string
	Tokens string
}

// Ready reports whether dir contains a full Q8 model.
func Ready(dir string) bool {
	_, err := normalize(dir)
	return err == nil
}

// Resolve finds a downloaded model. explicit must be that directory when set.
// Otherwise HEARLINE_MODEL_DIR, ./models/<package>, the user cache, and the
// executable directory are tried.
func Resolve(explicit string) (Paths, error) {
	if explicit != "" {
		return normalize(explicit)
	}
	var candidates []string
	if env := os.Getenv("HEARLINE_MODEL_DIR"); env != "" {
		candidates = append(candidates, env)
	}
	if root := repoRoot(); root != "" {
		candidates = append(candidates,
			filepath.Join(root, "models", ModelDirName),
			filepath.Join(root, "models", "sensevoice-small-q8"),
		)
	}
	if cache, err := os.UserCacheDir(); err == nil {
		candidates = append(candidates, filepath.Join(cache, "hearline", "sensevoice-small-q8"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "models", ModelDirName))
	}
	var last error
	for _, c := range candidates {
		p, err := normalize(c)
		if err == nil {
			return p, nil
		}
		last = err
	}
	if last == nil {
		last = fmt.Errorf("no candidate directories")
	}
	return Paths{}, fmt.Errorf("找不到 SenseVoice-Small Q8（model.int8.onnx + tokens.txt）。运行 hearline download-model。%w", last)
}

func normalize(dir string) (Paths, error) {
	if p, err := pathsIn(dir); err == nil {
		return p, nil
	}
	nested := filepath.Join(dir, ModelDirName)
	if p, err := pathsIn(nested); err == nil {
		return p, nil
	}
	return Paths{}, fmt.Errorf("%s 缺少 %s（需大于 %d 字节）或 %s", dir, ModelFile, minModelBytes, TokensFile)
}

func pathsIn(dir string) (Paths, error) {
	model := filepath.Join(dir, ModelFile)
	tokens := filepath.Join(dir, TokensFile)
	fi, err := os.Stat(model)
	if err != nil {
		return Paths{}, err
	}
	if fi.Size() < minModelBytes {
		return Paths{}, fmt.Errorf("%s is %d bytes", model, fi.Size())
	}
	if _, err := os.Stat(tokens); err != nil {
		return Paths{}, err
	}
	return Paths{Dir: dir, Model: model, Tokens: tokens}, nil
}

// TestWav returns the official fixture shipped inside the model archive.
func TestWav(dir, name string) string {
	return filepath.Join(dir, "test_wavs", name)
}

func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
