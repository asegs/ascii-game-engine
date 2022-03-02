package main

import (
	"fmt"
	"strings"
)

//An abstraction for the characters needed to move the cursor on the xy plane.
type Direction rune
const (
	LEFT Direction = 'D'
	RIGHT = 'C'
	DOWN = 'B'
	UP = 'A'
)

type HistoryNode struct {
	Record * Recorded
	Previous * HistoryNode
}

type HistoryStack struct {
	Top * HistoryNode
	Length int
}

func (h * HistoryStack) add(r * Recorded) {
	newTop := &HistoryNode{
		Record:   r,
		Previous: h.Top,
	}
	h.Top = newTop
	h.Length++
}

func (h * HistoryStack) pop() * Recorded {
	toReturn := h.Top
	if toReturn == nil {
		return nil
	}
	h.Top = h.Top.Previous
	return toReturn.Record
}

//The maximum number of custom functions the terminal can hold in memory via channel.
const MAX_MESSAGES int = 1000

//The char sequence to RESET the cursor style.
const RESET string = "\033[0m"

/**
A format string that contains a "%s" for the text to place in between style escape chars.
Can be compiled to update this format string from style modifiers unlimited times.
Format: The string that has the current "compiled" format, for example:
	"\033[38;2;200;12;181;m%s\033[0m"
Modifiers: The array of style modifiers, some seen in the example above, to apply when the Context is compiled.
 */
type Context struct {
	Format string
	Modifiers [] string
}

/**
The current state of a tile on the terminal.
Format: The context that wrote that pixel.
ShownSymbol: The foreground text character.
BackgroundCode: The symbol describing the style of the background.

For example:
	Format: Pointer to context, like "draw red background with white foreground".
	ShownSymbol: '@' to show player.
	BackgroundCode: '1' to indicate wall.
 */
type Recorded struct {
	Format * Context
	ShownSymbol byte
	BackgroundCode byte
}

/**
The local variables of the terminal instance.

Row: What row the cursor is on.
Col: What column the cursor is on.
Height: The max height of the terminal.
Width: The max width of the terminal.
CustomFeed: The channel by which composed functions come in to be applied.
Associations: A quick lookup of Recorded objects mapped from characters.
DataHistory: A table of what state each cell is in and a Depth length history of previous states.
Depth: The length of the history stored for each cell.
DefaultRecorded: The default Recorded Context for a given cell.
 */
type Terminal struct {
	Row int
	Col int
	Height int
	Width int
	CustomFeed chan func(terminal * Terminal)
	Associations map[byte] * Recorded
	DataHistory [][][] * Recorded
	Depth int
	DefaultRecorded * Recorded
}

/**
Initializes a terminal instance with a given height and width, the defaultRecorded value to be saved in memory and the history depth.

Writes the empty box of the given size to the terminal to clear a space.
Writes the necessary terminal cells into memory.
Moves the cursor back to the top left.
Starts the background thread to receive functions and apply them.

Finally, returns the reference to the terminal object.
 */
func createTerminal(height int,width int,defaultRecorded * Recorded,history int)*Terminal{
	stored := make([][][] * Recorded,height)
	for i:=0;i<height;i++{
		sRow := make([][] * Recorded,width)
		for b := 0;b<width;b++{
			print(" ")
			records := make([] * Recorded, history)
			for n := 0;n < history;n++{
				records[n] = defaultRecorded
			}
			sRow[b] = records
		}
		println()
		stored[i] = sRow
	}
	terminal := &Terminal{
		Row:  height,
		Col:  0,
		CustomFeed: make(chan func(terminal * Terminal),MAX_MESSAGES),
		Associations: make(map[byte]*Recorded),
		DataHistory: stored,
		Height:height,
		Width:width,
		Depth:history,
		DefaultRecorded: defaultRecorded,
	}
	terminal.moveTo(0,0)
	go terminal.handleRenders()
	return terminal
}

/**
Given a Direction, or text character, moves the cursor n times in the given direction dir.
Changes the state of the terminal to match the new position.
 */
func (t * Terminal) moveCursor(n int,dir Direction){
	fmt.Printf("\033[%d%c",n,dir)
	switch dir {
	case UP:
		t.Row -= n
		return
	case DOWN:
		t.Row += n
		return
	case LEFT:
		t.Col -= n
		return
	case RIGHT:
		t.Col += n
		return
	}
}

/**
Given a new pair of coordinates, calculates the difference in current position to the new position.
If the new position is out of bounds, sets it to the max value in that direction.
Uses moveCursor to update cursor position by going over the difference in the right direction.
 */
func (t * Terminal) moveTo(newRow int,newCol int){
	if newRow >= t.Height{
		newRow = t.Height - 1
	}

	if newRow < 0{
		newRow = 0
	}

	if newCol >= t.Width{
		newCol = t.Width - 1
	}

	if newCol < 0{
		newCol = 0
	}
	if newRow - t.Row > 0{
		t.moveCursor(newRow - t.Row,DOWN)
	}else if newRow - t.Row < 0 {
		t.moveCursor((newRow - t.Row) * -1,UP)
	}

	if newCol - t.Col > 0{
		t.moveCursor(newCol - t.Col,RIGHT)
	}else if newCol - t.Col < 0 {
		t.moveCursor((newCol - t.Col) * -1,LEFT)
	}
}

/**
Clears n tiles (replaces with nothing) at the given coordinates.  Doesn't update in memory, and not used.
 */
func (t * Terminal) wipeNTilesAt(tiles int, row int, col int){
	t.moveTo(row,col)
	println(RESET)
	for tiles > 0 {
		fmt.Printf(" ")
		tiles --
	}
}

/**
Writes a string to the terminal, shifts the cursor over back to the start after writing.
 */
func (t * Terminal) printRender(message string,txtLen int){
	t.Col += txtLen
	print(message)
	t.moveCursor(txtLen,LEFT)
}

/**
Takes a string of text, writes it to the end of the line and then cuts off the rest.
Updates the proper blocks of memory to track the current state.

One issue is the background code for each cell will never be anything other than ' '.
 */
func (t * Terminal) trimAndUpdateString(style * Context, text string) string{
	over := len(text) + t.Col - (t.Width - 1)
	if over > 0{
		text = text[0:len(text) - over]
	}
	for i := 0;i<len(text);i++{
		t.updateAtPos(t.Row,t.Col + i,&Recorded{
			Format: style,
			ShownSymbol:   text[i],
			BackgroundCode: ' ',
		})
	}
	return text
}

/**
Places a string at a given location using printRender.
Used after substituting message for format character and compiling Context.
 */
func (t * Terminal) placeAt(message string, row int, col int,txtLen int){
	t.moveTo(row,col)
	t.printRender(message,txtLen)

}


/**
Shifts every history item for a current cell back one and inserts a new current one.
Loses the oldest item forever.
 */
func (t * Terminal) updateAtPos(row int,col int,record * Recorded){
	for i := 1; i < t.Depth;i++{
		t.DataHistory[row][col][i - 1] = t.DataHistory[row][col][i]
	}
	t.DataHistory[row][col][t.Depth - 1] = record
}

/**
Replaces the current terminal state at a given coordinate with the previous one, discards the current state.
Sets the oldest state to the default state.
Prints the data at the previous state and performs all standard printing operations.

Could possibly use printRender.
 */
func (t * Terminal) undoAtPos(row int,col int){
	for i := t.Depth - 2;i>=0;i--{
		t.DataHistory[row][col][i + 1] = t.DataHistory[row][col][i]
	}
	t.DataHistory[row][col][0] = t.DefaultRecorded
	t.moveTo(row,col)
	if t.Col >= width - 1{
		return
	}
	t.Col ++
	var character [1] byte
	character[0] = t.DataHistory[row][col][t.Depth - 1].ShownSymbol
	fmt.Printf(t.DataHistory[row][col][t.Depth - 1].Format.Format,character)
	t.moveCursor(1,LEFT)

}


/**
Creates a new Context.  Has no traits to begin with, just prints text.
 */
func initContext () * Context {
	return &Context{
		Format:    "%s",
		Modifiers: make([] string,0),
	}
}

/**
Adds a basic style (preset color, flashing, etc.) to the modifiers array.
 */
func (ctx * Context) addSimpleStyle(styleConst int) * Context{
	ctx.Modifiers = append(ctx.Modifiers,fmt.Sprintf("%d;",styleConst))
	return ctx
}

/**
Adds a background color to the modifiers array.
 */
func (ctx * Context) addRgbStyleFg(r int,g int,b int) * Context{
	ctx.Modifiers = append(ctx.Modifiers,fmt.Sprintf("38;2;%d;%d;%d;",r,g,b))
	return ctx
}

/**
Adds a foreground color to the modifiers array.
*/
func (ctx * Context) addRgbStyleBg(r int, g int,b int) * Context{
	ctx.Modifiers = append(ctx.Modifiers,fmt.Sprintf("48;2;%d;%d;%d;",r,g,b))
	return ctx
}

/**
Removes the first RGB style from the modifiers array.
Often used for composing new styles.
Takes a boolean to remove the foreground style or background style.
 */
func (ctx * Context) removeRgbStyle (fg bool) * Context {
	key := "38;2"
	if !fg {
		key = "48;2"
	}
	for i,modifier := range ctx.Modifiers {
		if strings.Contains(modifier,key) {
			if i == len(ctx.Modifiers) - 1 {
				ctx.Modifiers = ctx.Modifiers[0:i]
			}else {
				ctx.Modifiers = append(ctx.Modifiers[0:i],ctx.Modifiers[i+1:]...)
			}
			return ctx
		}
	}
	return ctx
}

/**
Makes a deep copy of a context.  Often used for composing new styles on the fly.
 */
func (ctx * Context) copyContext () * Context {
	newMods := make([]string,len(ctx.Modifiers))
	for i,modifier := range ctx.Modifiers {
		newMods[i] = modifier
	}
	return &Context{
		Format:    ctx.Format,
		Modifiers: newMods,
	}
}

/**
Returns the color modifier for a certain Context.
Either the foreground or background color based on the boolean.
If none exists, returns empty string.
 */
func (ctx * Context) getColorInfo (fg bool) string {
	key := "38;2"
	if !fg {
		key = "48;2"
	}
	for _,modifier := range ctx.Modifiers {
		if strings.Contains(modifier,key) {
			return modifier
		}
	}
	return ""
}

/**
Adds a raw style (typed out/string) to a given Context.
Often used to pass color info styles pulled directly from another Context.
 */
func (ctx * Context) addStyleRaw (modifier string) * Context {
	ctx.Modifiers = append(ctx.Modifiers,modifier)
	return ctx
}

/**
Builds the format string according to the current style modifiers.
First starts off with the proper escape code.
Next adds each style from the modifiers list.
Next deletes the final semicolon and replaces it with an 'm'.
Next adds an %s for use in printf.
Finally adds a RESET escape code to turn off the escape code for future writes.

Sets format equal to this string.
 */
func (ctx * Context) compile () * Context{
	newFmt := "\033["
	for _,modifier := range ctx.Modifiers {
		newFmt += modifier
	}
	newFmt = newFmt[0:len(newFmt) - 1] + "m%s\033[0m"
	ctx.Format = newFmt
	return ctx
}


/**
Writes text with an associated style at a certain pair of coordinates.
 */
func (t * Terminal) writeStyleAt(style * Context,text string,row int,col int){
	text = t.trimAndUpdateString(style,text)
	t.placeAt(fmt.Sprintf(style.Format,text),row,col,len(text))
}


/**
Writes just the styled text at a location, without loading into history.
Used for loading from saves.
 */
func (t * Terminal) writeStyleAtNoHistory(style * Context,text string,row int,col int){
	t.placeAt(fmt.Sprintf(style.Format,text),row,col,len(text))
}

/**
Performs an undo on a certain cell on the terminal given that a certain byte matches the expected value.
Can be used to match foreground values or background using boolean.

The reason for the conditional is that if something has already overwritten the space, it should not be reset upon leaving it.
 */
func (t * Terminal) undoConditional(row int,col int,match byte,matchForeground bool){
	if matchForeground {
		if t.DataHistory[row][col][t.Depth - 1].ShownSymbol == match{
			t.undoAtPos(row,col)
		}else{
			LogString(fmt.Sprintf("%d,%d",row,col))
			LogString(string(t.DataHistory[row][col][t.Depth-1].ShownSymbol))
			LogString("Didn't perform undo due to overwrite.")
		}
	}else{
		if t.DataHistory[row][col][t.Depth - 1].BackgroundCode == match{
			t.undoAtPos(row,col)
		}else{
			LogString("Didn't perform undo due to overwrite.")
		}
	}
}

/**
If a character exists in the association table, writes the proper format for that char.
Does it at a given location.
 */
func (t * Terminal) placeCharLookup(char byte,row int,col int){
	if format, ok := t.Associations[char]; ok {
		t.moveTo(row,col)
		if t.Col >= width - 1{
			return
		}
		t.Col ++
		var character [1] byte
		character[0] = format.ShownSymbol
		fmt.Printf(format.Format.Format,character)
		t.updateAtPos(row,col,format)
		t.moveCursor(1,LEFT)
	}
}

/**
Writes a certain character with no style at a particular location.
 */
func (t * Terminal) placeCharRaw(char byte,row int,col int){
	t.moveTo(row,col)
	if t.Col >= width - 1{
		return
	}
	t.Col ++
	var character [1] byte
	character[0] = char
	fmt.Printf(t.DataHistory[row][col][t.Depth-1].Format.Format,character)
	t.moveCursor(1,LEFT)
}

/**
Places a certain unassociated character with a provided format.
Requires the data to create a recording, including the format Context, as well as foreground and background values.
Adds the character to the history with a new Recorded value created from the aforementioned data.
 */
func (t * Terminal) placeCharFormat(char byte,row int,col int,format * Context,bgCode byte){
	t.moveTo(row,col)
	if t.Col >= width - 1 {
		return
	}
	t.Col ++
	var character [1] byte
	character[0] = char
	fmt.Printf(format.Format,character)
	t.updateAtPos(row,col,&Recorded{
		Format: format,
		ShownSymbol:   char,
		BackgroundCode:   bgCode,
	})
	t.moveCursor(1,LEFT)
}

/*
Composition functions
 */

/**
Composes and queues a function that places a character with a format at a specific position.
 */
func (t * Terminal) sendPlaceCharFormat(char byte, row int, col int, format *Context, code byte) {
	t.CustomFeed <- func(terminal *Terminal) {
		terminal.placeCharFormat(char,row,col,format,code)
	}
}

/**
Composes and queues a function that associates a Recorded reference with a character.
 */
func (t * Terminal) sendCharAssociation(char byte,recorded * Recorded) {
	t.CustomFeed <- func(term *Terminal) {
		term.Associations[char] = recorded
	}
}

/**
Composes and queues a function prints text with a given style Context.
*/
func (t * Terminal) sendPrintStyleAtCoord(style * Context,row int,col int,text string) {
	t.CustomFeed <- func(term *Terminal) {
		term.writeStyleAt(style,text,row,col)
	}
}

/**
Composes and queues a function that looks up a certain character in the map and prints it with the associated Recorded object.
 */
func (t * Terminal) sendPlaceCharAtCoord(char byte,row int,col int) {
	t.CustomFeed <- func(term *Terminal) {
			t.placeCharLookup(char,row,col)
	}
}

/**
Composes and queues a function that checks to see if a character has a mapping.
If so, performs a conditional undo with matchFg and writes the character at the new location.
 */
func (t * Terminal) sendPlaceCharAtCoordCondUndo(char byte,row int,col int,lastRow int,lastCol int,match byte,matchFg bool) {
	t.CustomFeed <- func(term *Terminal) {
		if format, ok := term.Associations[char]; ok {
			term.undoConditional(lastRow,lastCol,match,matchFg)
			term.moveTo(row,col)
			if term.Col >= width - 1{
				return
			}
			term.Col ++
			var character [1] byte
			character[0] = char
			fmt.Printf(format.Format.Format,character)
			term.updateAtPos(row,col,format)
			term.moveCursor(1,LEFT)
		}
	}
}

/**
Composed and queues a function that conditionally undoes a character at a given location.
 */
func (t * Terminal) sendUndoAtLocationConditional(row int,col int,match byte,matchFg bool){
	t.CustomFeed <- func(term *Terminal) {
		term.undoConditional(row,col,match,matchFg)
	}
}

/**
Composes and queues a function that associates a character with a Recorded reference.
The character associated is used as the background code.
 */
func (t * Terminal) assoc(char byte,format * Context,fg byte){
	t.sendCharAssociation(char,&Recorded{
		Format: format,
		ShownSymbol: fg,
		BackgroundCode: char,
	})
}

/**
Runs a loop pulling functions from the function queue and running them on the given terminal.
Only will perform functions sequentially.
 */
func (t * Terminal) handleRenders(){
	var custom func (t * Terminal)
	for true{
	 custom = <- t.CustomFeed
		custom(t)
	}
}

/**
Find the tile of the terminal where the foreground matches the provided symbol.
Return the Coord instance at which this is true.
 */
func (t * Terminal) getCoordsForCursor(cursor byte) [] * Coord{
	cursors := make([] * Coord,0)
	for row,items := range t.DataHistory {
		for col, item := range items {
			if item[t.Depth - 1].ShownSymbol == cursor {
				cursors = append(cursors,&Coord{
					Row: row,
					Col: col,
				})
			}
		}
	}
	return cursors
}

/**
Only to be used on something where undoing it is not required.
 */
func (t * Terminal) sendRawFmtString(raw string,effectiveSize int, row int, col int){
	t.CustomFeed <- func(terminal *Terminal) {
		t.moveTo(row,col)
		t.printRender(raw,effectiveSize)
	}
}



