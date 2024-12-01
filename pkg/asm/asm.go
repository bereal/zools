package asm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
)

type Sjasmplus struct {
	code           []string
	include        []string
	includeContent map[string][]byte
}

func NewSjasmplus() *Sjasmplus {
	return &Sjasmplus{nil, nil, make(map[string][]byte)}
}

func (s *Sjasmplus) AddInclude(i ...string) {
	s.include = append(s.include, i...)
}

func (s *Sjasmplus) AddIncludeContent(name string, content []byte) {
	s.includeContent[name] = content
}

func (s *Sjasmplus) ReadIncludeContent(name string, r io.Reader) error {
	d, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.AddIncludeContent(name, d)
	return nil
}

func (s *Sjasmplus) Compile(code string) ([]byte, error) {
	asmpath, err := exec.LookPath("sjasmplus")
	if err != nil {
		return nil, err
	}

	tempdir, err := os.MkdirTemp(os.TempDir(), "code_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempdir)

	createFile := func(name string, data []byte) error {
		p := path.Join(tempdir, name)
		return os.WriteFile(p, data, 0600)

	}

	var codeFileName string
	for i := 0; ; i++ {
		codeFileName = fmt.Sprintf("code_%d.z80", i)
		if _, ok := s.includeContent[codeFileName]; !ok {
			createFile(codeFileName, []byte(code))
			break
		}
	}

	for k, v := range s.includeContent {
		createFile(k, v)
	}

	args := []string{"sjasmplus", codeFileName, "--raw=-"}
	for _, i := range s.include {
		args = append(args, fmt.Sprintf("-I%s", i))
	}

	errbuf := &bytes.Buffer{}
	cmd := exec.Cmd{
		Path:   asmpath,
		Args:   args,
		Dir:    tempdir,
		Stderr: errbuf,
	}
	result, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), string(errbuf.Bytes()))
	}
	return result, nil
}
