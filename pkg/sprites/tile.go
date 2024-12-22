package sprites

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"

	"github.com/bereal/zools/pkg/color"
)

type Cell struct {
	img image.Image
}

func (c Cell) Encode(bg color.ZXAttr) (color.ZXAttr, []byte) {
	var ink color.ZXAttr

	var rows []byte
	for y := c.img.Bounds().Min.Y; y < c.img.Bounds().Max.Y; y++ {
		var row byte
		for x := c.img.Bounds().Min.X; x < c.img.Bounds().Max.X; x++ {
			px := c.img.At(x, y)
			var bit byte
			_, _, _, a := px.RGBA()
			if a > 0 {
				attr := color.RGBtoZX(px)
				if attr != bg {
					bit = 1
					ink = attr
				}
			}
			row = (row << 1) | bit
		}
		rows = append(rows, row)
	}
	return (ink | bg.Invert()), rows
}

type Tile struct {
	img WithSubImage
}

func (t Tile) Split(w, h int) []Tile {
	subs := splitImage(t.img, w, h)
	dedups := deduplicateImages(subs)

	res := make([]Tile, len(dedups))
	for i, img := range dedups {
		res[i] = Tile{img.(WithSubImage)}
	}
	return res
}

func (t Tile) EncodeBinary(bg color.ZXAttr) []byte {
	var res []byte
	imgs := splitImage(t.img, 8, 8)
	for _, img := range imgs {
		attr, bitmap := Cell{img}.Encode(bg)
		res = append(res, byte(attr))
		res = append(res, bitmap...)
	}
	return res
}

func (t Tile) EncodeAsm(name string, bg color.ZXAttr) []string {
	bin := t.EncodeBinary(bg)
	var lines []string
	label := strings.ReplaceAll(name, "-", "_")

	lines = append(lines, fmt.Sprintf("%s:", label))
	cell := 0
	var visibleCells []bool
	for i := 0; i < len(bin); i += 9 {
		attr := color.ZXAttr(bin[i])
		lines = append(lines, fmt.Sprintf("\tdb 0x%02x", attr))
		visible := color.ZXAttr(attr).Invert() != attr
		visibleCells = append(visibleCells, visible)
		if visible {
			lines = append(lines, fmt.Sprintf("\tdw .cell_%d", cell))
		} else {
			lines = append(lines, "\tdw 0")
		}
		cell++
	}

	cell = 0
	for i := 0; i < len(bin); i += 9 {
		if visibleCells[cell] {
			lines = append(lines, fmt.Sprintf(".cell_%d:", cell))
			for j := 1; j <= 8; j++ {
				lines = append(lines, fmt.Sprintf("\tdg %s", encodeGraphicsByte(bin[i+j])))
			}
		}
		cell++
	}
	return lines
}

func (t Tile) EncodePNG(w io.Writer) error {
	return png.Encode(w, t.img)
}

func ReadTile(r io.Reader) (*Tile, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	return &Tile{img.(WithSubImage)}, nil
}
