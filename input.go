package main

import (
	"os"
	"os/exec"
	"time"
)

const (
	ESCAPE byte = 27
	BRACKET = 91
	MOVE_LEFT = 131
	MOVE_RIGHT = 130
	MOVE_DOWN = 129
	MOVE_UP = 128
)

type StdIn struct {
	events chan byte
}

func initializeInput() * StdIn {
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
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
	var buf = make([]byte, 1)
	var c byte

	ranksToMovement := 0
	for true{
		os.Stdin.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
		os.Stdin.Read(buf)
		c = buf[0]
		if c == ESCAPE{
			ranksToMovement ++
		}else if c == BRACKET && ranksToMovement == 1{
			ranksToMovement ++
		}else if ranksToMovement == 2 && 65 <= c && c <= 68{
			switch c {
			case 65:
				s.events <- MOVE_UP
				break
			case 66:
				s.events <- MOVE_DOWN
				break
			case 67:
				s.events <- MOVE_RIGHT
				break
			case 68:
				s.events <- MOVE_LEFT
				break
			}
			ranksToMovement = 0
		}else{
			s.events <- c
		}
	}
}