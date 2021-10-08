package emu_test

import (
	"embed"
	"image/png"
	"testing"

	"github.com/bereal/zools/pkg/emu"
	"github.com/koron-go/z80"
	"github.com/stretchr/testify/assert"
)

//go:embed test_data/*
var testData embed.FS

func TestDumpScreen(t *testing.T) {
	scr, err := testData.ReadFile("test_data/zx128.scr")
	if !assert.NoError(t, err) {
		return
	}

	memory := make([]byte, 0x10000)
	for i := 0; i < len(scr); i++ {
		memory[0x4000+i] = scr[i]
	}

	img := emu.DumpScreen(z80.DumbMemory(memory), 0x4000)

	pngFile, err := testData.Open("test_data/zx128.png")
	if !assert.NoError(t, err) {
		return
	}
	pngImg, err := png.Decode(pngFile)
	if !assert.NoError(t, err) {
		return
	}

	assert.True(t, emu.ImagesEqual(img, pngImg))
}
