package maps

import (
	"encoding/xml"
	"os"
	"strconv"
	"strings"
)

type Row []int

type Map []Row

type tmxMap struct {
	Layers []struct {
		Data struct {
			Encoding string `xml:"encoding,attr"`
			Map      string `xml:",chardata"`
		} `xml:"data"`
	} `xml:"layer"`
}

func ReadTMX(tmpPath string) (Map, error) {
	f, err := os.Open(tmpPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tmx tmxMap
	if err := xml.NewDecoder(f).Decode(&tmx); err != nil {
		return nil, err
	}

	csv := tmx.Layers[0].Data.Map

	res := make(Map, 0)
	for _, rowS := range strings.Split(csv, "\n") {
		rowS = strings.TrimSpace(rowS)
		rowS = strings.TrimSuffix(rowS, ",")
		if rowS == "" {
			continue
		}
		vals := strings.Split(rowS, ",")
		row := make(Row, len(vals))
		for i, v := range vals {
			row[i], err = strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
		}
		res = append(res, row)
	}

	return res, nil
}

func (m Map) EncodeAsm() []string {
	var lines []string
	for _, row := range m {
		rowS := make([]string, len(row))
		for i, v := range row {
			rowS[i] = strconv.Itoa(v)
		}
		lines = append(lines, "\tdb "+strings.Join(rowS, ","))
	}
	return lines
}

func (m Map) EncodeBinary() []byte {
	res := make([]byte, 0)
	for _, row := range m {
		for _, v := range row {
			res = append(res, byte(v))
		}
	}
	return res
}
