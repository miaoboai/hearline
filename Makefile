# Hearline v0.1 — Linux CPU build.
# Requires Go 1.27.3+ (MyGo) and a C toolchain (sherpa-onnx).
GO ?= go
export CGO_ENABLED := 1

BIN := bin/hearline
MODEL_DIR ?= models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17

.PHONY: build test test-asr demo download sources tidy

build:
	mkdir -p bin
	$(GO) build -p 2 -trimpath -o $(BIN) ./cmd/hearline

tidy:
	$(GO) mod tidy

test:
	$(GO) test -p 2 -count=1 -short ./...

# Loads SenseVoice-Small Q8 and decodes the official zh fixture when it is present.
test-asr: build
	HEARLINE_MODEL_DIR=$(CURDIR)/$(MODEL_DIR) $(GO) test -p 1 -count=1 -timeout 15m ./internal/asr/sensevoice

download: build
	./$(BIN) download-model --dir $(CURDIR)/$(MODEL_DIR)

demo: build
	./$(BIN) demo --model-dir $(CURDIR)/$(MODEL_DIR) --out /tmp/hearline-subs.txt --screenshot /tmp/hearline-subs.png --notes /tmp/hearline-notes.md

sources: build
	./$(BIN) sources

tidy-check: tidy
	git diff --exit-code go.mod go.sum
