package main

import (
	"bufio"
	"fmt"
	"os"
)

func main () {
	go HandleLog()
	reader := bufio.NewScanner(os.Stdin)
	fmt.Println("(S)erver mode, or (C)lient mode? (T/V):")
	reader.Scan()
	entry := reader.Text()
	if entry == "S" {
		serve()
	}else if entry == "CT" {
		render()
	} else {
		renderVisual()
	}
}
