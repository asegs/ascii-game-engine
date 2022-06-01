package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func main () {
	go HandleLog()
	reader := bufio.NewScanner(os.Stdin)
	fmt.Println("Choose the demo you are interested in:")
	fmt.Println("(N)etworked Chase")
	fmt.Println("(W)ordle")
	fmt.Println("(A)dvanced Networked Chase - In progress")
	reader.Scan()
	text := reader.Text()
	if text == "W" {
		fmt.Println("Enter the upper percentile to get all words below (0 is easiest):")
		reader.Scan()
		upper,_ := strconv.Atoi(reader.Text())
		fmt.Println("Enter the lower percentile to get all words above (100 is hardest):")
		reader.Scan()
		lower,_ := strconv.Atoi(reader.Text())
		runWordleDemo(upper,lower)
	} else if text == "A" {
		fmt.Println("(S)erver mode, or (C)lient mode?:")
		reader.Scan()
		entry := reader.Text()
		if entry == "S" {
			serve()
		}else {
			render()
		}
	}
}
