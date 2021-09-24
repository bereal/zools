package fonts_test

import (
	"testing"

	"github.com/bereal/zools/pkg/fonts"
	"github.com/stretchr/testify/assert"
)

func TestParseLine(t *testing.T) {
	width, data := fonts.ParseCharLine("  %%% %")
	assert.Equal(t, byte(58), data)
	assert.Equal(t, 7, width)
}

func TestParseChar(t *testing.T) {
	data := `
 %%
%  %
%%%%
%  %`
	char := fonts.ParseCharacter(data)
	assert.EqualValues(t, []byte{6 << 4, 9 << 4, 15 << 4, 9 << 4, 0, 0, 0, 4}, char.Encode())
}

func TestFont(t *testing.T) {
	char_1 := `
 %%
%  %
%%%%
%  %`

	char_2 := `
%%
 %%
 % %
 %%`

	f := fonts.NewFont()
	f.ParseChars(map[string]string{
		"a": char_1,
		"ъ": char_2,
	})

	result := f.Encode()
	for i := 0; i < len(result); i++ {
		if result[i] > 2 {
			break
		}
	}

	v1 := result[('a'-32)*8 : ('a'-32)*8+8]
	assert.EqualValues(t, []byte{6 << 4, 9 << 4, 15 << 4, 9 << 4, 0, 0, 0, 4}, v1)

	v2 := result[(223-64)*8 : (223-64)*8+8]
	assert.EqualValues(t, []byte{12 << 4, 6 << 4, 5 << 4, 6 << 4, 0, 0, 0, 4}, v2)
}
