package z80_test

import (
	"fmt"
	"image/png"
	"os"
	"testing"

	"github.com/bereal/zools/pkg/emu"
	"github.com/koron-go/z80"
	"github.com/stretchr/testify/assert"
)

func saveScreen(p string, cpu *z80.CPU) {
	img := emu.DumpScreen(cpu.Memory, 0x4000)
	f, _ := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	err := png.Encode(f, img)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

func TestPutSprite(t *testing.T) {
	code := `
		ORG 0x8000
		LD HL, sprites
		LD DE, 0x4600
		LD B, 1
		LD C, 4

		CALL put_sprite_width_16
		HALT

		INCLUDE "sprite.z80"

sprites:
		INCBIN "test_data/astronaut.bin"
	`
	speccy, err := run(code, 0x8000)
	assert.NoError(t, err)
	speccy.DumpState()

	saveScreen("test.png", speccy.CPU)
}
