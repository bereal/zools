package color

import (
	"image/color"
	"testing"
)

func TestColors(t *testing.T) {
	RGBtoZX(color.NRGBA64{0xffff, 0xffff, 0xffff, 0})
}
