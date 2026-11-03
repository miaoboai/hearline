package sensevoice

import (
	"os"
	"strings"
	"testing"

	"github.com/miaoboai/hearline/internal/audio"
	"github.com/miaoboai/hearline/internal/config"
)

func TestDecodeOfficialZh(t *testing.T) {
	if testing.Short() {
		t.Skip("skip SenseVoice load in -short")
	}
	dir := os.Getenv("HEARLINE_MODEL_DIR")
	paths, err := Resolve(dir)
	if err != nil {
		t.Skip(err)
	}
	wav := TestWav(paths.Dir, "zh.wav")
	samples, err := audio.ReadMono16k(wav)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default().ASR
	cfg.ModelDir = paths.Dir
	cfg.NumThreads = 2
	rec, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Close()
	got, err := rec.DecodePCM(samples, audio.SampleRate)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("zh text=%q lang=%q emotion=%q event=%q", got.Text, got.Language, got.Emotion, got.Event)
	if !strings.Contains(got.Text, "点") || !strings.Contains(got.Text, "下午") {
		t.Fatalf("unexpected transcript %q", got.Text)
	}
	if got.Language != "" && !strings.Contains(got.Language, "zh") {
		t.Fatalf("language %q", got.Language)
	}
}
