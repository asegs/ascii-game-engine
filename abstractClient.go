package main

type AbstractClient interface {
	DrawAt (toDraw * TilePair, row int, col int)
	Init (def * TilePair, rows int, cols int)
	SetWindow (window * ClientWindow)
}
