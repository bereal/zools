package sprites

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path"
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
	img  WithSubImage
	Name string
}

func (t Tile) Split(w, h int) []*Tile {
	subs := splitImage(t.img, w, h)
	dedups := deduplicateImages(subs)

	res := make([]*Tile, len(dedups))
	for i, img := range dedups {
		res[i] = &Tile{img.(WithSubImage), fmt.Sprintf("%s-%d", t.Name, i)}
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

func (t Tile) EncodeAsm(bg color.ZXAttr) []string {
	bin := t.EncodeBinary(bg)
	var lines []string

	lines = append(lines, fmt.Sprintf("%s:", t.Name))
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

func ReadTilePNG(pngPath string) (*Tile, error) {
	name := strings.TrimSuffix(path.Base(pngPath), path.Ext(pngPath))
	r, err := os.Open(pngPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	return &Tile{img.(WithSubImage), strings.ReplaceAll(name, "-", "_")}, nil
}

type Tileset []*Tile

type xmlTile struct {
	ID    int `xml:"id,attr"`
	Image struct {
		Source string `xml:"source,attr"`
	} `xml:"image"`
}

type xmlTileset struct {
	Tiles []xmlTile `xml:"tile"`
}

func ReadTSX(tsxPath string) (Tileset, error) {
	data, err := os.ReadFile(tsxPath)
	if err != nil {
		return nil, err
	}
	var tileset xmlTileset
	err = xml.Unmarshal([]byte(data), &tileset)
	if err != nil {
		return nil, err
	}

	basedir := path.Dir(tsxPath)
	res := make(Tileset, len(tileset.Tiles))
	for i, tile := range tileset.Tiles {
		imgPath := path.Join(basedir, tile.Image.Source)
		tileImg, err := ReadTilePNG(imgPath)
		if err != nil {
			return nil, err
		}
		res[i] = tileImg
	}
	println("Tileset loaded", len(res), "tiles")
	return res, nil
}
