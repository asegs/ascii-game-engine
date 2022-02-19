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
	}else if text == "N" {
		runNetworked()
	}
}