package main

import (
	"os/exec"
)

/**
This is currently built using FFPlay, provided by FFMpeg.
Right now, the reason is that this project is to be free of dependencies.
However, this may be difficult to handle on every OS, so it may end up using a library for sound.

If necessary, Go supports cancelling processes after any amount of time via callback,
so we could easily add a "playFor" function.
 */

type Audio struct {
	Filename string
	Handle * exec.Cmd
}

func play(filename string) (error, * Audio ) {
	cmd := exec.Command("ffplay","nodisp","autoexit",filename)
	err := cmd.Start()
	return err,&Audio{Filename: filename,Handle: cmd}
}

func (a * Audio) stop() error {
	err := a.Handle.Process.Kill()
	return err
}

