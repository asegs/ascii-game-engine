package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type SavedFrame struct {
	Colors []string `json:"colors"`
	Pixels [][]int `json:"pixels"`
}

type Frame struct {
	Colors [] * Context
	Pixels [][] int
}

func hexStrToDec(twoChars string)int {
	output, err := strconv.ParseUint(twoChars, 16, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	return int(output)
}

func buildFrame(filename string) * Frame {
	file,err := ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
	}
	var simpleFrame * SavedFrame
	err = json.Unmarshal(file,&simpleFrame)
	if err != nil {
		fmt.Println(err.Error())
	}
	colors := make([] * Context,len(simpleFrame.Colors))
	for i,color := range simpleFrame.Colors {
		r := hexStrToDec(color[1:3])
		g := hexStrToDec(color[3:5])
		b := hexStrToDec(color[5:7])
		colors[i] = initContext().addRgbStyleBg(r,g,b).finish()
	}
	return &Frame{
		Colors: colors,
		Pixels: simpleFrame.Pixels,
	}
}

func (t * Terminal) drawFrame (frame * Frame){
	for i := 0 ; i < len(frame.Pixels) ; i ++ {
		t.CustomFeed <- func(terminal *Terminal) {
			for b := 0 ; b < len(frame.Pixels[0]) ; b ++ {
				//place char at the right coordinates
			}
		}
	}
}
