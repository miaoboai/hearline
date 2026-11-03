package ui

import (
	"strings"

	"github.com/egoist/mygo"
	g "github.com/egoist/mygo/ui"
)

func (m *Model) view(c *g.Context) {
	if c == nil || m == nil {
		return
	}
	t := c.Theme()
	g.Column(c).Fill().Padding(20).Gap(12).Children(func() {
		g.Row(c).Gap(12).Children(func() {
			g.Text(c, "Hearline").FontSize(22).Bold()
			label := "CPU · SenseVoice-Small Q8"
			if m.Engine != "" {
				label = "CPU · " + m.Engine
			}
			g.Text(c, label).FontSize(13).TextColor(t.TextMuted)
		})
		g.Text(c, m.Status).FontSize(13).TextColor(t.TextMuted)
		if m.LLMStatus != "" {
			g.Text(c, "LLM  "+m.LLMStatus).FontSize(12).TextColor(t.TextMuted)
		}

		g.Text(c, "音源").FontSize(12).TextColor(t.TextMuted)
		if len(m.Labels) == 0 {
			g.Text(c, "没有实时音源。Linux 需要 PulseAudio 或 PipeWire（pipewire-pulse）；也可以播放 WAV。").FontSize(13)
		} else {
			g.Select(c.Key("source"), &m.Selected, m.Labels).MinWidth(420)
		}
		if m.DeviceNote != "" {
			g.Text(c, m.DeviceNote).FontSize(12).TextColor(t.TextMuted)
		}
		g.Row(c).Gap(8).Children(func() {
			if g.Button(c.Key("refresh"), "刷新音源").Clicked() {
				m.refresh(nil)
			}
			start := g.PrimaryButton(c.Key("start"), "开始听")
			if !m.Ready || m.Listening {
				start = start.Disabled(true)
			}
			if start.Clicked() {
				m.start(m.selectedID())
			}
			stop := g.Button(c.Key("stop"), "停止")
			if !m.Listening {
				stop = stop.Disabled(true)
			}
			if stop.Clicked() {
				m.stop()
				m.Status = "已停止"
			}
		})

		g.Text(c, "WAV 文件").FontSize(12).TextColor(t.TextMuted)
		g.TextInput(c.Key("wav"), &m.WAV).Placeholder("路径，例如 models/.../test_wavs/zh.wav").MinWidth(420)
		g.Row(c).Gap(8).Children(func() {
			if g.Button(c.Key("browse"), "浏览…").Clicked() {
				m.browse()
			}
			play := g.Button(c.Key("play"), "播放 WAV")
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

		g.Scroll(c).Grow(1).MinHeight(220).Padding(16).Radius(12).Background(g.Hex("#111318")).Children(func() {
			g.Column(c).Gap(8).Children(func() {
				if len(m.Lines) == 0 {
					g.Text(c, "字幕会显示在这里").FontSize(22).TextColor(g.Hex("#a1a1aa"))
					return
				}
				for i, line := range m.Lines {
					text := line.Text
					if line.Language != "" {
						text = line.Language + "  " + text
					}
					if line.Partial {
						text = "… " + text
					}
					size := float32(16)
					color := g.Hex("#d4d4d8")
					el := g.Text(c, text)
					if i == len(m.Lines)-1 {
						size = 28
						color = g.Hex("#fafafa")
						el = el.Bold()
					}
					el.FontSize(size).TextColor(color)
				}
			})
		})
	})
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
