package asm

import (
	"fmt"
	"io"
	"log"
	"slices"
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
	if label != "" {
		m.lines = append(m.lines, label)
	}
	return m
}

func (m *Builder) Str(label, s string) *Builder {
	koi8, err := charmap.KOI8R.NewEncoder().Bytes([]byte(s))
	if err != nil {
		log.Fatal(err.Error())
	}

	processed := make([]byte, 0, len(koi8))
	var esc bool
	for _, c := range koi8 {
		if esc {
			switch c {
			case '<':
				processed = append(processed, 1)
			case '\\':
				processed = append(processed, '\\')
			}
			esc = false
			continue
		}

		if c == '\\' {
			esc = true
			continue
		}

		processed = append(processed, c)
	}
	return m.DEFB(label, append(processed, 0))
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
	chunks := slices.Chunk(data, 16)
	println("DEFB")
	for chunk := range chunks {
		elems := make([]string, 0, len(chunk))
		for _, b := range chunk {
			elems = append(elems, fmt.Sprintf("0x%02x", b))
		}
		fmt.Printf("chunk: %v\n", elems)
		m.Linef(label, "db %s", strings.Join(elems, ", "))
		label = ""
	}
	return m
}

func (m *Builder) DEFW(label string, data []int) *Builder {
	elems := make([]string, 0, len(data))
	for _, b := range data {
		elems = append(elems, fmt.Sprintf("0x%02x", b))
	}
	return m.Linef(label, "dw %s", strings.Join(elems, ", "))
}

func (m *Builder) DEFG(label string, data []int) *Builder {
	encodeGraphicsByte := func(b byte) string {
		var res string
		var mask byte
		for mask = 0x80; mask != 0; mask >>= 1 {
			if b&mask != 0 {
				res += "#"
			} else {
				res += "."
			}
		}
		return res
	}

	m.Label(label)
	for _, b := range data {
		m.Linef("", "dg %s", encodeGraphicsByte(byte(b)))
	}

	return m
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
