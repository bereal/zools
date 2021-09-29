package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bereal/zools/pkg/fonts"
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

func main() {
	cmd := &cobra.Command{
		Use: "zools [cmd] [options]",
	}

	packCmd := &cobra.Command{
		Use:  "pack-font file1 [...file2]",
		Run:  packFont,
		Args: cobra.MinimumNArgs(1),
	}

	packCmd.Flags().StringP("output", "o", "", "")

	cmd.AddCommand(packCmd)

	err := cmd.Execute()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
}
