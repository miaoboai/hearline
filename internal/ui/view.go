package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/pipeline"
)

const emptyHint = "开始听写后字幕会出现在这里"

func (m *Model) view(c *g.Context) {
	if c == nil || m == nil {
		return
	}
	c.SetTheme(hearlineTheme())
	if m.Gallery == galleryStates {
		m.primaryCollage(c)
		return
	}
	m.shortcuts(c)
	g.Column(c).Fill().Gap(0).Background(colorGround).Children(func() {
		m.header(c)
		if m.Panel == panelSettings {
			g.Column(c).Grow(1).Padding(PadWindow).Children(func() {
				m.settingsPage(c)
			})
			return
		}
		m.listenPage(c)
	})
	if m.Panel != panelSettings {
		m.notesOverlay(c)
		m.confirmQuitDialog(c)
	}
}

func (m *Model) primaryCollage(c *g.Context) {
	type shot struct {
		name string
		look btnLook
	}
	shots := []shot{
		{"default", lookDefault},
		{"hover", lookHover},
		{"pressed", lookPressed},
		{"disabled", lookDisabled},
	}
	g.Row(c).Fill().Padding(Space3).Gap(Space3).AlignItems(g.Center).Justify(g.Center).Background(colorGround).Children(func() {
		for _, shot := range shots {
			shot := shot
			g.Column(c).Gap(GapInline).AlignItems(g.Center).Children(func() {
				text(c, shot.name, TypeCaption, colorMuted)
				paintPrimary(c, "state-"+shot.name, "开始听", false, shot.look)
			})
		}
	})
}

func (m *Model) shortcuts(c *g.Context) {
	c.OnShortcut(0, g.KeySpace, m.toggleListen)
	c.OnShortcut(g.Cmd, g.KeyL, m.toggleListen)
	c.OnShortcut(g.Cmd, g.KeyE, func() { m.exportKind("srt") })
	c.OnShortcut(g.Cmd, g.KeyComma, func() { m.Panel = panelSettings })
	c.OnShortcut(g.Cmd|g.Shift, g.KeyC, m.copyCurrent)
	c.OnShortcut(0, g.KeyEscape, func() {
		if m.ConfirmQuit {
			m.ConfirmQuit = false
			return
		}
		if m.ExportOpen {
			m.ExportOpen = false
			return
		}
		if m.NotesOpen || m.Panel == panelNotes {
			m.NotesOpen = false
			m.Panel = panelListen
			return
		}
		if m.Panel != panelListen {
			m.Panel = panelListen
			return
		}
		if m.Listening {
			m.stop()
		}
	})
}

func (m *Model) header(c *g.Context) {
	g.Row(c).MinHeight(HeaderH).Padding(0, PadHeaderX).AlignItems(g.Center).Gap(GapInline).Shrink(0).Children(func() {
		g.Column(c).Gap(0).Shrink(0).Children(func() {
			text(c, "Hearline", TypeBrand, colorInk).SingleLine()
			text(c, "本地听写", TypeSub, colorQuiet).SingleLine()
		})
		m.sessionChip(c)
		g.Spacer(c)
		if m.notesReady() {
			if _, clicked := paintLink(c, "notes", "纪要", false); clicked {
				m.openNotes()
			}
		}
		if _, clicked := paintLink(c, "settings", "设置", false); clicked {
			if m.Panel == panelSettings {
				m.Panel = panelListen
			} else {
				m.Panel = panelSettings
				m.ConfirmQuit = false
				m.NotesOpen = false
			}
		}
		m.exportMenu(c)
	})
}

func (m *Model) exportMenu(c *g.Context) {
	disabled := len(m.finals()) == 0
	link, clicked := paintLink(c, "export", "导出", disabled)
	if disabled {
		m.ExportOpen = false
	} else if clicked {
		m.ExportOpen = !m.ExportOpen
	}
	g.PopoverBase(c, link, &m.ExportOpen, func(panel g.Element) {
		panel.MinWidth(180).Padding(SpaceHalf).Gap(SpaceHalf).Radius(RadiusMenu).
			Background(colorCard).Border(1, colorLine).
			Shadow(0, 4, 12, 0, shadowOverlay)
		if menuRow(c, "export-txt", "文本 (.txt)", false) {
			m.ExportOpen = false
			m.exportKind("txt")
		}
		if menuRow(c, "export-srt", "字幕 (.srt)", false) {
			m.ExportOpen = false
			m.exportKind("srt")
		}
		if menuRow(c, "export-md", "纪要 (.md)", docEmpty(m.Doc)) {
			m.ExportOpen = false
			m.exportKind("md")
		}
	})
}

func (m *Model) sessionChip(c *g.Context) {
	switch {
	case m.Fault != "" && !m.Listening:
		text(c, "出错", TypeCaption, colorMuted).Shrink(0)
	case !m.Ready && m.Fault == "":
		text(c, "准备中", TypeCaption, colorQuiet).Shrink(0)
	case m.Listening:
		g.Row(c).Gap(GapInline).AlignItems(g.Center).Shrink(0).Children(func() {
			g.Box(c).Size(ListenDot, ListenDot).Radius(ListenDot / 2).Background(colorAccent)
			text(c, "正在听 · "+m.clock(), TypeCurrentMeta, colorInk).SingleLine()
		})
	case m.DockNote != "":
		text(c, m.DockNote, TypeCaption, colorQuiet).Shrink(0)
	case len(m.finals()) > 0:
		text(c, fmt.Sprintf("已停止 · 共 %d 句", len(m.finals())), TypeCaption, colorQuiet).Shrink(0)
	}
}

func (m *Model) clock() string {
	d := m.ShownFor
	if m.Listening && d == 0 && !m.Started.IsZero() {
		d = time.Since(m.Started)
	}
	if d < 0 {
		d = 0
	}
	sec := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d", sec/60, sec%60)
}

func (m *Model) listenPage(c *g.Context) {
	m.stage(c)
}

func (m *Model) confirmQuitDialog(c *g.Context) {
	g.DialogBase(c, &m.ConfirmQuit, func(back, panel g.Element) {
		back.Background(maskScrim)
		panel.Padding(Space3).Gap(GapBlock).Radius(RadiusOverlay).Background(colorCard).
			Shadow(0, 4, 12, 0, shadowOverlay).MinWidth(280)
		text(c, "停止并退出？", TypeBody, colorInk)
		g.Row(c).Gap(GapInline).Justify(g.End).Children(func() {
			if _, clicked := paintLink(c, "stay", "继续听", false); clicked {
				m.ConfirmQuit = false
			}
			if paintPrimaryLive(c, "quit", "退出", false) {
				m.allowClose = true
				m.stop()
				if m.win != nil {
					m.win.Close()
				}
			}
		})
	})
}

func (m *Model) stage(c *g.Context) {
	g.Column(c).Grow(1).MinHeight(240).Padding(PadStageY, PadStageX).Radius(RadiusStage).Gap(GapBlock).
		Background(colorStage).
		Children(func() {
			finals := m.finals()
			partial, hasPartial := m.livePartial()
			if len(finals) == 0 && !hasPartial {
				m.emptyStage(c)
				return
			}
			if msg := m.faultText(); msg != "" {
				text(c, msg, TypeBody, colorOnDark).Shrink(0)
			}
			g.Spacer(c)
			var history []pipeline.Line
			var current *pipeline.Line
			switch {
			case hasPartial:
				history = finals
			case len(finals) > 0:
				history = finals[:len(finals)-1]
				last := finals[len(finals)-1]
				current = &last
			}
			if len(history) > 0 {
				m.history(c, history)
			}
			if hasPartial {
				face := TypePartial
				if len([]rune(partial.Text)) > 42 {
					face = TypePartialLong
				}
				text(c, partial.Text, face, colorPartial).Italic().Shrink(0)
			} else if current != nil {
				m.currentLine(c, *current)
			}
			g.Row(c).Shrink(0).Children(func() {
				m.stageAction(c)
			})
		})
}

func (m *Model) emptyStage(c *g.Context) {
	hint := ""
	switch {
	case m.faultText() != "":
		hint = m.faultText()
	case !m.Ready:
		hint = "正在准备识别器…"
	case m.Listening:
		hint = "正在听"
	default:
		hint = emptyHint
	}
	// Same bottom stack as a listening stage: hint where the current line
	// sits, primary on the same lower edge as 「停止」.
	g.Column(c).Grow(1).Gap(GapBlock).Children(func() {
		g.Spacer(c)
		if hint != "" {
			col := colorHist
			face := TypeEmpty
			if m.faultText() != "" {
				col = colorOnDark
				face = TypeBody
			}
			text(c, hint, face, col).Shrink(0)
		}
		g.Row(c).Shrink(0).Children(func() {
			m.stageAction(c)
		})
	})
}

func (m *Model) faultText() string {
	if m == nil || m.Fault == "" {
		return ""
	}
	return fault.UserMessage(m.Fault)
}

func (m *Model) stageAction(c *g.Context) {
	if m.Listening {
		if paintPrimaryLive(c, "listen", "停止", false) {
			m.toggleListen()
		}
		return
	}
	if m.Fault != "" {
		if action := fault.Action(m.Fault); action != "" {
			if paintPrimaryLive(c, "fault-"+m.Fault, action, m.Downloading) {
				m.faultAction(m.Fault)
			}
			return
		}
	}
	disabled := !m.Ready || m.selectedID() == ""
	if paintPrimaryLive(c, "listen", "开始听", disabled) {
		m.toggleListen()
	}
}

func (m *Model) history(c *g.Context, lines []pipeline.Line) {
	reveal := len(lines) > 6 && m.follow != len(m.Lines)
	build := func() {
		for i, line := range lines {
			when := ""
			if line.Start > 0 || line.End > 0 {
				when = stamp(line.Start)
			}
			row := g.Row(c).Gap(GapInline).AlignItems(g.Baseline).Shrink(0)
			if when != "" {
				row = row.Tooltip(when)
			}
			row = row.Children(func() {
				if when != "" && row.Hovered() {
					text(c, when, TypeCaption, colorHist).Shrink(0)
				}
				text(c, line.Text, TypeHist, colorHist).Grow(1)
			})
			if reveal && i == len(lines)-1 {
				row.ScrollIntoView()
			}
		}
	}
	if len(lines) > 6 {
		g.Scroll(c).Height(200).Shrink(0).Gap(GapInline).Children(build)
		m.follow = len(m.Lines)
		return
	}
	g.Column(c).Gap(GapInline).Shrink(0).Children(build)
}

func (m *Model) currentLine(c *g.Context, line pipeline.Line) {
	g.Column(c).Gap(Space1).Shrink(0).Children(func() {
		if label := langLabel(line.Language); label != "" {
			text(c, label, TypeCurrentMeta, colorHist)
		}
		face := TypeCurrent
		if len([]rune(line.Text)) > 42 {
			face = TypeCurrentLong
		}
		text(c, line.Text, face, colorOnDark)
	})
}

func (m *Model) chooseWAV(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	m.WAV = path
	m.FileName = filepath.Base(path)
	m.DockNote = ""
	m.clearFault(fault.FileInvalid)
}

func (m *Model) browse() {
	win := m.win
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:  win,
			Title:   "选择音频",
			Filters: []mygo.FileFilter{{Name: "WAV", Extensions: []string{"wav"}}},
		})
		if win == nil || m.closed.Load() {
			return
		}
		win.Update(func() {
			if err != nil || len(paths) == 0 {
				return
			}
			m.chooseWAV(paths[0])
		})
	}()
}

func (m *Model) copyCurrent() {
	text := ""
	if partial, ok := m.livePartial(); ok {
		text = partial.Text
	} else if lines := m.finals(); len(lines) > 0 {
		text = lines[len(lines)-1].Text
	}
	text = strings.TrimSpace(text)
	if text == "" || m.win == nil {
		return
	}
	mygo.Clipboard.WriteText(text)
	m.DockNote = "已复制当前句"
}

func (m *Model) footnote() string {
	if m.DockNote != "" {
		return m.DockNote
	}
	if m.Downloading {
		return "正在下载模型"
	}
	if m.Fault != "" {
		return ""
	}
	if !m.Ready {
		return "正在加载识别器"
	}
	if m.Listening {
		return ""
	}
	if len(m.finals()) > 0 {
		return "已停止"
	}
	return "可以开始听"
}
