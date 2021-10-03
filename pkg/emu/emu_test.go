package emu_test

import (
	"os/exec"
	"testing"

	"github.com/bereal/zools/pkg/emu"
	"github.com/koron-go/z80"
	"github.com/stretchr/testify/assert"
)

func requireSjasm(t *testing.T) {
	if _, err := exec.LookPath("sjasmplus"); err != nil {
		t.Skip(err.Error())
	}
}

func TestSjasm(t *testing.T) {
	requireSjasm(t)

	code := `
		XOR A
		INC A
	`
	result, err := emu.Sjasmplus(code, nil)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRun(t *testing.T) {
	requireSjasm(t)

	code := `
		ORG 0x8000
		XOR A
		INC A

		HALT
	`

	cpu, err := emu.Run(code, nil, 0x8000)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, byte(1), cpu.AF.Hi)
}

type point struct {
	x int
	y int
}

func TestDumpImage(t *testing.T) {
	memory := z80.DumbMemory(make([]byte, 0x10000))
	memory[0x4102] = 0x55
	memory[0x4203] = 0x55

	expectedPoints := []point{
		{17, 1}, {19, 1}, {21, 1}, {23, 1},
		{25, 2}, {27, 2}, {29, 2}, {31, 2},
	}

	pointMap := make(map[point]struct{})
	for _, p := range expectedPoints {
		pointMap[p] = struct{}{}
	}

	img := emu.DumpScreen(memory, 0x4000)
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			col := img.At(x, y)
			r, g, b, a := col.RGBA()
			if _, ok := pointMap[point{x, y}]; ok {
				assert.EqualValues(t, r, 0)
				assert.EqualValues(t, g, 0)
				assert.EqualValues(t, b, 0)
			} else {
				assert.EqualValues(t, r, 0xffff)
				assert.EqualValues(t, g, 0xffff)
				assert.EqualValues(t, b, 0xffff)
			}
			assert.EqualValues(t, a, 0xffff)
		}
	}
}
