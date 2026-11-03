// Package resident keeps the speech model and a local LLM from sitting in
// memory together. Remote notes do not take a residency slot.
package resident

import (
	"fmt"
	"sync"
)

const both = "识别器和本机模型不能同时留在内存里"

var (
	mu            sync.Mutex
	asrCount      int
	llmHeld       bool
	localRunnable bool
)

// LocalRunnable reports whether a local LLM is actually linked.
// v0.1 does not link llama.cpp, so this stays false and nothing is loaded.
func LocalRunnable() bool {
	mu.Lock()
	defer mu.Unlock()
	return localRunnable
}

// HoldASR marks one recognizer resident. It fails when a local LLM is loaded.
func HoldASR() error {
	mu.Lock()
	defer mu.Unlock()
	if llmHeld {
		return fmt.Errorf("%s", both)
	}
	asrCount++
	return nil
}

// ReleaseASR drops one recognizer hold.
func ReleaseASR() {
	mu.Lock()
	defer mu.Unlock()
	if asrCount > 0 {
		asrCount--
	}
}

// ASRHeld reports whether a recognizer is resident.
func ASRHeld() bool {
	mu.Lock()
	defer mu.Unlock()
	return asrCount > 0
}

// HoldLocal marks a local LLM resident. It fails when a recognizer is loaded
// or when no local runner is linked.
func HoldLocal() error {
	mu.Lock()
	defer mu.Unlock()
	if !localRunnable {
		return fmt.Errorf("本机模型尚未接入")
	}
	if asrCount > 0 {
		return fmt.Errorf("%s", both)
	}
	llmHeld = true
	return nil
}

// ReleaseLocal drops the local LLM hold.
func ReleaseLocal() {
	mu.Lock()
	defer mu.Unlock()
	llmHeld = false
}
