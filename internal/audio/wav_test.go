package audio

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestReadWAVPCM16(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tone.wav")
	raw := []int16{0, 16384, -16384, 32767}
	writePCM16(t, path, 16000, raw)
	samples, err := ReadMono16k(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != len(raw) {
		t.Fatalf("len %d", len(samples))
	}
	if samples[1] < 0.49 || samples[1] > 0.51 {
		t.Fatalf("sample1 %v", samples[1])
	}
}

func TestOfficialFixtureParses(t *testing.T) {
	root := repoRoot()
	if root == "" {
		t.Skip("no repo root")
	}
	path := filepath.Join(root, "models", "sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17", "test_wavs", "zh.wav")
	if _, err := os.Stat(path); err != nil {
		t.Skip("official zh.wav not downloaded")
	}
	samples, err := ReadMono16k(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) < 80000 {
		t.Fatalf("zh.wav samples %d", len(samples))
	}
	if RMS(samples) < 0.01 {
		t.Fatalf("zh.wav rms %v, fixture looks silent", RMS(samples))
	}
}

func writePCM16(t *testing.T, path string, rate int, samples []int16) {
	t.Helper()
	data := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(s))
	}
	// 44-byte canonical header.
	hdr := make([]byte, 44)
	copy(hdr[0:], "RIFF")
	binary.LittleEndian.PutUint32(hdr[4:], uint32(36+len(data)))
	copy(hdr[8:], "WAVE")
	copy(hdr[12:], "fmt ")
	binary.LittleEndian.PutUint32(hdr[16:], 16)
	binary.LittleEndian.PutUint16(hdr[20:], 1)
	binary.LittleEndian.PutUint16(hdr[22:], 1)
	binary.LittleEndian.PutUint32(hdr[24:], uint32(rate))
	binary.LittleEndian.PutUint32(hdr[28:], uint32(rate*2))
	binary.LittleEndian.PutUint16(hdr[32:], 2)
	binary.LittleEndian.PutUint16(hdr[34:], 16)
	copy(hdr[36:], "data")
	binary.LittleEndian.PutUint32(hdr[40:], uint32(len(data)))
	if err := os.WriteFile(path, append(hdr, data...), 0o644); err != nil {
		t.Fatal(err)
	}
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
