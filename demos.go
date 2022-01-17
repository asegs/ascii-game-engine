package main

import (
	"bufio"
	"fmt"
	"os"
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
		runWordleDemo()
	}else if text == "N" {
		runNetworked()
	}
}
