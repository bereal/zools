package color

import (
	"fmt"
	"image/color"
	"log"
)

type ZXAttr uint8

func (c ZXAttr) AsInk() ZXAttr {
	return c & 0b01000111
}

func (c ZXAttr) Invert() ZXAttr {
	ink := c & 0b111
	paper := c & 0b111000
	return (c & 0b01000000) | (paper >> 3) | (ink << 3)
}

func parseRGB(s string) (res color.NRGBA) {
	_, err := fmt.Sscanf(s, "%02x%02x%02x", &res.R, &res.G, &res.B)
	res.A = 0xff
	if err != nil {
		log.Fatal(err, s)
	}
	return
}

func RGBtoZX(col color.Color) (res ZXAttr) {
	// rgba := color.NRGBAModel.Convert(col).(color.NRGBA)
	defer func() {
		fmt.Printf("Converted %T %v to %v\n", col, col, res)
	}()

	idx := rgbPalette.Index(col)
	fmt.Printf("Index: %d\n", idx)
	if idx < 8 {
		return ZXAttr(idx)
	}

	return ZXAttr(0x40 | (idx - 7))
}

var rgbPaletteStr = []string{
	"000000",
	"0100CE",
	"CF0100",
	"CF01CE",
	"00CF15",
	"01CFCF",
	"CFCF15",
	"CFCFCF",

	"0200FD",
	"FF0201",
	"FF02FD",
	"00FF1C",
	"02FFFF",
	"FFFF1D",
	"FFFFFF",
}

var rgbPalette color.Palette

func init() {
	for _, s := range rgbPaletteStr {
		rgbPalette = append(rgbPalette, parseRGB(s))
	}
	fmt.Printf("Initializing color palette %v\n", rgbPalette)
}
