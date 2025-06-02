package text

import (
	"fmt"
	"io"

	"github.com/bereal/zools/pkg/asm"
	"golang.org/x/text/encoding/charmap"
	"gopkg.in/yaml.v2"
)

type Block interface {
	Encode(label string, b *asm.Builder)
}

type Raw []byte

func (r Raw) Encode(label string, b *asm.Builder) {
	b.DEFB(label, r)
}

type IntRef string

func (s IntRef) Encode(label string, b *asm.Builder) {
	b.DEFB("", []byte{2})
	b.Ref("", string(s))
}

type Composite []Block

func (t Composite) Encode(label string, b *asm.Builder) {
	b.Label(label)
	for _, blk := range t {
		blk.Encode("", b)
	}
}

const (
	stateRaw = iota
	stateEsc
	stateRef
)

func parseBlock(text string) (Block, error) {
	var state int
	var blocks []Block
	var raw Raw
	var ref string

	encoder := charmap.KOI8R.NewEncoder()

	commitRaw := func() {
		if len(raw) > 0 {
			blocks = append(blocks, raw)
			raw = nil
		}
	}

	for _, c := range text {
		switch state {
		case stateRaw:
			if c == '\\' {
				state = stateEsc
				continue
			}

			if c == '{' {
				commitRaw()
				state = stateRef
			}

			enc, err := encoder.Bytes([]byte(string(c)))
			if err != nil {
				return nil, err
			}
			raw = append(raw, enc...)

		case stateEsc:
			switch c {
			case '<':
				raw = append(raw, 1)
			case 'n':
				raw = append(raw, '\n')
			default:
				raw = append(raw, []byte(string(c))...)
			}
			state = stateRaw

		case stateRef:
			if c == '}' {
				blocks = append(blocks, IntRef(ref))
				ref = ""
				state = stateRaw
			} else {
				ref += string(c)
			}
		}
	}

	commitRaw()
	blocks = append(blocks, Raw([]byte{0}))
	return Composite(blocks), nil
}

type I18nText map[string]Block

func (t I18nText) Encode(label string, langOrder []string, b *asm.Builder) {
	b.Label(label)

	innerLabels := make([]string, 0, len(langOrder))

	for _, lang := range langOrder {
		innerLabels = append(innerLabels, fmt.Sprintf(".%s", lang))
	}

	refLabels := make([]string, 0, len(langOrder))
	for i, lang := range langOrder {
		if _, ok := t[lang]; !ok {
			i = 0
		}
		refLabels = append(refLabels, innerLabels[i])
	}

	b.Ref("", refLabels...)

	for i, lang := range langOrder {
		if block, ok := t[lang]; ok {
			block.Encode(innerLabels[i], b)
		}
	}
}

func (t I18nText) EncodeMono(label string, lang string, b *asm.Builder) {
	b.Label(label)
	if block, ok := t[lang]; ok {
		block.Encode("", b)
	} else {
		b.DEFB("", []byte{0})
	}
}

type I18nBundle map[string]I18nText

func ReadI18nBundle(r io.Reader) (I18nBundle, error) {
	raw := map[string]map[string]string{}
	res := I18nBundle{}

	err := yaml.NewDecoder(r).Decode(&raw)
	if err != nil {
		return nil, err
	}

	for name, text := range raw {
		i18n := I18nText{}
		for lang, str := range text {
			i18n[lang], err = parseBlock(str)
			if err != nil {
				return nil, fmt.Errorf("failed to parse block %w", err)
			}
		}
		res[name] = i18n
	}
	return res, nil
}

func (b I18nBundle) Encode(langOrder []string, w io.Writer) error {
	for name, i18n := range b {
		builder := asm.NewBuilder()
		i18n.Encode(name, langOrder, builder)
		if err := builder.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (b I18nBundle) EncodeMono(lang string, w io.Writer) error {
	for name, i18n := range b {
		builder := asm.NewBuilder()
		i18n.EncodeMono(name, lang, builder)
		if err := builder.Write(w); err != nil {
			return err
		}
	}
	return nil
}
