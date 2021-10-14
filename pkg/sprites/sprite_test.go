package sprites_test

import (
	"embed"
	"testing"

	"github.com/bereal/zools/pkg/sprites"
	"github.com/stretchr/testify/assert"
)

//go:embed test_data/*.png
var content embed.FS

var sprite1 = []byte{
	0,
	0b01111110,
	0b01000010,
	0b01011010,
	0b01011010,
	0b01000010,
	0b01111110,
	0,
}

var sprite2 = []byte{
	0,
	0b01111110,
	0b01000010,
	0b01111110,
	0b01111110,
	0b01000010,
	0b01111110,
	0,
}

var sprite3 = []byte{
	0,
	0b01111110,
	0b01011010,
	0b01011010,
	0b01011010,
	0b01011010,
	0b01111110,
	0,
}

var sprite4 = []byte{
	0,
	0b01111110,
	0b01011010,
	0b01111110,
	0b01111110,
	0b01011010,
	0b01111110,
	0,
}

var mask = []byte{0xff, 0x81, 0x81, 0x81, 0x81, 0x81, 0x81, 0xff}

func interleave(a, b []byte) []byte {
	var c []byte
	for i, v := range a {
		c = append(c, v, b[i])
	}
	return c
}

func masked(sprite []byte) []byte { return interleave(mask, sprite) }

func interleaveMasked(a, b []byte) []byte {
	var c []byte
	for i := 0; i < len(a); i += 2 {
		c = append(c, a[i], a[i+1], b[i], b[i+1])
	}
	return c
}

func concat(s ...[]byte) []byte {
	v := []byte{}
	for _, b := range s {
		v = append(v, b...)
	}
	return v
}

func invert(b []byte) []byte {
	inv := make([]byte, len(b))
	for i, v := range b {
		inv[i] = ^v & ^mask[i]
	}
	return inv
}

func reversed(b []byte) []byte {
	res := make([]byte, 0, len(b))
	for i := len(b) - 1; i >= 0; i-- {
		res = append(res, b[i])
	}
	return res
}

func TestSingleCellSpriteEncoding(t *testing.T) {
	f, err := content.Open("test_data/test_sprite.png")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	sprite, err := sprites.ReadSprite(f)
	if !assert.NoError(t, err) {
		return
	}

	result := sprite.EncodeByColumns(false)
	assert.EqualValues(t, sprite1, result)

	result = sprite.EncodeByColumns(true)
	assert.EqualValues(t, masked(sprite1), result)

	result = sprite.Invert().EncodeByColumns(false)
	assert.EqualValues(t, invert(sprite1), result)

	result = sprite.Invert().EncodeByColumns(true)
	assert.EqualValues(t, masked(invert(sprite1)), result)
}

func readSpriteSheet(w, h int) ([]sprites.Sprite, error) {
	f, err := content.Open("test_data/test_spritesheet.png")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sprites, err := sprites.ReadSpriteSheet(f, w, h)
	if err != nil {
		return nil, err
	}
	return sprites, nil

}

func TestSpriteSheet8x16ByCol(t *testing.T) {
	sprites, err := readSpriteSheet(8, 16)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, sprites, 2)

	assert.EqualValues(t, concat(sprite1, sprite3), sprites[0].EncodeByColumns(false))
	assert.EqualValues(t, concat(sprite2, sprite4), sprites[1].EncodeByColumns(false))
	assert.EqualValues(t, concat(masked(sprite1), masked(sprite3)), sprites[0].EncodeByColumns(true))
	assert.EqualValues(t, concat(masked(sprite2), masked(sprite4)), sprites[1].EncodeByColumns(true))

	assert.EqualValues(t, concat(invert(sprite1), invert(sprite3)), sprites[0].Invert().EncodeByColumns(false))
}

func TestSpriteSheet16x8ByCol(t *testing.T) {
	sprites, err := readSpriteSheet(16, 8)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, sprites, 2)
	assert.EqualValues(t, concat(sprite1, sprite2), sprites[0].EncodeByColumns(false))
	assert.EqualValues(t, concat(sprite3, sprite4), sprites[1].EncodeByColumns(false))
}

func TestSpriteSheet8x16UpsideDownByCol(t *testing.T) {
	sprites, err := readSpriteSheet(8, 16)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, sprites, 2)
	assert.EqualValues(t, reversed(concat(sprite1, sprite3)), sprites[0].FlipV().EncodeByColumns(false))
	assert.EqualValues(t, reversed(concat(sprite2, sprite4)), sprites[1].FlipV().EncodeByColumns(false))
}

func TestSpriteSheet8x16ByRow(t *testing.T) {
	sprites, err := readSpriteSheet(8, 16)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, concat(sprite1, sprite3), sprites[0].EncodeByRows(false))
	assert.EqualValues(t, concat(sprite2, sprite4), sprites[1].EncodeByRows(false))
	assert.EqualValues(t, concat(masked(sprite1), masked(sprite3)), sprites[0].EncodeByRows(true))
	assert.EqualValues(t, concat(masked(sprite2), masked(sprite4)), sprites[1].EncodeByRows(true))
}

func TestSpriteSheet16x8ByRow(t *testing.T) {
	sprites, err := readSpriteSheet(16, 8)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, interleave(sprite1, sprite2), sprites[0].EncodeByRows(false))
	assert.EqualValues(t, interleave(sprite3, sprite4), sprites[1].EncodeByRows(false))
	assert.EqualValues(t, interleaveMasked(masked(sprite1), masked(sprite2)), sprites[0].EncodeByRows(true))
	assert.EqualValues(t, interleaveMasked(masked(sprite3), masked(sprite4)), sprites[1].EncodeByRows(true))

}
