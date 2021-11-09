package emu

import (
	"context"
	"fmt"
	"image"

	"github.com/bereal/zools/pkg/asm"
	"github.com/koron-go/z80"
)

type Speccy struct {
	CPU *z80.CPU
}

func NewSpeccy() *Speccy {
	memory := z80.DumbMemory(make([]byte, 0x10000))
	return &Speccy{&z80.CPU{Memory: memory}}
}

func (s *Speccy) DumpScreen(addr uint16) image.Image {
	return DumpScreen(s.CPU.Memory, addr)
}

func (s *Speccy) ClearScreen() {
	ClearScreen(s.CPU.Memory, 0x4000, GetScreenAttr(White, Black, false))
}

func (s *Speccy) Load(addr uint16, data []byte) {
	for _, b := range data {
		s.CPU.Memory.Set(addr, b)
		addr++
	}
}

func (s *Speccy) DumpState() {
	fmt.Printf(`AF=%04x
BC=%04x
DE=%04x
HL=%04x
`, s.CPU.AF.U16(), s.CPU.BC.U16(), s.CPU.DE.U16(), s.CPU.HL.U16())
}

func (s *Speccy) Run(addr uint16) error {
	s.CPU.PC = addr
	return s.CPU.Run(context.Background())
}

func CompileAndRun(code string, include []string, addr uint16) (*Speccy, error) {
	sj := asm.NewSjasmplus()
	if len(include) > 0 {
		sj.AddInclude(include...)
	}
	d, err := sj.Compile(code)
	if err != nil {
		return nil, err
	}

	println(len(d))

	s := NewSpeccy()
	s.ClearScreen()

	s.Load(addr, d)
	if err = s.Run(addr); err != nil {
		return nil, err
	}

	return s, nil
}
