package emu

import (
	"image"
	"image/color"

	"github.com/koron-go/z80"
)

func ScreenAttrToColors(attrs uint8) (paper color.RGBA, ink color.RGBA) {
	var colValue uint8
	if attrs&0x40 == 0 {
		colValue = 0xc0
	} else {
		colValue = 0xff
	}

	v := func(bit uint8) uint8 {
		if (1<<bit)&attrs == 0 {
			return 0
		}
		return colValue
	}

	paper = color.RGBA{G: v(5), R: v(4), B: v(3), A: 0xff}
	ink = color.RGBA{G: v(2), R: v(1), B: v(0), A: 0xff}
	return
}

type Pallette [][2]color.RGBA

func (p Pallette) GetCellColor(col, row uint8) (paper color.RGBA, ink color.RGBA) {
	idx := int(row)<<5 + int(col)
	elem := p[idx]
	paper = elem[0]
	ink = elem[1]
	return
}

func (p Pallette) GetPixelColor(x, y uint8) (paper color.RGBA, ink color.RGBA) {
	return p.GetCellColor(x>>3, y>>3)
}

func CreatePallette(mem z80.Memory, addr uint16) Pallette {
	p := make([][2]color.RGBA, 0)
	var i uint16
	for i = 0; i < 0x500; i++ {
		paper, ink := ScreenAttrToColors(mem.Get(addr + i))
		p = append(p, [2]color.RGBA{paper, ink})
	}
	return p
}

func DumpScreen(mem z80.Memory, addr uint16) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 255, 191))
	palette := CreatePallette(mem, addr+0x1800)
	var i uint16
	for i = 0; i < 0x1800; i++ {
		x := uint8((i & 0x1f) << 3)
		y := uint8((i>>5)&0xc0 | (i>>2)&0x38 | (i>>8)&7)
		v := mem.Get(addr + i)

		paper, ink := palette.GetPixelColor(x, y)

		var mask uint8
		for mask = 0x80; mask != 0; mask >>= 1 {
			var col color.RGBA
			if mask&v == 0 {
				col = paper
			} else {
				col = ink
			}
			img.SetRGBA(int(x), int(y), col)
			x++
		}
	}
	return img
}

func ImagesEqual(i1, i2 image.Image) bool {
	b1 := i1.Bounds()
	b2 := i2.Bounds()

	if !(b1.Dx() == b2.Dx() && b1.Dy() == b2.Dy()) {
		return false
	}

	min1 := b1.Min
	min2 := b2.Min

	for x := 0; x <= b1.Dx(); x++ {
		for y := 0; y <= b1.Dy(); y++ {
			c1 := i1.At(min1.X+x, min1.Y+y)
			c2 := i2.At(min2.X+x, min2.Y+y)
			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()
			if !(r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2) {
				return false
			}
		}
	}
	return true
}
