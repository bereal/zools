package asm_test

import (
	"os/exec"
	"testing"

	"github.com/bereal/zools/pkg/asm"
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

	sj := asm.NewSjasmplus()
	result, err := sj.Compile(code)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestSjasmIncludeContent(t *testing.T) {
	requireSjasm(t)

	code := `
		INCLUDE "test.z80"
	`

	include := `
		XOR A
		INC A
	`

	sj := asm.NewSjasmplus()
	sj.AddIncludeContent("test.z80", []byte(include))
	result, err := sj.Compile(code)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}
