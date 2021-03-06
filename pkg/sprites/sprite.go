package sprites

import (
	"fmt"
	"image"
	"image/png"
	"io"
)

type Sprite struct {
	img      image.Image
	inverted bool
	flippedV bool
}

func ReadSprite(r io.Reader) (*Sprite, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	return &Sprite{img, false, false}, nil
}

func (s Sprite) Invert() Sprite {
	return Sprite{s.img, !s.inverted, s.flippedV}
}

func (s Sprite) FlipV() Sprite {
	return Sprite{s.img, s.inverted, !s.flippedV}
}

func (s Sprite) EncodeByColumns(masked bool) []byte {
	size := s.img.Bounds().Size()

	encodeColumn := func(col int) []byte {
		var result []byte
		for row := 0; row < size.Y; row++ {
			mask, sprite := s.encodeChunk(col, row)
			if masked {
				result = append(result, mask)
			}
			result = append(result, sprite)
		}

		if s.flippedV {
			reverseBytes(result)
		}
		return result
	}

	var encoded []byte
	width := (size.X / 8) * 8
	for col := 0; col < width; col += 8 {
		encoded = append(encoded, encodeColumn(col)...)
	}

	return encoded
}

func (s Sprite) EncodeByRows(masked bool) []byte {
	rows := s.encodeRows(masked)
	var encoded []byte
	for _, row := range rows {
		encoded = append(encoded, row...)
	}

	return encoded
}

func (s Sprite) EncodeZigZag(masked bool) []byte {
	var encoded []byte
	rows := s.encodeRows(masked)
	chunkSize := 1
	if masked {
		chunkSize = 2
	}
	for i, row := range rows {
		if i&1 == 1 {
			reverseChunks(row, chunkSize)
		}
		encoded = append(encoded, row...)
	}

	return encoded
}

func (s Sprite) encodeRows(masked bool) [][]byte {
	size := s.img.Bounds().Size()

	width := (size.X / 8) * 8
	encodeRow := func(row int) []byte {
		var result []byte
		for col := 0; col < width; col += 8 {
			mask, sprite := s.encodeChunk(col, row)
			if masked {
				result = append(result, mask)
			}
			result = append(result, sprite)
		}
		return result
	}

	var rows [][]byte
	for row := 0; row < size.Y; row++ {
		rows = append(rows, encodeRow(row))
	}

	if s.flippedV {
		reverseSlices(rows)
	}
	return rows
}

func (s Sprite) encodeChunk(col, row int) (byte, byte) {
	var mask, sprite byte
	for i := col; i < col+8; i++ {
		m, s := s.pixelAt(i, row)
		mask = (mask << 1) + m
		sprite = (sprite << 1) + s
	}
	return mask, sprite
}

func (s Sprite) heightInPx() int {
	b := s.img.Bounds()
	return b.Max.Y - b.Min.Y
}

func (s Sprite) widthInPx() int {
	b := s.img.Bounds()
	return (b.Max.X - b.Min.X)
}

func (s Sprite) pixelAt(x, y int) (byte, byte) {
	min := s.img.Bounds().Min
	color := s.img.At(min.X+x, min.Y+y)
	r, g, b, a := color.RGBA()
	var mask, sprite byte
	if r == 0 && g == 0 && b == 0 {
		sprite = 1
	}

	if s.inverted {
		sprite ^= 1
	}

	if a == 0 {
		mask = 1
		sprite = 0
	}
	return mask, sprite
}

func ReadSpriteSheet(r io.Reader, w, h int) ([]Sprite, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	sub, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})

	if !ok {
		return nil, fmt.Errorf("format %s doesn't support sub images", format)
	}

	var sprites []Sprite
	size := img.Bounds()
	maxX := (size.Max.X / w) * w
	maxY := (size.Max.Y / h) * h

	for y := 0; y < maxY; y += h {
		for x := 0; x < maxX; x += w {
			r := image.Rectangle{
				image.Point{x, y},
				image.Point{x + w, y + h},
			}
			si := sub.SubImage(r)
			sprites = append(sprites, Sprite{si, false, false})
		}
	}
	return sprites, nil
}

func reverseBytes(d []byte) {
	j := len(d) - 1
	for i := 0; i < j; i++ {
		d[i], d[j] = d[j], d[i]
		j--
	}
}

func reverseSlices(d [][]byte) {
	j := len(d) - 1
	for i := 0; i < j; i++ {
		d[i], d[j] = d[j], d[i]
		j--
	}
}

func reverseChunks(d []byte, chunkSize int) {
	j := len(d) - chunkSize
	for i := 0; i < j; i += chunkSize {
		for k := 0; k < chunkSize; k++ {
			d[i+k], d[j+k] = d[j+k], d[i+k]
		}
		j -= chunkSize
	}
}
