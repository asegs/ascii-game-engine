package main

//An abstraction for the characters needed to move the cursor on the xy plane.
type Direction byte

const (
	LEFT  Direction = 'D'
	RIGHT           = 'C'
	DOWN            = 'B'
	UP              = 'A'
)

//The maximum number of custom functions the terminal can hold in memory via channel.
const MAX_MESSAGES int = 1000

type ClientWindow struct {
	Height      int
	Width       int
	CustomFeed  chan func()
	DataHistory [][]*History
	DefaultFg   byte
	DefaultBg   byte
	Renderer    AbstractClient
}

func createClientWindow(height int, width int, defaultFg byte, defaultBg byte, renderer AbstractClient) *ClientWindow {
	renderer.Init(defaultFg, defaultBg, height, width)
	stored := make([][]*History, height)
	for i := 0; i < height; i++ {
		sRow := make([]*History, width)
		for b := 0; b < width; b++ {
			sRow[b] = &History{
				Fg: &HistoryStack{
					Top: &HistoryNode{
						Record:   defaultFg,
						Previous: nil,
					},
					Length: 1,
				},
				Bg: &HistoryStack{
					Top: &HistoryNode{
						Record:   defaultBg,
						Previous: nil,
					},
					Length: 1,
				},
			}
		}
		stored[i] = sRow
	}
	window := &ClientWindow{
		CustomFeed:  make(chan func(), MAX_MESSAGES),
		DataHistory: stored,
		Height:      height,
		Width:       width,
		DefaultFg:   defaultFg,
		DefaultBg:   defaultBg,
		Renderer:    renderer,
	}
	renderer.SetWindow(window)
	go window.handleRenders()
	return window
}

/**
Runs a loop pulling functions from the function queue and running them on the given terminal.
Only will perform functions sequentially.
*/
func (w *ClientWindow) handleRenders() {
	var custom func()
	for true {
		custom = <-w.CustomFeed
		custom()
	}
}

/**
Shifts every history item for a current cell back one and inserts a new current one.
Loses the oldest item forever.
*/
func (w *ClientWindow) updateAtPos(row int, col int, char byte, updateFg bool) {
	if updateFg {
		w.DataHistory[row][col].Fg.add(char)
	} else {
		w.DataHistory[row][col].Bg.add(char)
	}

}

/**
Performs an undo on a certain cell on the terminal given that a certain byte matches the expected value.
Can be used to match foreground values or background using boolean.

The reason for the conditional is that if something has already overwritten the space, it should not be reset upon leaving it.
*/
func (w *ClientWindow) undoConditional(row int, col int, match byte, matchForeground bool) {
	if matchForeground {
		w.DataHistory[row][col].Fg.removeLastMatching(match)
	} else {
		w.DataHistory[row][col].Bg.removeLastMatching(match)
	}
	w.Renderer.DrawAt(w.DataHistory[row][col].Fg.top(), w.DataHistory[row][col].Bg.top(), row, col, false)
}

func (w *ClientWindow) placeFgCharAtCoord(char byte, row int, col int, bulk bool) {
	w.updateAtPos(row, col, char, true)
	w.Renderer.DrawAt(w.DataHistory[row][col].Fg.top(), w.DataHistory[row][col].Bg.top(), row, col, bulk)
}

func (w *ClientWindow) placeBgCharAtCoord(char byte, row int, col int, bulk bool) {
	w.updateAtPos(row, col, char, false)
	w.Renderer.DrawAt(w.DataHistory[row][col].Fg.top(), w.DataHistory[row][col].Bg.top(), row, col, bulk)

}

/**
Composes and queues a function that looks up a certain character in the map and prints it with the associated Recorded object.
*/
func (w *ClientWindow) sendPlaceFgCharAtCoord(char byte, row int, col int, bulk bool) {
	w.CustomFeed <- func() {
		w.placeFgCharAtCoord(char, row, col, bulk)
	}
}

func (w *ClientWindow) sendPlaceBgCharAtCoord(char byte, row int, col int, bulk bool) {
	w.CustomFeed <- func() {
		w.placeBgCharAtCoord(char, row, col, bulk)
	}
}

/**
Composes and queues a function that checks to see if a character has a mapping.
If so, performs a conditional undo with matchFg and writes the character at the new location.
*/
func (w *ClientWindow) sendPlaceFgCharAtCoordCondUndo(char byte, row int, col int, lastRow int, lastCol int, match byte) {
	w.CustomFeed <- func() {
		w.undoConditional(lastRow, lastCol, match, true)
		w.placeFgCharAtCoord(char, row, col, false)
	}
}

func (w *ClientWindow) sendPlaceBgCharAtCoordCondUndo(char byte, row int, col int, lastRow int, lastCol int, match byte) {
	w.CustomFeed <- func() {
		w.undoConditional(lastRow, lastCol, match, false)
		w.placeBgCharAtCoord(char, row, col, false)
	}
}

/**
Composed and queues a function that conditionally undoes a character at a given location.
*/
func (w *ClientWindow) sendUndoAtLocationConditional(row int, col int, match byte, matchFg bool) {
	w.CustomFeed <- func() {
		w.undoConditional(row, col, match, matchFg)
	}
}
