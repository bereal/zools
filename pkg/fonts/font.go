package fonts

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"gopkg.in/yaml.v3"
)

type CharLine struct {
	width int
	data  byte
}

func ParseCharLine(s string) (int, byte) {
	var data byte
	width := len(strings.TrimRight(s, " "))

	for _, c := range fmt.Sprintf("%8s", s) {
		var i byte
		if c != ' ' {
			i = 1
		}
		data = (data << 1) + i
	}
	data <<= (8 - width)
	return width, data
}

type Character struct {
	width int
	lines []byte
}

func ParseCharacter(rawData string) Character {
	var parsedLines []byte
	var width int
	rawLines := strings.Split(rawData, "\n")
	if len(rawLines[0]) == 0 {
		rawLines = rawLines[1:]
	}
	for _, line := range rawLines {
		line_width, data := ParseCharLine(line)
		if line_width > width {
			width = line_width
		}
		parsedLines = append(parsedLines, data)
	}

	return Character{width, parsedLines}
}

func (c Character) Encode() []byte {
	result := make([]byte, 8)
	for i, b := range c.lines {
		result[i] = b
	}
	result[7] = byte(c.width)
	return result
}

type Font struct {
	chars map[rune]Character
}

func NewFont() *Font {
	return &Font{make(map[rune]Character)}
}

func (f *Font) AddChar(key string, char Character) {
	var r rune

	for _, c := range key {
		r = c
		break
	}
	f.chars[r] = char
}

func (f *Font) ParseChars(data map[string]string) {
	for k, v := range data {
		char := ParseCharacter(v)
		f.AddChar(k, char)
	}
}

func (f *Font) Encode() []byte {
	result := make([]byte, 8)
	result[7] = 2

	var gap = 0
	var i byte
	for i = 33; i < 255; i++ {
		if i >= 128 && i < 128+32 {
			continue
		}

		key := charmap.KOI8R.DecodeByte(i)
		if char, ok := f.chars[key]; ok {
			if gap > 0 {
				result = append(result, make([]byte, 8*gap)...)
			}
			gap = 0
			result = append(result, char.Encode()...)
		} else {
			gap += 1
		}
	}
	return result
}

func (f *Font) Write(w io.Writer) error {
	_, err := w.Write(f.Encode())
	return err
}

func (f *Font) ReadYaml(r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	m := make(map[string]string)
	if err = yaml.Unmarshal(data, &m); err != nil {
		return nil
	}

	f.ParseChars(m)
	return nil
}
