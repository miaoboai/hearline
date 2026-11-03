package ui

import (
	"runtime"

	g "github.com/egoist/mygo/ui"
)

// Frozen Hearline tokens (design system v0.2). Screens read these
// constants; docs/hearline-design-system.md is not a second source.
const (
	HexGround         = "#F3F4F6"
	HexCard           = "#FFFFFF"
	HexInk            = "#15171C"
	HexMuted          = "#5A5F6B"
	HexQuiet          = "#667085"
	HexLine           = "#E8EAEE"
	HexAccent         = "#C2410C"
	HexAccentHover    = "#A83410"
	HexAccentPressed  = "#8F2C0D"
	HexAccentDisabled = "#F0B3A3"
	HexStage          = "#15171C"
	HexOnDark         = "#F7F8FA"
	HexHist           = "#A8ADB8"
	HexPartial        = "#C8CDD6"
	HexDanger         = "#D92D20"
	HexOnAccent       = "#FFFFFF"
	HexFocus          = "#2F6FED"

	Space1        float32 = 8
	Space2        float32 = 16
	Space3        float32 = 24
	Space4        float32 = 32
	Space5        float32 = 40
	SpaceHalf     float32 = 4
	PadWindow     float32 = 16
	PadHeaderX    float32 = 16
	HeaderH       float32 = 56
	PadStageX     float32 = 16
	PadStageY     float32 = 12
	GapBlock      float32 = 16
	GapInline     float32 = 8
	RowSettings   float32 = 44
	RadiusCtrl    float32 = 8
	RadiusStage   float32 = 12
	RadiusOverlay float32 = 12
	RadiusMenu    float32 = 8
	PrimaryMinH   float32 = 36
	PrimaryPadX   float32 = 20
	ListenDot     float32 = 8

	FontMac     = "SF Pro Text"
	FontWindows = "Segoe UI Variable"
	FontLinux   = "Inter"
	FontMonoMac = "ui-monospace"
)

// TypeFace is one row of the type scale: size, weight, line height in px.
type TypeFace struct {
	Size   float32
	Weight int
	Line   float32
}

var (
	TypeBrand       = TypeFace{20, 600, 28}
	TypeSub         = TypeFace{12, 400, 16}
	TypeCurrent     = TypeFace{28, 600, 36}
	TypeCurrentLong = TypeFace{24, 600, 32}
	TypeCurrentMeta = TypeFace{12, 500, 16}
	TypeHist        = TypeFace{15, 400, 22}
	TypePartial     = TypeFace{26, 400, 34}
	TypePartialLong = TypeFace{24, 400, 32}
	TypeEmpty       = TypeFace{13, 400, 18}
	TypeTextLink    = TypeFace{13, 500, 18}
	TypePrimaryBtn  = TypeFace{15, 600, 20}
	TypeBody        = TypeFace{14, 400, 20}
	TypeCaption     = TypeFace{12, 400, 16}
	TypeAbout       = TypeFace{13, 400, 20}
)

var (
	colorGround         = g.Hex(HexGround)
	colorCard           = g.Hex(HexCard)
	colorInk            = g.Hex(HexInk)
	colorMuted          = g.Hex(HexMuted)
	colorQuiet          = g.Hex(HexQuiet)
	colorLine           = g.Hex(HexLine)
	colorAccent         = g.Hex(HexAccent)
	colorAccentHover    = g.Hex(HexAccentHover)
	colorAccentPressed  = g.Hex(HexAccentPressed)
	colorAccentDisabled = g.Hex(HexAccentDisabled)
	colorStage          = g.Hex(HexStage)
	colorOnDark         = g.Hex(HexOnDark)
	colorHist           = g.Hex(HexHist)
	colorPartial        = g.Hex(HexPartial)
	colorDanger         = g.Hex(HexDanger)
	colorOnAccent       = g.Hex(HexOnAccent)
	colorFocus          = g.Hex(HexFocus)
	colorOnAccentDim    = g.RGBA(255, 255, 255, 0.80)
	shadowOverlay       = g.RGBA(0, 0, 0, 0.08)
	shadowCard          = g.RGBA(0, 0, 0, 0.04)
	maskScrim           = g.RGBA(0, 0, 0, 0.40)
)

func fontFamily() string {
	switch runtime.GOOS {
	case "darwin":
		return FontMac
	case "windows":
		return FontWindows
	default:
		return FontLinux
	}
}

func fontMono() string {
	switch runtime.GOOS {
	case "darwin":
		return FontMonoMac
	case "windows":
		return "Cascadia Mono"
	default:
		return "DejaVu Sans Mono"
	}
}

func text(c *g.Context, s string, face TypeFace, col g.Color) g.Element {
	return g.Text(c, s).Font(fontFamily()).FontSize(face.Size).FontWeight(face.Weight).FixedLineHeight(face.Line).TextColor(col)
}
