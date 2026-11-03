package resident

import "testing"

func TestASRAndLocalLLMExcludeEachOther(t *testing.T) {
	t.Cleanup(func() {
		mu.Lock()
		asrCount = 0
		llmHeld = false
		localRunnable = false
		mu.Unlock()
	})
	if LocalRunnable() {
		t.Fatal("v0.1 must not report a local LLM as runnable")
	}
	if err := HoldASR(); err != nil {
		t.Fatal(err)
	}
	if err := HoldASR(); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	localRunnable = true
	mu.Unlock()
	if err := HoldLocal(); err == nil {
		t.Fatal("local LLM loaded beside the recognizer")
	}
	ReleaseASR()
	ReleaseASR()
	if ASRHeld() {
		t.Fatal("recognizer still held")
	}
	if err := HoldLocal(); err != nil {
		t.Fatal(err)
	}
	if err := HoldASR(); err == nil {
		t.Fatal("recognizer loaded beside the local LLM")
	}
	ReleaseLocal()
}
