package ui

import (
	"math"
	"runtime"
	"testing"

	g "github.com/egoist/mygo/ui"
)

func TestTokensAreCode(t *testing.T) {
	if HexAccent != "#E8552B" {
		t.Fatalf("accent token %s", HexAccent)
	}
	if colorAccent != g.Hex(HexAccent) {
		t.Fatal("colorAccent is not HexAccent")
	}
	if TypeCurrent.Size != 28 || TypeCurrent.Weight != 600 || TypeCurrent.Line != 36 {
		t.Fatalf("type.current %+v", TypeCurrent)
	}
	if TypeCurrentLong.Size != 24 || TypeCurrentLong.Line != 32 {
		t.Fatalf("type.current long %+v", TypeCurrentLong)
	}
	if RadiusCtrl != 8 || PrimaryMinH != 36 || PrimaryPadX != 20 {
		t.Fatalf("control geometry r=%v h=%v pad=%v", RadiusCtrl, PrimaryMinH, PrimaryPadX)
	}
	if Space1 != 8 || Space2 != 16 || PadWindow != 16 || PadStageX != 16 || PadStageY != 12 {
		t.Fatal("spacing tokens drifted")
	}
	if HexQuiet != "#8B909A" || HexLine != "#E8EAEE" || HexFocus != "#2F6FED" {
		t.Fatal("color tokens drifted")
	}
	theme := hearlineTheme()
	if theme.Font != fontFamily() {
		t.Fatalf("theme font %q", theme.Font)
	}
	if runtime.GOOS == "linux" && theme.Font != "Inter" {
		t.Fatalf("linux font %q", theme.Font)
	}
	if theme.Focus != colorFocus || theme.Radius != RadiusCtrl {
		t.Fatal("theme does not use the focus or radius tokens")
	}
}

func TestContrastFloors(t *testing.T) {
	cases := []struct {
		name string
		fg   string
		bg   string
		min  float64
	}{
		{"ink/card", HexInk, HexCard, 12},
		{"ink/ground", HexInk, HexGround, 12},
		{"muted/card", HexMuted, HexCard, 4.5},
		{"onDark/stage", HexOnDark, HexStage, 12},
		{"hist/stage", HexHist, HexStage, 4.5},
		{"partial/stage", HexPartial, HexStage, 3},
		{"quiet/stage", HexQuiet, HexStage, 4.5},
	}
	for _, tc := range cases {
		got := contrastHex(tc.fg, tc.bg)
		if got < tc.min {
			t.Fatalf("%s contrast %.2f < %.1f", tc.name, got, tc.min)
		}
	}
}

func contrastHex(fg, bg string) float64 {
	return contrast(parseHex(fg), parseHex(bg))
}

func parseHex(h string) [3]float64 {
	c := g.Hex(h)
	return [3]float64{float64(c.R), float64(c.G), float64(c.B)}
}

func contrast(fg, bg [3]float64) float64 {
	l1 := relLuma(fg)
	l2 := relLuma(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relLuma(rgb [3]float64) float64 {
	ch := func(v float64) float64 {
		v = v / 255
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(rgb[0]) + 0.7152*ch(rgb[1]) + 0.0722*ch(rgb[2])
}
