package ui

import (
	"fmt"
	"strings"

	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/notes"
)

func (m *Model) settingsPage(c *g.Context) {
	focusModel := m.Focus == focusModel
	if focusModel {
		m.Focus = ""
	}
	g.Column(c).Grow(1).Gap(GapInline).Children(func() {
		g.Scroll(c).Grow(1).Children(func() {
			m.settingsBody(c, focusModel)
		})
		g.Row(c).Gap(GapInline).AlignItems(g.Center).Shrink(0).MinHeight(RowSettings).Children(func() {
			text(c, "导出到", TypeBody, colorInk).Shrink(0)
			textField(c, "export-path", &m.ExportPath, "", false, false)
			if paintPrimaryLive(c, "save", "保存", false) {
				m.saveSettings()
			}
		})
		if m.SaveNote != "" {
			text(c, m.SaveNote, TypeCaption, colorQuiet).Shrink(0)
		}
	})
}

func (m *Model) settingsBody(c *g.Context, focusModel bool) {
	g.Column(c).Gap(0).Children(func() {
		text(c, "识别", TypeCaption, colorMuted).Margin(0, 0, Space1, 0)
		m.settingRow(c, "语言", func() {
			selectField(c, "lang", &m.LangLabel, langLabels, false)
		})
		m.settingRow(c, "灵敏度", func() {
			selectField(c, "vad", &m.VADLabel, vadLabels, false)
		})
		m.settingRow(c, "音源", func() {
			g.Row(c).Gap(GapInline).AlignItems(g.Center).Grow(1).Children(func() {
				if len(m.Labels) == 0 {
					text(c, "没找到系统声，可从文件听写", TypeCaption, colorMuted).Grow(1)
				} else {
					selectField(c, "source", &m.Selected, m.Labels, m.Focus == focusSource)
					if m.Focus == focusSource {
						m.Focus = ""
					}
				}
				if _, clicked := paintLink(c, "refresh", "刷新", false); clicked {
					m.refresh(nil)
					if m.Fault == fault.CaptureUnavailable && len(m.Labels) > 0 {
						m.clearFault(fault.CaptureUnavailable)
					}
				}
			})
		})
		m.settingRow(c, "从文件", func() {
			g.Row(c).Gap(GapInline).AlignItems(g.Center).Grow(1).Children(func() {
				if _, clicked := paintLink(c, "browse", "选择…", false); clicked {
					m.browse()
				}
				name := m.FileName
				if name == "" {
					name = "还没有选择文件"
				}
				el := text(c, name, TypeCaption, colorMuted).Grow(1)
				if m.WAV != "" {
					el.Tooltip(m.WAV)
				}
				disabled := !m.Ready || m.Listening || strings.TrimSpace(m.WAV) == ""
				if _, clicked := paintLink(c, "play", "听这个文件", disabled); clicked && !disabled {
					m.listenFile()
				}
			})
		})
		m.settingRow(c, "叠在其他窗口上", func() {
			if toggle(c, "ontop", &m.OnTop) && m.win != nil {
				m.win.SetAlwaysOnTop(m.OnTop)
			}
		})

		text(c, "关于", TypeCaption, colorMuted).Margin(Space2, 0, Space1, 0)
		m.aboutBlock(c)
		hairline(c)

		text(c, "纪要", TypeCaption, colorMuted).Margin(Space2, 0, Space1, 0)
		m.fieldBlock(c, "地址", func() {
			textField(c, "base", &m.BaseURL, "https://…/v1", false, false)
		})
		m.fieldBlock(c, "模型", func() {
			textField(c, "remote-model", &m.ModelName, "模型名", false, false)
		})
		m.fieldBlock(c, "密钥", func() {
			textField(c, "key", &m.APIKey, "只保存在本机", false, false).Password()
		})

		g.Row(c).MinHeight(RowSettings).AlignItems(g.Center).Children(func() {
			label := "高级"
			if m.ShowAdvanced {
				label = "高级  收起"
			}
			if _, clicked := paintLink(c, "advanced", label, false); clicked {
				m.ShowAdvanced = !m.ShowAdvanced
			}
		})
		hairline(c)
		if m.ShowAdvanced {
			m.fieldBlock(c, "模型路径", func() {
				field := textField(c, "model", &m.ModelDir, "留空则自动查找", focusModel, true)
				if focusModel {
					field.ScrollIntoView()
				}
			})
			m.settingRow(c, "线程", func() {
				g.Row(c).Gap(GapInline).AlignItems(g.Center).Children(func() {
					if _, clicked := paintLink(c, "thr-dec", "减少", m.Threads <= 1); clicked && m.Threads > 1 {
						m.Threads--
					}
					n := m.Threads
					if n <= 0 {
						n = 2
					}
					text(c, fmt.Sprintf("%d / 4", n), TypeBody, colorInk)
					if _, clicked := paintLink(c, "thr-inc", "增加", m.Threads >= 4); clicked && m.Threads < 4 {
						m.Threads++
					}
				})
			})
			text(c, "SenseVoice-Small · 本机 CPU", TypeCaption, colorQuiet).Margin(Space1, 0, Space1, 0)
		}
	})
}

func (m *Model) settingRow(c *g.Context, name string, control func()) {
	g.Column(c).Shrink(0).Children(func() {
		g.Row(c).MinHeight(RowSettings).AlignItems(g.Center).Gap(GapBlock).Children(func() {
			text(c, name, TypeBody, colorInk).Shrink(0)
			g.Spacer(c)
			control()
		})
		hairline(c)
	})
}

func (m *Model) fieldBlock(c *g.Context, name string, field func()) {
	g.Column(c).Gap(SpaceHalf).Shrink(0).Padding(Space1, 0, Space1, 0).Children(func() {
		text(c, name, TypeCaption, colorMuted)
		field()
	})
	hairline(c)
}

func (m *Model) aboutBlock(c *g.Context) {
	version := m.Version
	if version == "" {
		version = "0.1"
	}
	text(c, "Hearline "+version, TypeBody, colorInk)
	g.Column(c).Gap(SpaceHalf).Padding(Space1, 0, 0, 0).Children(func() {
		privacyBlock(c, m.remoteDisclosure())
	})
}

func (m *Model) notesReady() bool {
	ok, _ := notes.Available(m.llm)
	return ok
}

func (m *Model) notesOverlay(c *g.Context) {
	if !m.notesReady() {
		m.NotesOpen = false
		if m.Panel == panelNotes {
			m.Panel = panelListen
		}
		return
	}
	g.DialogBase(c, &m.NotesOpen, func(back, panel g.Element) {
		back.Background(maskScrim)
		panel.Width(420).Padding(Space3).Gap(GapBlock).Radius(RadiusOverlay).Background(colorCard).
			Shadow(0, 4, 12, 0, shadowOverlay)
		g.Row(c).AlignItems(g.Center).Children(func() {
			text(c, "纪要", TypeBrand, colorInk).Grow(1)
			if _, clicked := paintLink(c, "close-notes", "关闭", false); clicked {
				m.NotesOpen = false
			}
		})
		disabled := m.NotesBusy || len(m.finals()) == 0
		label := "生成纪要"
		if m.NotesBusy {
			label = "正在整理"
		}
		g.Row(c).Children(func() {
			if paintPrimaryLive(c, "make-notes", label, disabled) && !disabled {
				m.wantNotes = true
				m.generateNotes()
			}
		})
		g.Scroll(c).MaxHeight(360).Gap(GapBlock).Children(func() {
			if docEmpty(m.Doc) && !m.NotesBusy && m.NotesErr == "" {
				text(c, "听写一段时间后，可在这里生成纪要", TypeEmpty, colorQuiet)
				return
			}
			if m.NotesErr != "" {
				text(c, m.NotesErr, TypeBody, colorDanger)
			}
			noteSection(c, "要点", m.Doc.Points)
			noteSection(c, "待办", m.Doc.Todos)
			if len(m.Doc.Terms) > 0 {
				text(c, "术语", TypeCaption, colorMuted)
				g.Column(c).Gap(Space1).Children(func() {
					for _, term := range m.Doc.Terms {
						line := term.Word
						if term.Note != "" {
							line += "  ·  " + term.Note
						}
						text(c, line, TypeBody, colorInk)
					}
				})
			}
		})
	})
	if !m.NotesOpen && m.Panel == panelNotes {
		m.Panel = panelListen
	}
}

func noteSection(c *g.Context, title string, items []string) {
	if len(items) == 0 {
		return
	}
	g.Column(c).Gap(Space1).Children(func() {
		text(c, title, TypeCaption, colorMuted)
		for _, item := range items {
			text(c, item, TypeBody, colorInk)
		}
	})
}

func docEmpty(doc notes.Doc) bool {
	return len(doc.Points) == 0 && len(doc.Todos) == 0 && len(doc.Terms) == 0
}

func (m *Model) remoteDisclosure() bool {
	if m == nil {
		return false
	}
	if m.Panel == panelSettings {
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
	if m.Threads < 1 {
		m.Threads = 1
	}
	if m.Threads > 4 {
		m.Threads = 4
	}
	cfg.ASR.NumThreads = m.Threads
	base := strings.TrimSpace(m.BaseURL)
	model := strings.TrimSpace(m.ModelName)
	switch {
	case base == "" && model == "":
		cfg.LLM.Backend = "local"
		m.Backend = 0
	case base == "" || model == "":
		m.SaveNote = "填写接口地址和模型名"
		return
	default:
		cfg.LLM.Backend = "remote"
		cfg.LLM.Remote.BaseURL = base
		cfg.LLM.Remote.Model = model
		cfg.LLM.Remote.APIKey = strings.TrimSpace(m.APIKey)
		m.Backend = 1
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
	if !m.notesReady() && m.Panel == panelNotes {
		m.Panel = panelListen
		m.NotesOpen = false
	}
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
	if prev.Threads() != next.Threads() {
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
