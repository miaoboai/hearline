package ui

import g "github.com/egoist/mygo/ui"

func hearlineTheme() *g.Theme {
	t := g.LightTheme()
	t.Background = colorGround
	t.Surface = colorCard
	t.SurfaceHover = colorGround
	t.SurfacePressed = colorLine
	t.Border = colorLine
	t.Text = colorInk
	t.TextMuted = colorMuted
	t.Accent = colorAccent
	t.AccentHover = colorAccentHover
	t.AccentPressed = colorAccentPressed
	t.AccentText = colorOnAccent
	t.Danger = colorDanger
	t.Focus = colorFocus
	t.Selection = g.RGBA(47, 111, 237, 0.20)
	t.Scrollbar = g.RGBA(21, 23, 28, 0.28)
	t.Radius = RadiusCtrl
	t.Spacing = Space1
	t.FontSize = TypeBody.Size
	t.Font = fontFamily()
	return t
}
