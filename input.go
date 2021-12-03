package main

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	ESCAPE byte = 27
	ENTER = 10
	BRACKET = 91
	MOVE_LEFT = 131
	MOVE_RIGHT = 130
	MOVE_DOWN = 129
	MOVE_UP = 128
	BACKSPACE = 127
	BACKSLASH = 92
)

type StdIn struct {
	events chan byte
}

func tput(arg string) error {
	cmd := exec.Command("tput", arg)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func initializeInput() * StdIn {
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	err := tput("civis")
	if err != nil {
		fmt.Println(err.Error())
	}
	scanner := make(chan byte,1000)
	input := &StdIn{events: scanner}
	go input.scanForInput()
	return input

}


/**
I want to support sending escape char if no new chars are entered with it
However input reads are blocking and so a timeout won't work
 */
func (s * StdIn) scanForInput(){
	// restore the echoing state when exiting
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()
	defer exec.Command("clear").Run()
	defer func() {
		err := tput("cnorm")
		if err != nil {
			fmt.Println(err.Error())
		}
	}()
	var buf = make([]byte, 1)
	var c byte

	ranksToMovement := 0
	for true{
		os.Stdin.Read(buf)
		c = buf[0]
		if c == ESCAPE{
			ranksToMovement ++
		}else if c == BRACKET && ranksToMovement == 1{
			ranksToMovement ++
		}else if ranksToMovement == 2 && 65 <= c && c <= 68{
			s.events <- c + 63
			ranksToMovement = 0
		}else{
			s.events <- c
		}
	}
}
