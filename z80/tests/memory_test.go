package ngn

import (
	"path"
	"runtime"
	"testing"

	"github.com/bereal/zools/pkg/emu"
	"github.com/koron-go/z80"
	"github.com/stretchr/testify/assert"
)

func resolve(s string) string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Clean(path.Join(filename, s))
}

var srcdir = resolve("../..")

func run(code string, addr uint16) (*z80.CPU, error) {
	return emu.Run(code, []string{srcdir}, addr)
}

func TestMemoryBallocQueueInit(t *testing.T) {
	code := `
		ORG 0x8000
		LD HL, 0xA000
		LD BC, 5
		LD DE, 0x10
		CALL balloc_init
		HALT

		INCLUDE "memory.z80"
	`

	cpu, err := run(code, 0x8000)
	if !assert.NoError(t, err) {
		return
	}

	var i uint16
	for i = 0; i < 5; i++ {
		nextAddr := 0xA000 + (i+1)*0x12
		assert.EqualValues(t, (nextAddr%0x100)+1, cpu.Memory.Get(0xA000+i*0x12))
		assert.EqualValues(t, nextAddr/0x100, cpu.Memory.Get(0xA000+i*0x12+1))
	}
}

func TestMemoryQueueAlloc(t *testing.T) {
	code := `
		ORG 0x8000

		LD HL, 0xA000
		LD BC, 5
		LD DE, 0x10
		CALL balloc_init

		CALL balloc
		PUSH HL
		POP DE

		LD HL, 0xA000
		CALL balloc

		HALT

		INCLUDE "memory.z80"
	`

	cpu, err := run(code, 0x8000)
	if !assert.NoError(t, err) {
		return
	}
	assert.EqualValues(t, 0xA002, cpu.DE.U16())
	// assert.EqualValues(t, 0xA014, cpu.HL.U16())
}

func TestMemoryQueueFree(t *testing.T) {
	code := `
		ORG 0x8000

		LD HL, 0xA000
		LD BC, 5
		LD DE, 0x10
		CALL balloc_init

		CALL balloc       ; block 1

		LD HL, 0xA000
		CALL balloc       ; block 2
		PUSH HL           ; Save the allocated address for future
		POP DE

		LD HL, 0xA000
		CALL balloc       ; block 3

		PUSH DE
		POP HL
		CALL balloc_free  ; free block 2

		LD HL, 0xA000
		CALL balloc       ; re-alloc block 2

		HALT

		INCLUDE "memory.z80"
	`

	cpu, err := run(code, 0x8000)
	if !assert.NoError(t, err) {
		return
	}
	assert.EqualValues(t, uint16(0xA014), cpu.DE.U16())
	assert.EqualValues(t, uint16(0xA014), cpu.HL.U16())
}

func TestMemoryQueueAppend(t *testing.T) {

}
