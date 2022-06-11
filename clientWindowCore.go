package main

import "fmt"

//An abstraction for the characters needed to move the cursor on the xy plane.
type Direction byte
const (
	LEFT Direction = 'D'
	RIGHT = 'C'
	DOWN = 'B'
	UP = 'A'
)

//The maximum number of custom functions the terminal can hold in memory via channel.
const MAX_MESSAGES int = 1000

type TilePair struct {
	ShownSymbol byte
	BackgroundCode byte
}

type ClientWindow struct {
	Height int
	Width int
	CustomFeed chan func()
	DataHistory [][] * HistoryStack
	DefaultTile * TilePair
	Renderer AbstractClient
}

func createClientWindow (height int, width int, defaultTile * TilePair, renderer AbstractClient) * ClientWindow {
	renderer.Init(defaultTile,height,width)
	stored := make([][] * HistoryStack,height)
	for i:=0;i<height;i++{
		sRow := make([] * HistoryStack,width)
		for b := 0;b<width;b++ {
			sRow[b] = &HistoryStack{
				Top: &HistoryNode{
					Record:   defaultTile,
					Previous: nil,
				},
				Length: 1,
			}
		}
		stored[i] = sRow
	}
	window := &ClientWindow{
		CustomFeed: make(chan func(),MAX_MESSAGES),
		DataHistory: stored,
		Height:height,
		Width:width,
		DefaultTile: defaultTile,
		Renderer: renderer,
	}
	go window.handleRenders()
	return window
}

/**
Runs a loop pulling functions from the function queue and running them on the given terminal.
Only will perform functions sequentially.
*/
func (w * ClientWindow) handleRenders(){
	var custom func ()
	for true{
		custom = <- w.CustomFeed
		custom()
	}
}

/**
Shifts every history item for a current cell back one and inserts a new current one.
Loses the oldest item forever.
*/
func (w * ClientWindow) updateAtPos(row int,col int,record * TilePair){
	w.DataHistory[row][col].add(record)
}

/**
Replaces the current terminal state at a given coordinate with the previous one, discards the current state.
Sets the oldest state to the default state.
Prints the data at the previous state and performs all standard printing operations.

Could possibly use printRender.
*/
func (w * ClientWindow) undoAtPos(row int,col int){
	w.DataHistory[row][col].pop()
	w.Renderer.DrawAt(w.DataHistory[row][col].top(),row,col)
}

/**
Performs an undo on a certain cell on the terminal given that a certain byte matches the expected value.
Can be used to match foreground values or background using boolean.

The reason for the conditional is that if something has already overwritten the space, it should not be reset upon leaving it.
*/
func (w * ClientWindow) undoConditional(row int,col int,match byte,matchForeground bool){
	if matchForeground {
		if w.DataHistory[row][col].top().ShownSymbol == match{
			w.undoAtPos(row,col)
		}else{
			LogString(fmt.Sprintf("%d,%d",row,col))
			LogString(string(w.DataHistory[row][col].top().ShownSymbol))
			LogString("Didn't perform undo due to overwrite.")
			w.DataHistory[row][col].removeFirstMatch(match,matchForeground)
		}
	}else{
		if w.DataHistory[row][col].top().BackgroundCode == match{
			w.undoAtPos(row,col)
		}else{
			LogString("Didn't perform undo due to overwrite.")
			w.DataHistory[row][col].removeFirstMatch(match,matchForeground)
		}
	}
}

/**
Composes and queues a function that looks up a certain character in the map and prints it with the associated Recorded object.
*/
func (w * ClientWindow) sendPlaceCharAtCoord(char byte,row int,col int) {
	w.CustomFeed <- func() {
		w.Renderer.DrawAt(&TilePair{
			ShownSymbol:    char,
			BackgroundCode: w.DataHistory[row][col].top().BackgroundCode,
		}, row, col)
	}
}

/**
Composes and queues a function that checks to see if a character has a mapping.
If so, performs a conditional undo with matchFg and writes the character at the new location.
*/
func (w * ClientWindow) sendPlaceCharAtCoordCondUndo(char byte,row int,col int,lastRow int,lastCol int,match byte,matchFg bool) {
	w.CustomFeed <- func() {
		w.undoConditional(lastRow,lastCol,match,matchFg)
		w.Renderer.DrawAt(&TilePair{
			ShownSymbol:    char,
			BackgroundCode: w.DataHistory[row][col].top().BackgroundCode,
		}, row, col)
	}
}

/**
Composed and queues a function that conditionally undoes a character at a given location.
*/
func (w * ClientWindow) sendUndoAtLocationConditional(row int,col int,match byte,matchFg bool){
	w.CustomFeed <- func() {
		w.undoConditional(row,col,match,matchFg)
	}
}

