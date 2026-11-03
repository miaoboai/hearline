// Package llm is the translation and notes backend. v0.1 stores the choice
// and can call an OpenAI-compatible remote, but the live subtitle loop does
// not invoke it. Local inference (llama.cpp) is intentionally not linked yet.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/miaoboai/hearline/internal/config"
)

// Describe is a one-line status for the UI. It never includes the API key.
func Describe(c config.LLM) string {
	switch c.Backend {
	case "remote":
		key := "未填写 key"
		if c.Remote.APIKey != "" {
			key = "key 已保存"
		}
		return fmt.Sprintf("远程 OpenAI 兼容 · %s · %s · %s（字幕环不调用）", c.Remote.Model, c.Remote.BaseURL, key)
	default:
		path := c.Local.ModelPath
		if path == "" {
			path = "未设置模型路径"
		}
		return fmt.Sprintf("本地优先 · %s · %s（字幕环不调用）", c.Local.Runtime, path)
	}
}

// Complete runs the configured backend. Remote failures can fall back to the
// local runner, which in v0.1 returns a clear "not linked" error.
func Complete(ctx context.Context, c config.LLM, prompt string) (string, error) {
	if err := (config.Config{LLM: c}).Validate(); err != nil && c.Backend == "remote" {
		return "", err
	}
	if c.Backend == "remote" {
		text, err := completeRemote(ctx, c.Remote, prompt)
		if err != nil && c.FallbackToLocal {
			_, localErr := completeLocal(c.Local, prompt)
			return "", fmt.Errorf("远程 LLM 失败: %v; 本地回退: %w", err, localErr)
		}
		return text, err
	}
	return completeLocal(c.Local, prompt)
}

func completeLocal(c config.LocalLLM, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("empty prompt")
	}
	path := c.ModelPath
	if path == "" {
		path = "(未设置 llm.local.model_path)"
	}
	runtime := c.Runtime
	if runtime == "" {
		runtime = "llama.cpp"
	}
	return "", fmt.Errorf("本地 LLM（%s）尚未接入 v0.1，模型路径 %s 只保存在配置里", runtime, path)
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func completeRemote(ctx context.Context, c config.RemoteLLM, prompt string) (string, error) {
	if c.BaseURL == "" || c.Model == "" {
		return "", fmt.Errorf("remote LLM requires base_url and model")
	}
	body, err := json.Marshal(chatRequest{
		Model:    c.Model,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint(c.BaseURL), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("remote LLM HTTP %s: %s", resp.Status, truncate(string(raw), 300))
	}
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("remote LLM response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("remote LLM: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("remote LLM returned no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func endpoint(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/chat/completions") {
		return base
	}
	return base + "/chat/completions"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
