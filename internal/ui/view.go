package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo"
	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/pipeline"
)

func (m *Model) view(c *g.Context) {
	if c == nil || m == nil {
		return
	}
	c.SetTheme(hearlineTheme())
	g.Column(c).Fill().Padding(28).Gap(18).Background(colorGround).Children(func() {
		m.header(c)
		m.caption(c)
		m.controls(c)
	})
}

func (m *Model) header(c *g.Context) {
	g.Row(c).AlignItems(g.Center).Gap(12).Shrink(0).Children(func() {
		g.Box(c).Size(16, 16).Radius(5).Background(colorAccent).Shrink(0)
		g.Column(c).Gap(2).Shrink(0).Children(func() {
			g.Text(c, "Hearline").FontSize(22).FontWeight(650).TextColor(colorInk)
			g.Text(c, "本地听写").FontSize(13).TextColor(colorQuiet)
		})
		g.Spacer(c)
		g.Row(c).Gap(8).AlignItems(g.Center).Shrink(0).Children(func() {
			pill(c, engineChip(m.Engine, m.Status), colorInk, colorCard, true)
			if label := cpuChip(m.Status); label != "" {
				pill(c, label, colorMuted, colorCard, true)
			}
			if label := llmChip(m.LLMStatus); label != "" {
				pill(c, label, colorMuted, colorCard, true)
			}
			if m.Listening {
				pill(c, "正在听", colorCard, colorAccent, false)
			}
		})
	})
}

func (m *Model) caption(c *g.Context) {
	g.Column(c).Grow(1).MinHeight(340).Padding(22, 32).Gap(12).Radius(22).
		Background(colorStage).
		Shadow(0, 16, 40, 0, g.RGBA(21, 23, 28, 0.16)).
		Children(func() {
			g.Row(c).AlignItems(g.Center).Gap(8).Shrink(0).Children(func() {
				dot := colorQuiet
				label := "字幕"
				if m.Listening || len(m.Lines) > 0 {
					dot = colorAccent
				}
				if m.Listening {
					label = "正在听"
				}
				g.Box(c).Size(8, 8).Radius(4).Background(dot).Shrink(0)
				g.Text(c, label).FontSize(12).FontWeight(600).TextColor(g.Hex("#C8CDD6"))
				g.Spacer(c)
				if n := len(m.Lines); n > 0 {
					g.Text(c, fmt.Sprintf("%d 句", n)).FontSize(12).TextColor(colorTime)
				}
			})
			g.Column(c).Grow(1).Justify(g.Center).Gap(18).Children(func() {
				if len(m.Lines) == 0 {
					g.Column(c).AlignItems(g.Center).Gap(8).Children(func() {
						g.Text(c, "放一段声音").FontSize(30).FontWeight(650).TextColor(colorOnDark)
						g.Text(c, "字幕会跟着声音出现").FontSize(15).TextColor(g.Hex("#9AA0AA"))
					})
					return
				}
				m.history(c)
				currentLine(c, m.Lines[len(m.Lines)-1])
			})
		})
}

func (m *Model) history(c *g.Context) {
	if len(m.Lines) < 2 {
		return
	}
	hist := m.Lines[:len(m.Lines)-1]
	if len(hist) > 4 {
		reveal := m.follow != len(m.Lines)
		g.Scroll(c).Height(188).Gap(10).Shrink(0).Children(func() {
			for i, line := range hist {
				row := historyRow(c, line)
				if reveal && i == len(hist)-1 {
					row.ScrollIntoView()
				}
			}
		})
		m.follow = len(m.Lines)
	} else {
		g.Column(c).Gap(10).Shrink(0).Children(func() {
			for _, line := range hist {
				historyRow(c, line)
			}
		})
	}
	g.Box(c).Height(1).Shrink(0).Background(g.RGBA(255, 255, 255, 0.08))
}

func historyRow(c *g.Context, line pipeline.Line) g.Element {
	row := g.Row(c).Gap(12).AlignItems(g.Start)
	row.Children(func() {
		g.Text(c, stamp(line.Start)).FontSize(12).TextColor(colorTime).Width(52).Shrink(0)
		g.Text(c, langOrDash(line)).FontSize(12).FontWeight(600).TextColor(g.Hex("#9AA0AA")).Width(72).Shrink(0)
		text := line.Text
		if line.Partial {
			text = "… " + text
		}
		g.Text(c, text).FontSize(15).TextColor(colorHist).Grow(1)
	})
	return row
}

func currentLine(c *g.Context, line pipeline.Line) {
	g.Column(c).Gap(10).Shrink(0).Children(func() {
		g.Row(c).Gap(10).AlignItems(g.Center).Children(func() {
			label := langLabel(line.Language)
			if line.Partial {
				label = "听写中"
			}
			if label != "" {
				pill(c, label, colorOnDark, g.Hex("#2C313A"), false)
			}
			g.Text(c, stamp(line.Start)).FontSize(13).TextColor(colorTime)
		})
		size := float32(32)
		if len([]rune(line.Text)) > 42 {
			size = 26
		}
		el := g.Text(c, line.Text).FontSize(size).FontWeight(650).TextColor(colorOnDark)
		if line.Partial {
			el.TextColor(g.Hex("#E4E6EB"))
		}
	})
}

func (m *Model) controls(c *g.Context) {
	g.Column(c).Shrink(0).Padding(16, 18).Gap(14).Radius(16).
		Background(colorCard).Border(1, colorLine).
		Shadow(0, 1, 2, 0, g.RGBA(21, 23, 28, 0.04)).
		Shadow(0, 10, 28, 0, g.RGBA(21, 23, 28, 0.06)).
		Children(func() {
			g.Column(c).Gap(8).Children(func() {
				sectionLabel(c, "音源")
				g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
					if len(m.Labels) == 0 {
						g.Text(c, "没有实时音源，可以播放 WAV").FontSize(14).TextColor(colorInk).Grow(1)
					} else {
						g.Select(c.Key("source"), &m.Selected, m.Labels).Grow(1).MinWidth(280)
					}
					if g.Button(c.Key("refresh"), "刷新").Shrink(0).Clicked() {
						m.refresh(nil)
					}
				})
				if note := deviceNote(m.DeviceNote); note != "" {
					g.Text(c, note).FontSize(12).TextColor(colorQuiet).SingleLine()
				}
			})

			g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
				start := g.PrimaryButton(c.Key("start"), "开始听").Shrink(0)
				if !m.Ready || m.Listening {
					start = start.Disabled(true)
				}
				if start.Clicked() {
					m.start(m.selectedID())
				}
				stop := g.Button(c.Key("stop"), "停止").Shrink(0)
				if !m.Listening {
					stop = stop.Disabled(true)
				}
				if stop.Clicked() {
					m.stop()
					m.Status = "已停止"
				}
				g.Spacer(c)
				if m.Listening {
					pill(c, "正在听", colorCard, colorAccent, false)
				} else if s := footerStatus(m); s != "" {
					g.Text(c, s).FontSize(13).TextColor(colorQuiet).SingleLine().MaxWidth(420)
				}
			})

			g.Column(c).Gap(8).Children(func() {
				sectionLabel(c, "WAV")
				g.Row(c).Gap(8).AlignItems(g.Center).Children(func() {
					g.TextInput(c.Key("wav"), &m.WAV).Placeholder("路径，例如 models/…/zh.wav").Grow(1).MinWidth(240)
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
							m.Status = "请先填写 WAV 路径"
						} else {
							m.start("wav:" + path)
						}
					}
				})
			})
		})
}

func sectionLabel(c *g.Context, text string) {
	g.Text(c, text).FontSize(12).FontWeight(600).TextColor(colorQuiet)
}

func pill(c *g.Context, label string, fg, bg g.Color, border bool) {
	row := g.Row(c).Padding(4, 10).Radius(20).Background(bg).AlignItems(g.Center).Shrink(0)
	if border {
		row = row.Border(1, colorLine)
	}
	row.Children(func() {
		g.Text(c, label).FontSize(12).FontWeight(600).TextColor(fg).SingleLine()
	})
}

func engineChip(engine, status string) string {
	blob := strings.ToLower(engine + " " + status)
	switch {
	case strings.Contains(blob, "sensevoice"), strings.Contains(blob, "sense-voice"), strings.Contains(blob, "sense voice"):
		return "SenseVoice · Q8"
	case strings.Contains(blob, "mock"):
		return "Mock"
	case strings.Contains(blob, "qwen"):
		return "Qwen3"
	case strings.TrimSpace(engine) != "":
		return engine
	default:
		return "识别器"
	}
}

func cpuChip(status string) string {
	if n := threadCount(status); n != "" {
		return "CPU · " + n + " 线程"
	}
	if strings.Contains(strings.ToUpper(status), "CPU") {
		return "CPU"
	}
	return ""
}

func threadCount(s string) string {
	low := strings.ToLower(s)
	i := strings.Index(low, "thread")
	if i < 0 {
		return ""
	}
	j := i
	for j > 0 && (low[j-1] == ' ' || low[j-1] == '·' || low[j-1] == '-' || low[j-1] == '=') {
		j--
	}
	end := j
	for j > 0 && low[j-1] >= '0' && low[j-1] <= '9' {
		j--
	}
	if j == end {
		return ""
	}
	return low[j:end]
}

func llmChip(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	if strings.Contains(s, "远程") {
		return "LLM · 远程"
	}
	return "LLM · 本地"
}

func deviceNote(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	low := strings.ToLower(s)
	if strings.Contains(low, "pulse") || strings.Contains(low, "pipewire") || strings.Contains(low, "connection refused") {
		return "没有连上系统声，WAV 仍然可以播放"
	}
	return s
}

func footerStatus(m *Model) string {
	if m == nil || m.Listening {
		return ""
	}
	s := strings.TrimSpace(m.Status)
	switch {
	case s == "":
		return ""
	case strings.HasPrefix(s, "已加载"), strings.Contains(s, "thread"), strings.Contains(s, "SenseVoice"):
		return "可以开始听"
	case strings.Contains(s, "正在加载"):
		return "正在加载识别器"
	default:
		return s
	}
}

func langLabel(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "":
		return ""
	case "zh", "zh-cn", "zh-hans":
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
		return code
	}
}

func langOrDash(line pipeline.Line) string {
	if line.Partial && line.Language == "" {
		return "听写中"
	}
	if label := langLabel(line.Language); label != "" {
		return label
	}
	return "—"
}

func stamp(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	ms := d.Milliseconds()
	return fmt.Sprintf("%02d:%02d.%d", ms/60000, (ms/1000)%60, (ms%1000)/100)
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
				m.Status = err.Error()
				return
			}
			if len(paths) > 0 {
				m.WAV = paths[0]
			}
		})
	}()
}
