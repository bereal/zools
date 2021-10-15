package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/bereal/zools/pkg/fonts"
	"github.com/bereal/zools/pkg/sprites"
	"github.com/spf13/cobra"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func packFont(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	font := fonts.NewFont()
	for _, infile := range args {
		f, err := os.Open(infile)
		check(err)

		defer f.Close()
		check(font.ReadYaml(f))
	}

	f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0644)
	check(err)

	font.Write(f)
	f.Close()
}

func parseSize(s string) (w int, h int) {
	re := regexp.MustCompile("^(\\d+)+x(\\d+)$")
	parts := re.FindStringSubmatch(s)
	if len(parts) == 0 {
		log.Fatalf("Incorrect size format: %s", s)
	}
	w, _ = strconv.Atoi(parts[1])
	h, _ = strconv.Atoi(parts[2])
	return
}

func encodeSprite(cmd *cobra.Command, args []string) {
	output := cmd.Flags().Lookup("output").Value.String()
	flipV, _ := cmd.Flags().GetBool("flip-vertical")
	invert, _ := cmd.Flags().GetBool("invert")
	masked, _ := cmd.Flags().GetBool("masked")
	direction, _ := cmd.Flags().GetString("direction")

	var encode func(s sprites.Sprite) []byte
	if direction == "rows" {
		encode = func(s sprites.Sprite) []byte { return s.EncodeByColumns(masked) }
	} else if direction == "columns" {
		encode = func(s sprites.Sprite) []byte { return s.EncodeByRows(masked) }
	} else if direction == "zigzag" {
		encode = func(s sprites.Sprite) []byte { return s.EncodeZigZag(masked) }
	} else {
		log.Fatalf("Invalid direction: %s", direction)
	}

	f, err := os.Open(args[0])
	check(err)

	size, _ := cmd.Flags().GetString("size")
	w, h := parseSize(size)
	sprites, err := sprites.ReadSpriteSheet(f, w, h)
	check(err)

	out, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0644)
	check(err)
	defer out.Close()

	for _, s := range sprites {
		if flipV {
			s = s.FlipV()
		}
		if invert {
			s = s.Invert()
		}
		_, err := out.Write(encode(s))
		check(err)
	}
}

func main() {
	cmd := &cobra.Command{
		Use: "zools [cmd] [options]",
	}

	packFontCmd := &cobra.Command{
		Use:  "pack-font file1 [...file2]",
		Run:  packFont,
		Args: cobra.MinimumNArgs(1),
	}

	packFontCmd.Flags().StringP("output", "o", "", "")

	encodeSpriteCmd := &cobra.Command{
		Use:  "encode-sprite file1",
		Run:  encodeSprite,
		Args: cobra.ExactArgs(1),
	}

	encodeSpriteCmd.Flags().StringP("output", "o", "", "")
	encodeSpriteCmd.Flags().BoolP("flip-vertical", "", false, "")
	encodeSpriteCmd.Flags().BoolP("invert", "", false, "")
	encodeSpriteCmd.Flags().StringP("size", "s", "16x16", "Size WxH")
	encodeSpriteCmd.Flags().BoolP("masked", "m", false, "")
	encodeSpriteCmd.Flags().StringP("direction", "d", "rows", "encoding direction")

	cmd.AddCommand(packFontCmd, encodeSpriteCmd)

	err := cmd.Execute()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
}
