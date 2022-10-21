package main

type AbstractClient interface {
	DrawAt(fg byte, bg byte, row int, col int)
	Init(defaultFg byte, defaultBg byte, rows int, cols int)
	SetWindow(window *ClientWindow)
}
