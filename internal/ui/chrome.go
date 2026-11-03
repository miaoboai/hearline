package ui

import g "github.com/egoist/mygo/ui"

// btnLook forces a primary state for the acceptance collage. lookAuto
// follows the pointer and the keyboard.
type btnLook int

const (
	lookAuto btnLook = iota
	lookDefault
	lookHover
	lookPressed
	lookDisabled
	lookFocus
)

func paintPrimary(c *g.Context, key, label string, disabled bool, look btnLook) bool {
	if look == lookDisabled {
		disabled = true
	}
	btn := g.ButtonBase(c.Key(key)).MinHeight(PrimaryMinH).Padding(8, PrimaryPadX).Radius(RadiusCtrl).
		Shrink(0).AlignItems(g.Center).Justify(g.Center).Cursor(g.CursorPointer).
		Background(primaryFill(look, disabled))
	if disabled {
		btn = btn.Disabled(true)
	}
	btn.FocusRing(false)
	if look == lookFocus || (look == lookAuto && btn.FocusVisible()) {
		btn.DrawOver(func(p *g.Painter, r g.Rect) {
			p.FocusRing(r, [4]float32{RadiusCtrl, RadiusCtrl, RadiusCtrl, RadiusCtrl})
		})
	}
	ink := colorOnAccent
	if disabled {
		ink = colorOnAccentDim
	}
	btn.Children(func() {
		text(c, label, TypePrimaryBtn, ink).SingleLine()
	})
	return btn.Clicked()
}

func primaryFill(look btnLook, disabled bool) g.Color {
	switch look {
	case lookHover:
		return colorAccentHover
	case lookPressed:
		return colorAccentPressed
	case lookDisabled:
		return colorAccentDisabled
	case lookFocus, lookDefault:
		return colorAccent
	}
	if disabled {
		return colorAccentDisabled
	}
	return colorAccent
}

// primaryFillLive resolves hover and pressed from the element. Call it
// after the button exists, and pass the color into Background before
// the frame paints. Hovered and Pressed report the previous pointer
// state, which is what this frame should show.
func primaryFillLive(btn g.Element, disabled bool) g.Color {
	switch {
	case disabled:
		return colorAccentDisabled
	case btn.Pressed():
		return colorAccentPressed
	case btn.Hovered():
		return colorAccentHover
	default:
		return colorAccent
	}
}

func paintPrimaryLive(c *g.Context, key, label string, disabled bool) bool {
	btn := g.ButtonBase(c.Key(key)).MinHeight(PrimaryMinH).Padding(8, PrimaryPadX).Radius(RadiusCtrl).
		Shrink(0).AlignItems(g.Center).Justify(g.Center).Cursor(g.CursorPointer)
	if disabled {
		btn = btn.Disabled(true)
	}
	btn.Background(primaryFillLive(btn, disabled))
	btn.FocusRing(false)
	if btn.FocusVisible() {
		btn.DrawOver(func(p *g.Painter, r g.Rect) {
			p.FocusRing(r, [4]float32{RadiusCtrl, RadiusCtrl, RadiusCtrl, RadiusCtrl})
		})
	}
	ink := colorOnAccent
	if disabled {
		ink = colorOnAccentDim
	}
	btn.Children(func() {
		text(c, label, TypePrimaryBtn, ink).SingleLine()
	})
	return btn.Clicked()
}

func paintLink(c *g.Context, key, label string, disabled bool) (g.Element, bool) {
	btn := g.ButtonBase(c.Key(key)).MinHeight(TypeTextLink.Line+Space1).Padding(SpaceHalf, Space1).
		Shrink(0).AlignItems(g.Center).Cursor(g.CursorPointer)
	if disabled {
		btn = btn.Disabled(true)
	}
	col := colorMuted
	switch {
	case disabled:
		col = colorQuiet
	case btn.Pressed() || btn.Hovered() || btn.FocusVisible():
		col = colorInk
	}
	btn.FocusRing(false)
	if btn.FocusVisible() {
		btn.DrawOver(func(p *g.Painter, r g.Rect) {
			p.FocusRing(r, [4]float32{RadiusCtrl, RadiusCtrl, RadiusCtrl, RadiusCtrl})
		})
	}
	btn.Children(func() {
		text(c, label, TypeTextLink, col).SingleLine()
	})
	return btn, btn.Clicked()
}

func menuRow(c *g.Context, key, label string, disabled bool) bool {
	btn := g.ButtonBase(c.Key(key)).MinHeight(PrimaryMinH).Padding(Space1, Space2).FillWidth().Radius(RadiusCtrl).AlignItems(g.Center)
	if disabled {
		btn = btn.Disabled(true)
	} else if btn.Hovered() || btn.Highlighted() {
		btn.Background(colorGround)
	}
	col := colorInk
	if disabled {
		col = colorQuiet
	}
	btn.FocusRing(false)
	btn.Children(func() {
		text(c, label, TypeBody, col).SingleLine()
	})
	return btn.Clicked()
}

func selectField(c *g.Context, key string, selected *string, options []string, focus bool) {
	if selected == nil {
		return
	}
	sel := g.SelectBase(c.Key(key), selected)
	trig := sel.Trigger.MinHeight(PrimaryMinH).Padding(0, Space2).Radius(RadiusCtrl).
		Border(1, colorLine).AlignItems(g.Center).Gap(GapInline).MinWidth(160).Cursor(g.CursorPointer)
	bg := colorCard
	if trig.Hovered() || sel.Open() {
		bg = colorGround
	}
	trig.Background(bg)
	if focus {
		trig.Focus()
	}
	trig.FocusRing(false)
	if trig.FocusVisible() {
		trig.DrawOver(func(p *g.Painter, r g.Rect) {
			p.FocusRing(r, [4]float32{RadiusCtrl, RadiusCtrl, RadiusCtrl, RadiusCtrl})
		})
	}
	label := *selected
	if label == "" {
		label = "选择"
	}
	trig.Children(func() {
		text(c, label, TypeBody, colorInk).SingleLine().Grow(1)
		chevron(c)
	})
	sel.Popup(func(panel g.Element) {
		panel.Background(colorCard).Radius(RadiusMenu).Padding(SpaceHalf).Border(1, colorLine).
			Shadow(0, 4, 12, 0, shadowOverlay).MinWidth(160).Gap(SpaceHalf)
		for _, opt := range options {
			item := sel.Item(opt).MinHeight(PrimaryMinH).Padding(Space1, Space2).Radius(RadiusCtrl).AlignItems(g.Center)
			if item.Highlighted() || opt == *selected {
				item.Background(colorGround)
			}
			opt := opt
			item.Children(func() {
				text(c, opt, TypeBody, colorInk).SingleLine()
			})
		}
	})
}

func textField(c *g.Context, key string, value *string, placeholder string, focus bool, mono bool) g.Element {
	el := g.TextInputBase(c.Key(key), value).MinHeight(PrimaryMinH).Padding(Space1, Space2).Radius(RadiusCtrl).
		Background(colorCard).Border(1, colorLine).FontSize(TypeBody.Size).FixedLineHeight(TypeBody.Line).Grow(1)
	if mono {
		el = el.Font(fontMono())
	} else {
		el = el.Font(fontFamily())
	}
	if placeholder != "" {
		el = el.Placeholder(placeholder).PlaceholderColor(colorQuiet)
	}
	if focus {
		el = el.Focus()
	}
	el.FocusRing(false)
	if el.FocusVisible() {
		el.DrawOver(func(p *g.Painter, r g.Rect) {
			p.FocusRing(r, [4]float32{RadiusCtrl, RadiusCtrl, RadiusCtrl, RadiusCtrl})
		})
	}
	return el
}

func toggle(c *g.Context, key string, on *bool) bool {
	sw := g.SwitchBase(c.Key(key), on).Width(40).Height(24).Radius(12).Padding(2).Shrink(0).AlignItems(g.Center)
	if on != nil && *on {
		sw.Background(colorAccent).Justify(g.End)
	} else {
		sw.Background(colorLine).Justify(g.Start)
	}
	sw.FocusRing(false)
	if sw.FocusVisible() {
		sw.DrawOver(func(p *g.Painter, r g.Rect) {
			p.FocusRing(r, [4]float32{12, 12, 12, 12})
		})
	}
	sw.Children(func() {
		g.Box(c).Size(20, 20).Radius(10).Background(colorCard)
	})
	return sw.Changed()
}

func chevron(c *g.Context) {
	g.Box(c).Size(16, 16).Shrink(0).Draw(func(p *g.Painter, r g.Rect) {
		var path g.Path
		path.MoveTo(r.X+r.W*0.22, r.Y+r.H*0.38).LineTo(r.X+r.W*0.5, r.Y+r.H*0.62).LineTo(r.X+r.W*0.78, r.Y+r.H*0.38)
		p.StrokePath(&path, 1.5, colorMuted)
	})
}

func hairline(c *g.Context) {
	g.Box(c).Height(1).FillWidth().Background(colorLine).Shrink(0)
}
