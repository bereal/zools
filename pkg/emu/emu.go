package emu

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/koron-go/z80"
)

func Sjasmplus(code string, include []string) ([]byte, error) {
	asmpath, err := exec.LookPath("sjasmplus")
	if err != nil {
		return nil, err
	}

	tempdir, err := ioutil.TempDir(os.TempDir(), "code_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempdir)

	p := path.Join(tempdir, "code.z80")
	ioutil.WriteFile(p, []byte(code), 0600)

	args := []string{"sjasmplus", "code.z80", "--raw=-"}
	for _, i := range include {
		args = append(args, fmt.Sprintf("-I%s", i))
	}

	cmd := exec.Cmd{
		Path: asmpath,
		Args: args,
		Dir:  tempdir,
	}
	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Run(code string, include []string, addr uint16) (*z80.CPU, error) {
	mem := z80.DumbMemory(make([]byte, 0x10000))
	bin, err := Sjasmplus(code, include)

	if err != nil {
		return nil, err
	}
	mem.Put(addr, bin...)
	cpu := &z80.CPU{Memory: mem}
	cpu.PC = addr
	if err = cpu.Run(context.Background()); err != nil {
		return nil, err
	}

	return cpu, nil
}

func DumpScreen(mem z80.Memory, addr uint16) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 255, 191))
	var i uint16
	for i = 0; i < 0x1800; i++ {
		x := (i & 0x1f) << 3
		y := (i>>5)&0xc0 | (i>>2)&0x38 | (i>>8)&7
		v := mem.Get(addr + i)

		var mask uint8
		for mask = 0x80; mask != 0; mask >>= 1 {
			var col color.RGBA
			col.A = 0xff
			if mask&v == 0 {
				col.R = 0xff
				col.G = 0xff
				col.B = 0xff
			}
			img.SetRGBA(int(x), int(y), col)
			x++
		}
	}
	return img
}
