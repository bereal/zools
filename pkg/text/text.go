package text

import (
	"fmt"
	"io"

	"github.com/bereal/zools/pkg/asm"
	"gopkg.in/yaml.v2"
)

type Block interface {
	Encode(label string, b *asm.Builder)
}

type RawString string

func (s RawString) Encode(label string, b *asm.Builder) {
	b.Str(label, string(s))
}

type Text []Block

func (t Text) Encode(label string, b *asm.Builder) {
	b.Label(label)
	for _, blk := range t {
		blk.Encode("", b)
	}
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
			i18n[lang] = RawString(str)
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
