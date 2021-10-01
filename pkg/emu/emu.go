package emu

import (
	"context"
	"fmt"
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
		println(string(err.(*exec.ExitError).Stderr))
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
