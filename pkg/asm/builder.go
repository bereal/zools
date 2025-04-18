package asm

import (
	"fmt"
	"io"
	"log"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

type Builder struct {
	lines []string
}

func NewBuilder() *Builder {
	return &Builder{lines: []string{}}
}

func (m *Builder) Label(label string) *Builder {
	m.lines = append(m.lines, label)
	return m
}

func (m *Builder) Str(label, s string) *Builder {
	koi8, err := charmap.KOI8R.NewEncoder().String(s)
	if err != nil {
		log.Fatal(err.Error())
	}
	m.DEFB(label, []byte(koi8))
	return m
}

func (m *Builder) Line(label, s string) *Builder {
	return m.Linef(label, "%s", s)
}

func (m *Builder) Linef(label, format string, a ...interface{}) *Builder {
	f := "%-7s " + format
	args := make([]interface{}, 0, len(a)+1)
	args = append(args, label)
	args = append(args, a...)
	m.lines = append(m.lines, fmt.Sprintf(f, args...))
	return m
}

func (m *Builder) DEFB(label string, data []byte) *Builder {
	elems := make([]string, 0, len(data))
	for _, b := range data {
		elems = append(elems, fmt.Sprintf("0x%02x", b))
	}
	return m.Linef(label, "db %s", strings.Join(elems, ", "))
}

func (m *Builder) DEFW(label string, data []int) *Builder {
	elems := make([]string, 0, len(data))
	for _, b := range data {
		elems = append(elems, fmt.Sprintf("0x%02x", b))
	}
	return m.Linef(label, "dw %s", strings.Join(elems, ", "))
}

func (m *Builder) Ref(label string, name ...string) *Builder {
	return m.Linef(label, "dw %s", strings.Join(name, ", "))
}

func (m Builder) Write(w io.Writer) error {
	for _, line := range m.lines {
		_, err := fmt.Fprintf(w, "%s\n", line)
		if err != nil {
			return err
		}
	}
	return nil
}
