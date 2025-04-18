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

	"github.com/bereal/zools/pkg/asm"
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

func (t Tile) EncodeAsm(bg color.ZXAttr, b *asm.Builder) {
	bin := t.EncodeBinary(bg)

	b.Label(t.Name)
	cell := 0
	var visibleCells []bool
	for i := 0; i < len(bin); i += 9 {
		attr := color.ZXAttr(bin[i])
		b.DEFB("", []byte{byte(attr)})
		visible := color.ZXAttr(attr).Invert() != attr
		visibleCells = append(visibleCells, visible)
		if visible {
			b.Ref("", fmt.Sprintf(".cell_%d", cell))
		} else {
			b.DEFW("", []int{0})
		}
		cell++
	}

	cell = 0
	for i := 0; i < len(bin); i += 9 {
		if visibleCells[cell] {
			b.Label(fmt.Sprintf(".cell_%d", cell))
			b.DEFB("", bin[i+1:i+9])
		}
		cell++
	}
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
