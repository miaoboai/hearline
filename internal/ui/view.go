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

func (m *Model) view(c *g.Context) {
	if c == nil || m == nil {
		return
	}
	c.SetTheme(hearlineTheme())
	m.shortcuts(c)
	g.Column(c).Fill().Padding(24).Gap(14).Background(colorGround).Children(func() {
		m.header(c)
		switch m.Panel {
		case panelSettings:
			m.settingsPage(c)
		case panelNotes:
			m.notesPage(c)
		default:
			m.listenPage(c)
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
	g.Row(c).AlignItems(g.Center).Gap(16).Shrink(0).Children(func() {
		g.Column(c).Gap(1).Shrink(0).Children(func() {
			g.Text(c, "Hearline").FontSize(20).FontWeight(650).TextColor(colorInk)
			g.Text(c, "本地听写").FontSize(12).TextColor(colorQuiet)
		})
		m.sessionChip(c)
		g.Spacer(c)
		if m.headerButton(c, "notes", "纪要", m.Panel == panelNotes) {
			if m.Panel == panelNotes {
				m.Panel = panelListen
			} else {
				m.openNotes()
			}
		}
		if m.headerButton(c, "settings", "设置", m.Panel == panelSettings) {
			if m.Panel == panelSettings {
				m.Panel = panelListen
			} else {
				m.Panel = panelSettings
				m.ConfirmQuit = false
			}
		}
		m.exportMenu(c)
	})
}

func (m *Model) headerButton(c *g.Context, key, label string, on bool) bool {
	btn := g.Button(c.Key(key), label).Shrink(0).MinHeight(36)
	if on {
		btn = btn.Background(g.Hex("#E8EAEE"))
	}
	return btn.Clicked()
}

func (m *Model) exportMenu(c *g.Context) {
	btn := g.PopoverMenuButton(c.Key("export"), "导出", func(menu *g.Menu) {
		if menu.Item("文本 (.txt)").Chosen() {
			m.exportKind("txt")
		}
		if menu.Item("字幕 (.srt)").Chosen() {
			m.exportKind("srt")
		}
		item := menu.Item("纪要 (.md)")
		if docEmpty(m.Doc) {
			item = item.Disabled(true)
		}
		if item.Chosen() {
			m.exportKind("md")
		}
	}).Shrink(0).MinHeight(36)
	if len(m.finals()) == 0 {
		btn = btn.Disabled(true)
	}
	_ = btn
}

func (m *Model) sessionChip(c *g.Context) {
	switch {
	case m.Fault != "" && !m.Listening:
		g.Text(c, "出错").FontSize(13).TextColor(colorMuted).Shrink(0)
	case !m.Ready && m.Fault == "":
		g.Text(c, "准备中").FontSize(13).TextColor(colorQuiet).Shrink(0)
	case m.Listening:
		g.Row(c).Gap(8).AlignItems(g.Center).Shrink(0).Children(func() {
			g.Box(c).Size(8, 8).Radius(4).Background(colorAccent)
			g.Text(c, "正在听 · "+m.clock()).FontSize(13).FontWeight(600).TextColor(colorInk)
		})
	case len(m.finals()) > 0:
		g.Text(c, fmt.Sprintf("已停止 · 共 %d 句", len(m.finals()))).FontSize(13).TextColor(colorQuiet).Shrink(0)
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
	g.Column(c).Grow(1).Gap(12).Children(func() {
		m.stage(c)
		if m.ConfirmQuit {
			m.confirmQuit(c)
		}
		m.dock(c)
	})
}

func (m *Model) confirmQuit(c *g.Context) {
	g.Row(c).Shrink(0).Padding(12, 16).Gap(12).Radius(14).AlignItems(g.Center).
		Background(colorCard).Border(1, colorLine).
		Children(func() {
			g.Text(c, "停止并退出？").FontSize(14).TextColor(colorInk).Grow(1)
			if g.Button(c.Key("stay"), "继续听").Shrink(0).MinHeight(36).Clicked() {
				m.ConfirmQuit = false
			}
			if g.Button(c.Key("quit"), "退出").Shrink(0).MinHeight(36).Clicked() {
				m.allowClose = true
				m.stop()
				if m.win != nil {
					m.win.Close()
				}
			}
		})
}

func (m *Model) stage(c *g.Context) {
	g.Column(c).Grow(1).MinHeight(400).Padding(22, 28, 26, 28).Radius(22).Gap(16).
		Background(colorStage).
		Shadow(0, 16, 40, 0, g.RGBA(21, 23, 28, 0.16)).
		Children(func() {
			if code := m.Fault; code != "" && fault.UserMessage(code) != "" {
				m.faultBar(c, code)
			}
			finals := m.finals()
			partial, hasPartial := m.livePartial()
			if len(finals) == 0 && !hasPartial {
				m.emptyStage(c)
				return
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
				g.Text(c, partial.Text).FontSize(26).Italic().TextColor(g.Hex("#8E939E")).Shrink(0)
			} else if current != nil {
				m.currentLine(c, *current)
			}
		})
}

func (m *Model) emptyStage(c *g.Context) {
	title, sub := "选一个音源，开始听", "字幕会跟着声音出现"
	switch {
	case !m.Ready && m.Fault == "":
		title, sub = "正在准备识别器…", "首次约需几秒，只在本机"
	case m.Listening:
		title, sub = "正在听…", "开口后字幕会出现在这里"
	case m.Fault != "":
		return
	}
	g.Column(c).Grow(1).AlignItems(g.Center).Justify(g.Center).Gap(8).Children(func() {
		g.Text(c, title).FontSize(18).TextColor(g.Hex("#C5CAD3"))
		g.Text(c, sub).FontSize(14).TextColor(g.Hex("#6B7080"))
	})
}

func (m *Model) faultBar(c *g.Context, code string) {
	text := fault.UserMessage(code)
	action := fault.Action(code)
	g.Row(c).Shrink(0).Padding(10, 14).Gap(12).Radius(12).AlignItems(g.Center).
		Background(g.Hex("#2A1A16")).
		Children(func() {
			g.Text(c, text).FontSize(14).TextColor(g.Hex("#F6D2C6")).Grow(1)
			if action == "" {
				return
			}
			if g.Button(c.Key("fault-"+code), action).Shrink(0).MinHeight(36).Clicked() {
				m.faultAction(code)
			}
		})
}

func (m *Model) history(c *g.Context, lines []pipeline.Line) {
	reveal := len(lines) > 6 && m.follow != len(m.Lines)
	build := func() {
		for i, line := range lines {
			row := g.Row(c).Gap(12).AlignItems(g.Baseline).Shrink(0).Children(func() {
				if line.Start > 0 || line.End > 0 {
					g.Text(c, stamp(line.Start)).FontSize(12).TextColor(colorTime).Shrink(0)
				}
				g.Text(c, line.Text).FontSize(15).TextColor(colorHist).Grow(1)
			})
			if reveal && i == len(lines)-1 {
				row.ScrollIntoView()
			}
		}
	}
	if len(lines) > 6 {
		g.Scroll(c).Height(200).Shrink(0).Gap(10).Children(build)
		m.follow = len(m.Lines)
		return
	}
	g.Column(c).Gap(10).Shrink(0).Children(build)
}

func (m *Model) currentLine(c *g.Context, line pipeline.Line) {
	g.Column(c).Gap(8).Shrink(0).Children(func() {
		if meta := lineMeta(line); meta != "" {
			g.Text(c, meta).FontSize(12).TextColor(colorTime)
		}
		size := float32(32)
		if len([]rune(line.Text)) > 42 {
			size = 26
		}
		g.Text(c, line.Text).FontSize(size).FontWeight(650).TextColor(colorOnDark)
	})
}

func (m *Model) dock(c *g.Context) {
	g.Column(c).Shrink(0).Padding(14, 16).Gap(12).Radius(16).
		Background(colorCard).Border(1, colorLine).
		Shadow(0, 8, 24, 0, g.RGBA(21, 23, 28, 0.05)).
		Children(func() {
			g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
				if len(m.Labels) == 0 {
					g.Text(c, "没找到系统声，可从文件听写").FontSize(14).TextColor(colorMuted).Grow(1)
				} else {
					sel := g.Select(c.Key("source"), &m.Selected, m.Labels).Grow(1).MinWidth(240)
					if m.Focus == focusSource {
						sel.Focus()
						m.Focus = ""
					}
				}
				if g.Button(c.Key("refresh"), "刷新").Shrink(0).MinHeight(36).Clicked() {
					m.refresh(nil)
					if m.Fault == fault.CaptureUnavailable && len(m.Labels) > 0 {
						m.clearFault(fault.CaptureUnavailable)
					}
				}
			})
			g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
				label := "开始听"
				if m.Listening {
					label = "停止"
				}
				var btn g.Element
				if m.Listening || (m.Ready && len(m.Labels) > 0) {
					btn = g.PrimaryButton(c.Key("listen"), label).Shrink(0).MinHeight(40)
				} else {
					btn = g.Button(c.Key("listen"), label).Shrink(0).MinHeight(40).Disabled(true)
				}
				if btn.Clicked() {
					m.toggleListen()
				}
				if !m.Listening && len(m.finals()) > 0 {
					if g.Button(c.Key("clear"), "清空").Shrink(0).MinHeight(36).Clicked() {
						m.clearLines()
					}
				}
				more := "更多"
				if m.More {
					more = "收起"
				}
				if g.Button(c.Key("more"), more).Shrink(0).MinHeight(36).Clicked() {
					m.More = !m.More
				}
				g.Spacer(c)
				if note := m.footnote(); note != "" {
					g.Text(c, note).FontSize(13).TextColor(colorQuiet)
				}
			})
			if m.More {
				m.fileRow(c)
			}
		})
}

func (m *Model) fileRow(c *g.Context) {
	g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
		if g.Button(c.Key("browse"), "从文件…").Shrink(0).MinHeight(36).Clicked() {
			m.browse()
		}
		name := m.FileName
		if name == "" {
			name = "还没有选择文件"
		}
		el := g.Text(c, name).FontSize(14).TextColor(colorMuted).Grow(1)
		if m.WAV != "" {
			el.Tooltip(m.WAV)
		}
		play := g.Button(c.Key("play"), "听这个文件").Shrink(0).MinHeight(36)
		if !m.Ready || m.Listening || strings.TrimSpace(m.WAV) == "" {
			play = play.Disabled(true)
		}
		if play.Clicked() {
			m.listenFile()
		}
	})
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
			m.WAV = paths[0]
			m.FileName = filepath.Base(paths[0])
			m.More = true
			m.DockNote = ""
			m.clearFault(fault.FileInvalid)
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
