package ui

import (
	"fmt"
	"strings"
	"time"

	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/pipeline"
)

const (
	privacyAudio  = "听写在本机完成，音频不会上传。"
	privacyRemote = "若纪要选用远程服务，只会发送已识别的文字，不会发送音频。"
	privacyOff    = "当前纪要未启用远程"
)

func privacyBlock(c *g.Context, remoteOn bool) {
	g.Text(c, privacyAudio).FontSize(14).TextColor(colorInk)
	if remoteOn {
		g.Text(c, privacyRemote).FontSize(14).TextColor(colorMuted)
		return
	}
	g.Text(c, privacyOff).FontSize(14).TextColor(colorQuiet)
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

func lineMeta(line pipeline.Line) string {
	var bits []string
	if label := langLabel(line.Language); label != "" {
		bits = append(bits, label)
	}
	if line.Start > 0 || line.End > 0 {
		bits = append(bits, stamp(line.Start))
	}
	return strings.Join(bits, " · ")
}

func stamp(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	ms := d.Milliseconds()
	return fmt.Sprintf("%02d:%02d.%d", ms/60000, (ms/1000)%60, (ms%1000)/100)
}
