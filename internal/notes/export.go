package notes

import (
	"fmt"
	"strings"
	"time"

	"github.com/miaoboai/hearline/internal/pipeline"
)

// Plain is the finished sentences, one per line.
func Plain(lines []pipeline.Line) string {
	finals := Finals(lines)
	var b strings.Builder
	for _, line := range finals {
		fmt.Fprintln(&b, strings.TrimSpace(line.Text))
	}
	return b.String()
}

// SRT is the finished sentences as SubRip cues.
func SRT(lines []pipeline.Line) string {
	finals := Finals(lines)
	var b strings.Builder
	for i, line := range finals {
		end := line.End
		if end <= line.Start {
			end = line.Start + time.Second
		}
		fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n\n", i+1, srtStamp(line.Start), srtStamp(end), strings.TrimSpace(line.Text))
	}
	return b.String()
}

func srtStamp(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	ms := d.Milliseconds()
	return fmt.Sprintf("%02d:%02d:%02d,%03d", ms/3600000, (ms/60000)%60, (ms/1000)%60, ms%1000)
}
