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

	result := sprite.Encode(false)
	assert.EqualValues(t, sprite1, result)

	result = sprite.Encode(true)
	assert.EqualValues(t, masked(sprite1), result)

	result = sprite.Invert().Encode(false)
	assert.EqualValues(t, invert(sprite1), result)

	result = sprite.Invert().Encode(true)
	assert.EqualValues(t, masked(invert(sprite1)), result)
}

func TestSpriteSheet8x16(t *testing.T) {
	f, err := content.Open("test_data/test_spritesheet.png")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	sprites, err := sprites.ReadSpriteSheet(f, 8, 16)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, sprites, 2)

	assert.EqualValues(t, concat(sprite1, sprite3), sprites[0].Encode(false))
	assert.EqualValues(t, concat(sprite2, sprite4), sprites[1].Encode(false))
	assert.EqualValues(t, concat(masked(sprite1), masked(sprite3)), sprites[0].Encode(true))
	assert.EqualValues(t, concat(masked(sprite2), masked(sprite4)), sprites[1].Encode(true))

	assert.EqualValues(t, concat(invert(sprite1), invert(sprite3)), sprites[0].Invert().Encode(false))
}

func TestSpriteSheet16x8(t *testing.T) {
	f, err := content.Open("test_data/test_spritesheet.png")
	if !assert.NoError(t, err) {
		return
	}
	defer f.Close()

	sprites, err := sprites.ReadSpriteSheet(f, 16, 8)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, sprites, 2)
	assert.EqualValues(t, concat(sprite1, sprite2), sprites[0].Encode(false))
	assert.EqualValues(t, concat(sprite3, sprite4), sprites[1].Encode(false))
}
