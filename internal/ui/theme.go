package ui

import g "github.com/egoist/mygo/ui"

// Colors follow the product brief: a cool grey ground, white cards, near-black
// ink, and one live accent. They are Hearline's own values, not borrowed assets.
var (
	colorGround = g.Hex("#F3F4F6")
	colorCard   = g.Hex("#FFFFFF")
	colorInk    = g.Hex("#15171C")
	colorStage  = g.Hex("#15171C")
	colorMuted  = g.Hex("#5A5F6B")
	colorQuiet  = g.Hex("#6B7080")
	colorLine   = g.Hex("#E4E6EB")
	colorAccent = g.Hex("#E8552B")
	colorHist   = g.Hex("#A8ADB8")
	colorTime   = g.Hex("#8B909A")
	colorOnDark = g.Hex("#F7F8FA")
)

func hearlineTheme() *g.Theme {
	t := g.LightTheme()
	t.Background = colorGround
	t.Surface = colorCard
	t.SurfaceHover = g.Hex("#F4F5F7")
	t.SurfacePressed = g.Hex("#E8EAEE")
	t.Border = g.Hex("#D8DBE2")
	t.Text = colorInk
	t.TextMuted = colorQuiet
	t.Accent = colorAccent
	t.AccentHover = g.Hex("#D44820")
	t.AccentPressed = g.Hex("#B83C18")
	t.AccentText = colorCard
	t.Focus = g.RGBA(232, 85, 43, 0.45)
	t.Selection = g.RGBA(232, 85, 43, 0.20)
	t.Scrollbar = g.RGBA(255, 255, 255, 0.28)
	t.Radius = 10
	t.Spacing = 5
	t.FontSize = 14
	return t
}
