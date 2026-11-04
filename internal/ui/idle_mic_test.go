package ui

import (
	"errors"
	"testing"

	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/audio"
	"github.com/miaoboai/hearline/internal/fault"
)

func TestMicListLeavesIdleReady(t *testing.T) {
	const gap = "macOS 系统声采集尚未实现：v0.1 计划用 Core Audio process tap（系统输出环回）。当前请使用 WAV 音源（hearline transcribe --wav）"
	m := Snapshot(nil, "", "已准备", "")
	m.refresh(func() ([]audio.Source, error) {
		return nil, errors.New(gap)
	})
	if m.Fault != fault.CaptureUnavailable {
		t.Fatalf("gap fault %q", m.Fault)
	}
	broken := g.NewTester(m.view, frameWidth, frameHeight)
	if broken.HasText("无法读取这个文件") || exactText(broken, "重选") {
		t.Fatalf("missing loopback still looks like a bad file: %v", broken.Texts())
	}
	if !broken.HasText("音源不可用") {
		t.Fatalf("gap texts %v", broken.Texts())
	}

	m.refresh(func() ([]audio.Source, error) {
		return []audio.Source{{
			ID:   "source:BuiltInMicrophoneDevice",
			Name: "麦克风 · MacBook Air麦克风",
			Kind: "mic",
		}}, nil
	})
	if m.Fault != "" {
		t.Fatalf("mic list left fault %q note %q", m.Fault, m.DeviceNote)
	}
	if m.selectedID() != "source:BuiltInMicrophoneDevice" {
		t.Fatalf("selected %q label %q", m.selectedID(), m.Selected)
	}
	idle := g.NewTester(m.view, frameWidth, frameHeight)
	if exactText(idle, "出错") || exactText(idle, "重选") || idle.HasText("无法读取这个文件") {
		t.Fatalf("idle still broken: %v", idle.Texts())
	}
	if !exactText(idle, "开始听") || !idle.HasText(emptyHint) {
		t.Fatalf("idle texts %v", idle.Texts())
	}
	want := g.Hex(HexAccent)
	if !hasNear(idle.Image(), int(want.R), int(want.G), int(want.B), 8) {
		t.Fatal("开始听 did not paint the enabled accent")
	}
	if err := idle.Click("开始听"); err != nil {
		t.Fatal(err)
	}

	m.Panel = panelSettings
	settings := g.NewTester(m.view, frameWidth, frameHeight)
	if !settings.HasText("麦克风 · MacBook Air麦克风") {
		t.Fatalf("settings hid the mic: %v", settings.Texts())
	}

	m.Panel = panelListen
	m.Fault = fault.Permission
	denied := g.NewTester(m.view, frameWidth, frameHeight)
	if !denied.HasText(fault.UserMessage(fault.Permission)) || exactText(denied, "重选") || denied.HasText("无法读取这个文件") {
		t.Fatalf("permission texts %v", denied.Texts())
	}
	if !exactText(denied, "开始听") || fault.Action(fault.Permission) != "" {
		t.Fatal("permission invented a button")
	}

	m.Fault = fault.FileInvalid
	m.chooseWAV("  /tmp/zh.wav  ")
	if m.Fault != "" || m.WAV != "/tmp/zh.wav" || m.FileName != "zh.wav" {
		t.Fatalf("choose wav fault %q path %q name %q", m.Fault, m.WAV, m.FileName)
	}
}
