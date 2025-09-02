package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bereal/zools/pkg/asm"
	"github.com/bereal/zools/pkg/fonts"
	"github.com/bereal/zools/pkg/maps"
	"github.com/bereal/zools/pkg/sprites"
	"github.com/bereal/zools/pkg/text"
	"github.com/spf13/cobra"
)

func check(err error, args ...string) {
	if err != nil {
		if len(args) > 0 {
			log.Fatalf("%s (%+v)", err.Error(), args)
		}
		log.Fatal(err)
	}
}

func encodeFont(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	font := fonts.NewFont()
	for _, infile := range args {
		f, err := os.Open(infile)
		check(err)

		defer f.Close()
		check(font.ReadYaml(f))
	}

	f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	check(err)

	b := asm.NewBuilder()
	font.EncodeAsm(b)
	b.Write(f)
	f.Close()
}

func parseSize(s string) (w int, h int) {
	re := regexp.MustCompile(`^(\d+)+x(\d+)$`)
	parts := re.FindStringSubmatch(s)
	if len(parts) == 0 {
		log.Fatalf("Incorrect size format: %s", s)
	}
	w, _ = strconv.Atoi(parts[1])
	h, _ = strconv.Atoi(parts[2])
	return
}

func splitTileset(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	if output == "" {
		basename := strings.TrimSuffix(args[0], path.Ext(args[0]))
		output = basename + "-%d.png"
	}
	size, _ := cmd.Flags().GetString("size")
	w, h := parseSize(size)

	tileset, err := sprites.ReadTilePNG(args[0])
	check(err)

	for i, s := range tileset.Split(w, h) {
		outputName := fmt.Sprintf(output, i)
		out, err := os.OpenFile(outputName, os.O_CREATE|os.O_RDWR, 0644)
		check(err)
		check(s.EncodePNG(out))
		defer out.Close()
	}
}

func encodeSprite(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	flipV, _ := cmd.Flags().GetBool("flip-vertical")
	invert, _ := cmd.Flags().GetBool("invert")
	masked, _ := cmd.Flags().GetBool("masked")
	direction, _ := cmd.Flags().GetString("direction")
	encoding, _ := cmd.Flags().GetString("encoding")

	var encode func(s sprites.Sprite) []byte
	switch direction {
	case "rows":
		encode = func(s sprites.Sprite) []byte { return s.EncodeByRows(masked) }
	case "columns":
		encode = func(s sprites.Sprite) []byte { return s.EncodeByColumns(masked) }
	case "zigzag":
		encode = func(s sprites.Sprite) []byte { return s.EncodeZigZag(masked) }
	case "cell":
		encode = func(s sprites.Sprite) []byte { return s.EncodeByCell(masked) }
	case "fifth-angel":
		encode = func(s sprites.Sprite) []byte { return s.EncodeFifthAngel() }
	default:
		log.Fatalf("Invalid direction: %s", direction)
	}

	out, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	check(err)
	defer out.Close()

	sort.Strings(args)
	for i, arg := range args {
		fmt.Printf("Processing %s\n", arg)
		f, err := os.Open(arg)
		check(err, args...)

		size, _ := cmd.Flags().GetString("size")
		w, h := parseSize(size)
		sprites, err := sprites.ReadSpriteSheet(f, w, h)
		check(err)
		check(f.Close())
		for j, s := range sprites {
			if flipV {
				s = s.FlipV()
			}
			if invert {
				s = s.Invert()
			}

			data := encode(s)
			switch encoding {
			case "binary":
				_, err := out.Write(data)
				check(err)
			case "asm":
				// TODO use the asm package when it's ready
				label := fmt.Sprintf("sprite_%d", i)
				if len(sprites) > 1 {
					label += fmt.Sprintf("_%d", j)
				}
				_, err = fmt.Fprintf(out, "%s:\n", label)
				check(err)
				for offs := 0; offs < len(data); offs += 8 {
					chunk := data[offs : offs+8]
					db := make([]string, len(chunk))
					for i, b := range chunk {
						db[i] = fmt.Sprintf("0x%02x", b)
					}
					_, err = fmt.Fprintf(out, "\tdb %s\n", strings.Join(db, ", "))
					check(err)
				}
			}
		}
	}
}

func encodeTile(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	encoding, _ := cmd.Flags().GetString("encoding")
	if encoding != "asm" && encoding != "binary" {
		log.Fatalf("Invalid encoding: %s", encoding)
	}

	out, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	check(err)
	defer out.Close()

	var tiles []*sprites.Tile

	for _, arg := range args {
		ext := path.Ext(arg)
		switch ext {
		case ".png":
			tile, err := sprites.ReadTilePNG(arg)
			check(err)
			tiles = append(tiles, tile)
		case ".tsx":
			tileset, err := sprites.ReadTSX(arg)
			check(err)
			tiles = append(tiles, tileset...)
		default:
			log.Fatalf("Unknown tileset file extension: %s", arg)
		}
	}

	if encoding == "asm" {
		fmt.Fprintln(out, "tiles_table:\n\tdw empty_tile")
		for _, tile := range tiles {
			_, err = fmt.Fprintf(out, "\tdw %s\n", tile.Name)
			check(err)
		}
		fmt.Fprintf(out, "empty_tile:\t.12 db 0\n")
	}

	for _, tile := range tiles {
		switch encoding {
		case "binary":
			_, err = out.Write(tile.EncodeBinary(0))
			check(err)
		case "asm":
			b := asm.NewBuilder()
			tile.EncodeAsm(0, b)
			check(b.Write(out))
		}
	}
}

func encodeMap(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	encoding, _ := cmd.Flags().GetString("encoding")
	if encoding != "asm" && encoding != "binary" {
		log.Fatalf("Invalid encoding: %s", encoding)
	}

	out, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0644)
	check(err)
	defer out.Close()

	var m maps.Map

	ext := path.Ext(args[0])
	switch ext {
	case ".tmx":
		m, err = maps.ReadTMX(args[0])
		check(err)
	default:
		log.Fatalf("Unknown map file extension: %s", ext)
	}

	switch encoding {
	case "binary":
		_, err = out.Write(m.EncodeBinary())
		check(err)
	case "asm":
		code := m.EncodeAsm()
		_, err = out.WriteString(strings.Join(code, "\n") + "\n")
		check(err)
	}
}

func encodeText(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	langs, _ := cmd.Flags().GetString("langs")
	if langs == "" {
		log.Fatalf("No languages specified")
	}
	langList := strings.Split(langs, ",")

	out, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	check(err)
	defer out.Close()

	f, err := os.Open(args[0])
	check(err)
	defer f.Close()

	bundle, err := text.ReadI18nBundle(f)
	check(err)

	if mono, _ := cmd.Flags().GetBool("mono"); mono {
		err = bundle.EncodeMono(langList[0], out)
		check(err)
	}
}

func main() {
	cmd := &cobra.Command{
		Use: "zools [cmd] [options]",
	}

	encodeFontCmd := &cobra.Command{
		Use:  "encode-fonts file1 [...file2]",
		Run:  encodeFont,
		Args: cobra.MinimumNArgs(1),
	}

	encodeFontCmd.Flags().StringP("output", "o", "", "")

	encodeSpriteCmd := &cobra.Command{
		Use:  "encode-sprite [files]",
		Run:  encodeSprite,
		Args: cobra.MinimumNArgs(1),
	}

	encodeSpriteCmd.Flags().StringP("output", "o", "", "")
	encodeSpriteCmd.Flags().BoolP("flip-vertical", "", false, "")
	encodeSpriteCmd.Flags().BoolP("invert", "", false, "")
	encodeSpriteCmd.Flags().StringP("size", "s", "16x16", "Size WxH")
	encodeSpriteCmd.Flags().BoolP("masked", "m", false, "")
	encodeSpriteCmd.Flags().StringP("direction", "d", "rows", "encoding direction")
	encodeSpriteCmd.Flags().StringP("encoding", "e", "binary", "encoding (binary, asm)")

	splitTile := &cobra.Command{
		Use:  "split-tile file1",
		Run:  splitTileset,
		Args: cobra.ExactArgs(1),
	}

	splitTile.Flags().StringP("output", "o", "", "")
	splitTile.Flags().StringP("size", "s", "16x16", "Size WxH")

	encodeTiles := &cobra.Command{
		Use:  "encode-tiles file",
		Run:  encodeTile,
		Args: cobra.MinimumNArgs(1),
	}
	encodeTiles.Flags().StringP("output", "o", "", "")
	encodeTiles.Flags().StringP("encoding", "e", "binary", "encoding (binary, asm)")
	// encodeTiles.Flags().IntP("add-attr", "a", 0, "add attribute to the tile by OR")

	encodeMap := &cobra.Command{
		Use:  "encode-map file",
		Run:  encodeMap,
		Args: cobra.ExactArgs(1),
	}
	encodeMap.Flags().StringP("output", "o", "", "")
	encodeMap.Flags().StringP("encoding", "e", "binary", "encoding (binary, asm)")

	encodeText := &cobra.Command{
		Use:  "encode-text file",
		Run:  encodeText,
		Args: cobra.ExactArgs(1),
	}

	encodeText.Flags().StringP("langs", "l", "", "")
	encodeText.Flags().StringP("output", "o", "", "")
	encodeText.Flags().BoolP("mono", "m", false, "do not generate i18n refs, only one language")

	cmd.AddCommand(encodeFontCmd, encodeSpriteCmd, splitTile, encodeTiles, encodeMap, encodeText)

	err := cmd.Execute()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
}
