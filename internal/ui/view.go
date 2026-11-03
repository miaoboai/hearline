package ui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/notes"
	"github.com/miaoboai/hearline/internal/pipeline"
)

func (m *Model) view(c *g.Context) {
	if c == nil || m == nil {
		return
	}
	c.SetTheme(hearlineTheme())
	m.shortcuts(c)
	g.Column(c).Fill().Padding(28).Gap(16).Background(colorGround).Children(func() {
		m.header(c)
		switch m.Screen {
		case screenSettings:
			m.settingsPage(c)
		case screenAbout:
			m.aboutPage(c)
		case screenNotes:
			m.notesPage(c)
		default:
			m.listenPage(c)
		}
	})
}

func (m *Model) shortcuts(c *g.Context) {
	c.OnShortcut(g.Cmd, g.KeyL, m.toggleListen)
	c.OnShortcut(g.Cmd, g.KeyE, m.export)
	c.OnShortcut(g.Cmd, g.KeyComma, func() { m.Screen = screenSettings })
	c.OnShortcut(0, g.KeyEscape, func() {
		if m.Listening {
			m.stop()
			return
		}
		if m.Screen != screenListen {
			m.Screen = screenListen
		}
	})
}

func (m *Model) header(c *g.Context) {
	g.Row(c).AlignItems(g.Center).Gap(18).Shrink(0).Children(func() {
		g.Column(c).Gap(1).Shrink(0).Children(func() {
			g.Text(c, "Hearline").FontSize(20).FontWeight(650).TextColor(colorInk)
			g.Text(c, "本地听写").FontSize(12).TextColor(colorQuiet)
		})
		g.Row(c).Gap(16).AlignItems(g.Center).Shrink(0).Children(func() {
			m.nav(c, "听写", screenListen)
			m.nav(c, "设置", screenSettings)
			m.nav(c, "关于", screenAbout)
		})
		g.Spacer(c)
		ok, reason := notes.Available(m.llm)
		if !ok {
			g.Text(c, reason).FontSize(12).TextColor(colorQuiet).Shrink(0)
		}
		notesBtn := g.Button(c.Key("notes"), "纪要").Shrink(0)
		if !ok {
			notesBtn = notesBtn.Disabled(true)
		}
		if notesBtn.Clicked() {
			m.openNotes()
		}
		exp := g.Button(c.Key("export"), "导出").Shrink(0)
		if len(m.finals()) == 0 {
			exp = exp.Disabled(true)
		}
		if exp.Clicked() {
			m.export()
		}
		if m.ExportNote != "" {
			g.Text(c, m.ExportNote).FontSize(12).TextColor(colorQuiet).Shrink(0)
		}
	})
}

func (m *Model) nav(c *g.Context, label string, screen int) {
	color := colorQuiet
	weight := 500
	if m.Screen == screen {
		color = colorInk
		weight = 650
	}
	g.Text(c, label).FontSize(14).FontWeight(weight).TextColor(color).Cursor(g.CursorPointer).OnClick(func() {
		m.Screen = screen
	})
}

func (m *Model) listenPage(c *g.Context) {
	g.Column(c).Grow(1).Gap(14).Children(func() {
		if code := m.Fault; code != "" {
			m.faultBar(c, code)
		}
		m.stage(c)
		m.controls(c)
	})
}

func (m *Model) faultBar(c *g.Context, code string) {
	text := ""
	action := ""
	switch code {
	case fault.ModelMissing:
		text = "还没找到识别模型"
		action = "填写路径"
	case fault.CaptureUnavailable:
		text = "当前音源听不到"
		action = "更换音源"
	case fault.DecodeBacklog:
		text = "识别有点跟不上，句子会稍晚出现"
	default:
		return
	}
	g.Row(c).Shrink(0).Padding(12, 16).Gap(12).Radius(14).AlignItems(g.Center).
		Background(colorCard).Border(1, colorLine).
		Children(func() {
			g.Text(c, text).FontSize(14).TextColor(colorInk).Grow(1)
			if action == "" {
				return
			}
			if g.Button(c.Key("fault-"+code), action).Shrink(0).Clicked() {
				m.faultAction(code)
			}
		})
}

func (m *Model) stage(c *g.Context) {
	g.Column(c).Grow(1).MinHeight(420).Padding(22, 32, 28, 32).Radius(22).
		Background(colorStage).
		Shadow(0, 16, 40, 0, g.RGBA(21, 23, 28, 0.16)).
		Children(func() {
			g.Row(c).AlignItems(g.Center).Shrink(0).Children(func() {
				if m.Listening {
					g.Text(c, "正在听").FontSize(12).TextColor(g.Hex("#9AA0AA"))
				}
				g.Spacer(c)
				if n := len(m.finals()); n > 0 {
					g.Text(c, fmt.Sprintf("%d 句", n)).FontSize(12).TextColor(g.Hex("#6B7080"))
				}
			})
			finals := m.finals()
			partial, hasPartial := m.livePartial()
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
			g.Spacer(c)
			g.Column(c).Gap(18).Shrink(0).Children(func() {
				if len(history) > 0 {
					m.history(c, history)
				}
				switch {
				case hasPartial:
					g.Text(c, partial.Text).FontSize(26).Italic().TextColor(g.Hex("#8E939E")).Shrink(0)
				case current != nil:
					m.currentLine(c, *current)
				default:
					hint := "听写会显示在这里"
					if !m.Ready && m.Fault == "" {
						hint = "正在准备"
					}
					g.Text(c, hint).FontSize(15).TextColor(g.Hex("#6B7080")).Shrink(0)
				}
			})
		})
}

func (m *Model) history(c *g.Context, lines []pipeline.Line) {
	reveal := len(lines) > 6 && m.follow != len(m.Lines)
	build := func() {
		for i, line := range lines {
			el := g.Text(c, line.Text).FontSize(16).TextColor(g.Hex("#B7BCC6"))
			if reveal && i == len(lines)-1 {
				el.ScrollIntoView()
			}
		}
	}
	if len(lines) > 6 {
		g.Scroll(c).Height(220).Shrink(0).Gap(12).Children(build)
		m.follow = len(m.Lines)
		return
	}
	g.Column(c).Gap(12).Shrink(0).Children(build)
}

func (m *Model) currentLine(c *g.Context, line pipeline.Line) {
	g.Column(c).Gap(8).Shrink(0).Children(func() {
		if meta := lineMeta(line); meta != "" {
			g.Text(c, meta).FontSize(12).TextColor(g.Hex("#8B909A"))
		}
		size := float32(32)
		if len([]rune(line.Text)) > 42 {
			size = 26
		}
		g.Text(c, line.Text).FontSize(size).FontWeight(650).TextColor(colorOnDark)
	})
}

func (m *Model) controls(c *g.Context) {
	g.Column(c).Shrink(0).Padding(14, 16).Gap(12).Radius(16).
		Background(colorCard).Border(1, colorLine).
		Shadow(0, 8, 24, 0, g.RGBA(21, 23, 28, 0.05)).
		Children(func() {
			g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
				if len(m.Labels) == 0 {
					g.Text(c, "没有实时音源").FontSize(14).TextColor(colorMuted).Grow(1)
				} else {
					sel := g.Select(c.Key("source"), &m.Selected, m.Labels).Grow(1).MinWidth(280)
					if m.Focus == focusSource {
						sel.Focus()
						m.Focus = ""
					}
				}
				if g.Button(c.Key("refresh"), "刷新").Shrink(0).Clicked() {
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
				if m.Listening || m.Ready {
					btn = g.PrimaryButton(c.Key("listen"), label).Shrink(0)
				} else {
					btn = g.Button(c.Key("listen"), label).Shrink(0).Disabled(true)
				}
				if btn.Clicked() {
					m.toggleListen()
				}
				g.Spacer(c)
				g.Text(c, shortcutHint()).FontSize(12).TextColor(colorQuiet)
			})
			g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
				field := g.TextInput(c.Key("wav"), &m.WAV).Placeholder("WAV 路径").Grow(1).MinWidth(240)
				if m.Focus == focusWAV {
					field.Focus()
					m.Focus = ""
				}
				if g.Button(c.Key("browse"), "浏览").Shrink(0).Clicked() {
					m.browse()
				}
				play := g.Button(c.Key("play"), "播放").Shrink(0)
				if !m.Ready || m.Listening {
					play = play.Disabled(true)
				}
				if play.Clicked() {
					path := strings.TrimSpace(m.WAV)
					if path == "" {
						m.Focus = focusWAV
						return
					}
					m.start("wav:" + path)
				}
			})
		})
}

func shortcutHint() string {
	mod := "Ctrl"
	if runtime.GOOS == "darwin" {
		mod = "⌘"
	}
	return mod + "+L 听写  ·  " + mod + "+E 导出"
}

func (m *Model) browse() {
	win := m.win
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:  win,
			Title:   "选择 WAV",
			Filters: []mygo.FileFilter{{Name: "WAV", Extensions: []string{"wav"}}},
		})
		if win == nil || m.closed.Load() {
			return
		}
		win.Update(func() {
			if err != nil {
				return
			}
			if len(paths) > 0 {
				m.WAV = paths[0]
			}
		})
	}()
}
