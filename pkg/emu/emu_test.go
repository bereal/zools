package emu_test

import (
	"os/exec"
	"testing"

	"github.com/bereal/zools/pkg/emu"
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
