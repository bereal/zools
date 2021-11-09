package emu_test

import (
	"testing"

	"github.com/bereal/zools/pkg/emu"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	code := `
		ORG 0x8000
		XOR A
		INC A

		HALT
	`

	s, err := emu.CompileAndRun(code, nil, 0x8000)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, byte(1), s.CPU.AF.Hi)
}
