package audio

import (
	_ "embed"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/miaoboai/hearline/internal/fault"
)

//go:embed Info.plist
var infoPlist string

func TestMicSourcesDefaultFirstAndNoFakeLoopback(t *testing.T) {
	got := micSources([]inputDevice{
		{ID: "usb-mic", Name: "USB Audio", Default: false},
		{ID: "BuiltInMicrophoneDevice", Name: "MacBook Air麦克风", Default: true},
		{ID: "  ", Name: "blank"},
		{ID: "headset", Name: "  "},
	})
	if len(got) != 3 {
		t.Fatalf("sources %#v", got)
	}
	if got[0].Kind != "mic" || got[0].ID != "source:BuiltInMicrophoneDevice" {
		t.Fatalf("default %+v", got[0])
	}
	if got[0].Name != "麦克风 · MacBook Air麦克风" {
		t.Fatalf("name %q", got[0].Name)
	}
	if got[2].Name != "麦克风 · headset" {
		t.Fatalf("empty name %+v", got[2])
	}
	ordered := micSources([]inputDevice{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
		{ID: "c", Name: "C", Default: true},
	})
	if len(ordered) != 3 || ordered[0].ID != "source:c" || ordered[1].ID != "source:a" || ordered[2].ID != "source:b" {
		t.Fatalf("default last %+v", ordered)
	}
	already := micSources([]inputDevice{
		{ID: "a", Name: "A", Default: true},
		{ID: "b", Name: "B"},
	})
	if already[0].ID != "source:a" || already[1].ID != "source:b" {
		t.Fatalf("default first %+v", already)
	}
	for _, s := range got {
		if s.Kind != "mic" {
			t.Fatalf("invented kind %q", s.Kind)
		}
	}
	if len(micSources(nil)) != 0 {
		t.Fatal("nil devices produced a source")
	}
}

func TestDeviceUIDRoundTrip(t *testing.T) {
	const uid = "BuiltInMicrophoneDevice"
	raw := make([]byte, 256)
	if err := putDeviceUID(raw, uid); err != nil {
		t.Fatal(err)
	}
	if got := deviceUID(raw); got != uid {
		t.Fatalf("uid %q", got)
	}
	if err := putDeviceUID(raw, ""); err == nil {
		t.Fatal("empty uid accepted")
	}
	if err := putDeviceUID(make([]byte, 4), "microphone"); err == nil {
		t.Fatal("overlong uid accepted")
	}
}

func TestFloat32FromPCM(t *testing.T) {
	s16 := []byte{0x00, 0x40, 0x00, 0xc0} // 0.5, -0.5
	mono := float32FromPCM(s16, 2, 1)
	if len(mono) != 2 || math.Abs(float64(mono[0]-0.5)) > 1e-4 || math.Abs(float64(mono[1]+0.5)) > 1e-4 {
		t.Fatalf("s16 %#v", mono)
	}
	// Stereo pair of 0.5 and -0.5 downmixes to 0.
	stereo := float32FromPCM(s16, 2, 2)
	if len(stereo) != 1 || stereo[0] != 0 {
		t.Fatalf("downmix %#v", stereo)
	}
	f32 := make([]byte, 4)
	binaryPutF32(f32, 0.25)
	got := float32FromPCM(f32, 4, 1)
	if len(got) != 1 || math.Abs(float64(got[0]-0.25)) > 1e-6 {
		t.Fatalf("f32 %#v", got)
	}
	if float32FromPCM(nil, 3, 1) != nil {
		t.Fatal("odd width accepted")
	}
}

func binaryPutF32(b []byte, v float32) {
	u := math.Float32bits(v)
	b[0] = byte(u)
	b[1] = byte(u >> 8)
	b[2] = byte(u >> 16)
	b[3] = byte(u >> 24)
}

func TestWrapCaptureBindsPermissionNotFile(t *testing.T) {
	denied := wrapCapture(errors.New("miniaudio: Permission denied"))
	if fault.Classify(denied.Error()) != fault.Permission {
		t.Fatalf("denied → %s (%v)", fault.Classify(denied.Error()), denied)
	}
	if strings.Contains(strings.ToLower(denied.Error()), "wav") {
		t.Fatalf("permission error mentions wav: %v", denied)
	}
	other := wrapCapture(errors.New("miniaudio: Failed to open backend device"))
	if fault.Classify(other.Error()) != fault.CaptureUnavailable {
		t.Fatalf("open → %s (%v)", fault.Classify(other.Error()), other)
	}
	if fault.Classify(microphoneDenied().Error()) != fault.Permission {
		t.Fatal(microphoneDenied())
	}
}

func TestMicrophonePlistDeclaresUsage(t *testing.T) {
	// Linked into the Darwin binary as __TEXT,__info_plist. Keep the
	// key here so a refactor cannot drop the TCC usage string.
	body := infoPlist
	for _, key := range []string{
		"NSMicrophoneUsageDescription",
		"com.miaoboai.hearline",
		"Hearline",
	} {
		if !strings.Contains(body, key) {
			t.Fatalf("Info.plist missing %s", key)
		}
	}
}
