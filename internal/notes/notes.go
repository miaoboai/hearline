// Package notes turns finished subtitle lines into a short meeting note.
// The local runner is not linked. Remote is used only when the user has
// filled an OpenAI-compatible endpoint. Otherwise notes stay unavailable.
package notes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/llm"
	"github.com/miaoboai/hearline/internal/pipeline"
)

const (
	ReasonLocal  = "本机模型尚未接入"
	ReasonRemote = "先在设置里填写远程服务"
	ReasonEmpty  = "还没有可整理的句子"
)

// Term is one name from the transcript.
type Term struct {
	Word string `json:"word"`
	Note string `json:"note"`
}

// Doc is a structured note. Empty slices are fine.
type Doc struct {
	Points []string `json:"points"`
	Todos  []string `json:"todos"`
	Terms  []Term   `json:"terms"`
}

// Available reports whether 纪要 can run. Local llama.cpp is not wired,
// so only a complete remote endpoint counts. The reason is for a disabled
// control, not an error after a click.
func Available(c config.LLM) (bool, string) {
	if c.Backend == "remote" {
		if strings.TrimSpace(c.Remote.BaseURL) == "" || strings.TrimSpace(c.Remote.Model) == "" {
			return false, ReasonRemote
		}
		return true, ""
	}
	return false, ReasonLocal
}

// Finals drops partial and empty rows. Notes and export use this list only.
func Finals(lines []pipeline.Line) []pipeline.Line {
	out := make([]pipeline.Line, 0, len(lines))
	for _, line := range lines {
		if line.Partial || strings.TrimSpace(line.Text) == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

// Generate asks the configured remote model. It refuses when Available is false.
func Generate(ctx context.Context, c config.LLM, lines []pipeline.Line) (Doc, error) {
	ok, reason := Available(c)
	if !ok {
		return Doc{}, fmt.Errorf("%s", reason)
	}
	finals := Finals(lines)
	if len(finals) == 0 {
		return Doc{}, fmt.Errorf("%s", ReasonEmpty)
	}
	raw, err := llm.Complete(ctx, c, prompt(finals))
	if err != nil {
		return Doc{}, err
	}
	doc, err := parse(raw)
	if err != nil {
		return Doc{}, err
	}
	return trim(doc), nil
}

func prompt(lines []pipeline.Line) string {
	var b strings.Builder
	b.WriteString("你是会议纪要编辑。只根据下面已经确定的句子整理，不要编造没说过的事。")
	b.WriteString("只输出一个 JSON 对象，不要 Markdown。")
	b.WriteString(`{"points":["要点"],"todos":["待办"],"terms":[{"word":"术语","note":"短解释"}]}`)
	b.WriteString("要点、待办、术语各最多 6 条。没有就用空数组。\n")
	for i, line := range lines {
		fmt.Fprintf(&b, "%d. %s\n", i+1, strings.TrimSpace(line.Text))
	}
	return b.String()
}

func parse(raw string) (Doc, error) {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			s = s[i : j+1]
		}
	}
	var doc Doc
	if err := json.Unmarshal([]byte(s), &doc); err != nil {
		return Doc{}, fmt.Errorf("纪要格式无法读取")
	}
	return doc, nil
}

func trim(doc Doc) Doc {
	doc.Points = clip(doc.Points, 6)
	doc.Todos = clip(doc.Todos, 6)
	if len(doc.Terms) > 6 {
		doc.Terms = doc.Terms[:6]
	}
	out := doc.Terms[:0]
	for _, term := range doc.Terms {
		term.Word = strings.TrimSpace(term.Word)
		term.Note = strings.TrimSpace(term.Note)
		if term.Word == "" {
			continue
		}
		out = append(out, term)
	}
	doc.Terms = out
	return doc
}

func clip(items []string, n int) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
		if len(out) == n {
			break
		}
	}
	return out
}

// Markdown writes finished sentences, and a note when one exists.
func Markdown(doc *Doc, lines []pipeline.Line) string {
	finals := Finals(lines)
	var b strings.Builder
	if doc != nil && (len(doc.Points) > 0 || len(doc.Todos) > 0 || len(doc.Terms) > 0) {
		b.WriteString("# 纪要\n\n")
		writeList(&b, "要点", doc.Points)
		writeList(&b, "待办", doc.Todos)
		if len(doc.Terms) > 0 {
			b.WriteString("## 术语\n\n")
			for _, term := range doc.Terms {
				if term.Note != "" {
					fmt.Fprintf(&b, "- **%s** — %s\n", term.Word, term.Note)
				} else {
					fmt.Fprintf(&b, "- **%s**\n", term.Word)
				}
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("# 字幕\n\n")
	if len(finals) == 0 {
		b.WriteString("（还没有确定的句子）\n")
		return b.String()
	}
	for _, line := range finals {
		fmt.Fprintf(&b, "- %s\n", strings.TrimSpace(line.Text))
	}
	return b.String()
}

func writeList(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", title)
	for _, item := range items {
		fmt.Fprintf(b, "- %s\n", item)
	}
	b.WriteString("\n")
}
