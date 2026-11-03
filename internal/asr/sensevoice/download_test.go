package sensevoice

import "testing"

func TestStripTopAndWanted(t *testing.T) {
	got := stripTop(ModelDirName + "/test_wavs/zh.wav")
	if got != "test_wavs/zh.wav" {
		t.Fatalf("strip %q", got)
	}
	if !wanted("test_wavs/zh.wav") || !wanted(ModelFile) || wanted("export-onnx.py") {
		t.Fatal("wanted filter mismatch")
	}
	if _, err := safeJoin("/tmp/model", "../etc/passwd"); err == nil {
		t.Fatal("expected unsafe path to fail")
	}
}
