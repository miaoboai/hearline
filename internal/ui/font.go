package ui

import (
	_ "embed"

	g "github.com/egoist/mygo/ui"
)

// Inter is embedded for Linux, where the design system asks for it inside
// the app. macOS and Windows keep SF Pro Text and Segoe UI Variable.

//go:embed fonts/Inter-Regular.ttf
var interRegular []byte

//go:embed fonts/Inter-Medium.ttf
var interMedium []byte

//go:embed fonts/Inter-SemiBold.ttf
var interSemiBold []byte

func init() {
	for _, face := range [][]byte{interRegular, interMedium, interSemiBold} {
		_ = g.RegisterFont(face, "")
	}
}
