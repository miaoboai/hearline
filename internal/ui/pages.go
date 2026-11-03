package ui

import (
	"strings"

	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/notes"
)

func (m *Model) settingsPage(c *g.Context) {
	remoteOn := m.remoteDisclosure()
	g.Scroll(c).Grow(1).Gap(12).Children(func() {
		card(c, func() {
			g.Text(c, "识别").FontSize(13).FontWeight(650).TextColor(colorQuiet)
			g.Row(c).Gap(12).Children(func() {
				labeledGrow(c, "语言", func() {
					g.Select(c.Key("lang"), &m.LangLabel, langLabels).Grow(1)
				})
				labeledGrow(c, "灵敏度", func() {
					g.Select(c.Key("vad"), &m.VADLabel, vadLabels).Grow(1)
				})
			})
			labeled(c, "模型路径", func() {
				field := g.TextInput(c.Key("model"), &m.ModelDir).Placeholder("留空则自动查找").MinWidth(360).Grow(1)
				if m.Focus == focusModel {
					field.Focus()
					m.Focus = ""
				}
			})
		})
		card(c, func() {
			g.Text(c, "纪要").FontSize(13).FontWeight(650).TextColor(colorQuiet)
			g.Segmented(c.Key("backend"), &m.Backend, "本机", "远程").Shrink(0)
			if m.Backend == 1 {
				labeled(c, "地址", func() {
					g.TextInput(c.Key("base"), &m.BaseURL).Placeholder("https://…/v1").MinWidth(360).Grow(1)
				})
				g.Row(c).Gap(12).Children(func() {
					labeledGrow(c, "模型", func() {
						g.TextInput(c.Key("remote-model"), &m.ModelName).Placeholder("模型名").Grow(1)
					})
					labeledGrow(c, "密钥", func() {
						g.TextInput(c.Key("key"), &m.APIKey).Password().Placeholder("只保存在本机").Grow(1)
					})
				})
			} else {
				g.Text(c, notes.ReasonLocal).FontSize(14).TextColor(colorMuted)
			}
		})
		card(c, func() {
			privacyBlock(c, remoteOn)
		})
		g.Row(c).Gap(12).AlignItems(g.End).Shrink(0).Children(func() {
			labeledGrow(c, "导出到", func() {
				g.TextInput(c.Key("export-path"), &m.ExportPath).MinWidth(360).Grow(1)
			})
			if g.PrimaryButton(c.Key("save"), "保存").Shrink(0).Clicked() {
				m.saveSettings()
			}
		})
		if m.SaveNote != "" {
			g.Text(c, m.SaveNote).FontSize(13).TextColor(colorQuiet).Shrink(0)
		}
	})
}

func (m *Model) aboutPage(c *g.Context) {
	remoteOn := m.remoteDisclosure()
	version := m.Version
	if version == "" {
		version = "0.1"
	}
	g.Column(c).Grow(1).Children(func() {
		card(c, func() {
			g.Text(c, "Hearline").FontSize(22).FontWeight(650).TextColor(colorInk)
			g.Text(c, version).FontSize(13).TextColor(colorQuiet)
			g.Text(c, "SenseVoice-Small · 本机 CPU").FontSize(14).TextColor(colorMuted)
			g.Column(c).Gap(6).Padding(8, 0, 0, 0).Children(func() {
				privacyBlock(c, remoteOn)
			})
			g.Text(c, shortcutHint()+"  ·  Esc 停止").FontSize(12).TextColor(colorQuiet)
		})
		g.Spacer(c)
	})
}

func (m *Model) notesPage(c *g.Context) {
	ok, reason := notes.Available(m.llm)
	g.Scroll(c).Grow(1).Gap(14).Children(func() {
		card(c, func() {
			g.Row(c).AlignItems(g.Center).Gap(12).Children(func() {
				g.Text(c, "纪要").FontSize(22).FontWeight(650).TextColor(colorInk).Grow(1)
				if !ok {
					g.Text(c, reason).FontSize(13).TextColor(colorQuiet)
					return
				}
				btn := g.PrimaryButton(c.Key("make-notes"), "整理").Shrink(0)
				if m.NotesBusy || len(m.finals()) == 0 {
					btn = btn.Disabled(true)
				}
				if btn.Clicked() {
					m.wantNotes = true
					m.generateNotes()
				}
			})
			if !ok {
				return
			}
			if m.NotesBusy {
				g.Text(c, "正在整理").FontSize(14).TextColor(colorQuiet)
			} else if m.NotesErr != "" {
				g.Text(c, m.NotesErr).FontSize(14).TextColor(colorMuted)
			}
			if docEmpty(m.Doc) && !m.NotesBusy && m.NotesErr == "" {
				g.Text(c, "整理后，要点会写在这里").FontSize(14).TextColor(colorQuiet)
				return
			}
			noteSection(c, "要点", m.Doc.Points)
			noteSection(c, "待办", m.Doc.Todos)
			if len(m.Doc.Terms) > 0 {
				g.Text(c, "术语").FontSize(12).FontWeight(650).TextColor(colorQuiet)
				g.Column(c).Gap(8).Children(func() {
					for _, term := range m.Doc.Terms {
						text := term.Word
						if term.Note != "" {
							text += "  ·  " + term.Note
						}
						g.Text(c, text).FontSize(15).TextColor(colorInk)
					}
				})
			}
		})
	})
}

func noteSection(c *g.Context, title string, items []string) {
	if len(items) == 0 {
		return
	}
	g.Column(c).Gap(8).Children(func() {
		g.Text(c, title).FontSize(12).FontWeight(650).TextColor(colorQuiet)
		for _, item := range items {
			g.Text(c, item).FontSize(16).TextColor(colorInk)
		}
	})
}

func docEmpty(doc notes.Doc) bool {
	return len(doc.Points) == 0 && len(doc.Todos) == 0 && len(doc.Terms) == 0
}

func card(c *g.Context, body func()) {
	g.Column(c).Shrink(0).Padding(18, 20).Gap(12).Radius(16).
		Background(colorCard).Border(1, colorLine).
		Shadow(0, 8, 24, 0, g.RGBA(21, 23, 28, 0.05)).
		Children(body)
}

func labeled(c *g.Context, name string, field func()) {
	g.Column(c).Gap(6).Children(func() {
		g.Text(c, name).FontSize(12).FontWeight(600).TextColor(colorQuiet)
		field()
	})
}

func labeledGrow(c *g.Context, name string, field func()) {
	g.Column(c).Gap(6).Grow(1).Children(func() {
		g.Text(c, name).FontSize(12).FontWeight(600).TextColor(colorQuiet)
		field()
	})
}

// remoteDisclosure is the settings choice while that page is open, and the
// saved backend everywhere else. The second privacy line follows it.
func (m *Model) remoteDisclosure() bool {
	if m == nil {
		return false
	}
	if m.Screen == screenSettings {
		return m.Backend == 1
	}
	return m.llm.Backend == "remote"
}

func (m *Model) saveSettings() {
	cfg := m.cfg
	prev := cfg.ASR
	cfg.ASR.Language = langValue(m.LangLabel)
	cfg.ASR.VAD = vadValue(m.VADLabel)
	cfg.ASR.ModelDir = strings.TrimSpace(m.ModelDir)
	if m.Backend == 1 {
		cfg.LLM.Backend = "remote"
		cfg.LLM.Remote.BaseURL = strings.TrimSpace(m.BaseURL)
		cfg.LLM.Remote.Model = strings.TrimSpace(m.ModelName)
		cfg.LLM.Remote.APIKey = strings.TrimSpace(m.APIKey)
		if cfg.LLM.Remote.BaseURL == "" || cfg.LLM.Remote.Model == "" {
			m.SaveNote = "填写接口地址和模型名"
			return
		}
	} else {
		cfg.LLM.Backend = "local"
	}
	if m.save == nil {
		m.SaveNote = "没能保存"
		return
	}
	if err := m.save(cfg); err != nil {
		m.SaveNote = "没能保存"
		return
	}
	m.cfg = cfg
	m.llm = cfg.LLM
	m.chunk.Sensitivity = cfg.ASR.VAD
	m.SaveNote = "已保存"
	if m.Fault == fault.ModelMissing || asrNeedsReload(prev, cfg.ASR) {
		if m.Listening {
			m.reloadASR = true
			return
		}
		m.reload()
	}
}

func asrNeedsReload(prev, next config.ASR) bool {
	if strings.TrimSpace(prev.ModelDir) != strings.TrimSpace(next.ModelDir) {
		return true
	}
	return !strings.EqualFold(prev.Language, next.Language)
}

var langLabels = []string{"自动", "中文", "English", "日本語", "한국어", "粤语"}

func langValue(label string) string {
	switch label {
	case "中文":
		return "zh"
	case "English":
		return "en"
	case "日本語":
		return "ja"
	case "한국어":
		return "ko"
	case "粤语":
		return "yue"
	default:
		return "auto"
	}
}

func langChoice(value string) string {
	switch strings.ToLower(value) {
	case "zh", "zh-cn":
		return "中文"
	case "en":
		return "English"
	case "ja":
		return "日本語"
	case "ko":
		return "한국어"
	case "yue":
		return "粤语"
	default:
		return "自动"
	}
}

var vadLabels = []string{"灵敏", "标准", "稳妥"}

func vadValue(label string) string {
	switch label {
	case "灵敏":
		return "high"
	case "稳妥":
		return "low"
	default:
		return "mid"
	}
}

func vadChoice(value string) string {
	switch strings.ToLower(value) {
	case "high", "灵敏":
		return "灵敏"
	case "low", "稳妥":
		return "稳妥"
	default:
		return "标准"
	}
}
