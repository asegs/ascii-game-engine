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

//all written text will not have a background ID symbol
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

func (t * Terminal) placeAt(message string, row int, col int,txtLen int){
	t.moveTo(row,col)
	t.printRender(message,txtLen)

}

func (t * Terminal) updateAtPos(row int,col int,record * Recorded){
	for i := 1; i < t.Depth;i++{
		t.DataHistory[row][col][i - 1] = t.DataHistory[row][col][i]
	}
	t.DataHistory[row][col][t.Depth - 1] = record
}

func (t * Terminal) undoAtPos(row int,col int){
	for i := t.Depth - 2;i>=0;i--{
		t.DataHistory[row][col][i + 1] = t.DataHistory[row][col][i]
	}
	t.DataHistory[row][col] = t.DataHistory[row][col]
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



func initContext () * Context {
	return &Context{
		Format:    "",
		Modifiers: make([] string,0),
	}
}

func (ctx * Context) addSimpleStyle(styleConst int) * Context{
	ctx.Modifiers = append(ctx.Modifiers,fmt.Sprintf("%d;",styleConst))
	return ctx
}

func (ctx * Context) addRgbStyleFg(r int,g int,b int) * Context{
	ctx.Modifiers = append(ctx.Modifiers,fmt.Sprintf("38;2;%d;%d;%d;",r,g,b))
	return ctx
}

func (ctx * Context) addRgbStyleBg(r int, g int,b int) * Context{
	ctx.Modifiers = append(ctx.Modifiers,fmt.Sprintf("48;2;%d;%d;%d;",r,g,b))
	return ctx
}

//only removes 1, maybe fine
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

func (ctx * Context) addStyleRaw (modifier string) * Context {
	ctx.Modifiers = append(ctx.Modifiers,modifier)
	return ctx
}

func (ctx * Context) compile () * Context{
	newFmt := "\033["
	for _,modifier := range ctx.Modifiers {
		newFmt += modifier
	}
	newFmt = newFmt[0:len(newFmt) - 1] + "m%s\033[0m"
	ctx.Format = newFmt
	return ctx
}

func (t * Terminal) writeStyleAt(style * Context,text string,row int,col int){
	text = t.trimAndUpdateString(style,text)
	t.placeAt(fmt.Sprintf(style.Format,text),row,col,len(text))
}

func (t * Terminal) writeStyleHere(style * Context,text string){
	text = t.trimAndUpdateString(style,text)
	t.printRender(fmt.Sprintf(style.Format,text),len(text))
}

func (t * Terminal) undoConditional(row int,col int,match byte,matchForeground bool){
	if matchForeground {
		if t.DataHistory[row][col][t.Depth - 1].ShownSymbol == match{
			t.undoAtPos(row,col)
		}
	}else{
		if t.DataHistory[row][col][t.Depth - 1].BackgroundCode == match{
			t.undoAtPos(row,col)
		}
	}
}

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

func (t * Terminal) sendPlaceCharFormat(char byte, row int, col int, format *Context, code byte) {
	t.CustomFeed <- func(terminal *Terminal) {
		terminal.placeCharFormat(char,row,col,format,code)
	}
}

//sends a function that associates a fixed char with a style detail when called
func (t * Terminal) sendCharAssociation(char byte,recorded * Recorded) {
	t.CustomFeed <- func(term *Terminal) {
		term.Associations[char] = recorded
	}
}

//sends a function that moves to a coordinate and prints text with a style when called
func (t * Terminal) sendPrintStyleAtCoord(style * Context,row int,col int,text string) {
	t.CustomFeed <- func(term *Terminal) {
		term.moveTo(row,col)
		term.writeStyleHere(style,text)
	}
}

//sends a function that when called places a key character at a coordinate
func (t * Terminal) sendPlaceCharAtCoord(char byte,row int,col int) {
	t.CustomFeed <- func(term *Terminal) {
			t.placeCharLookup(char,row,col)
	}
}
//bake undo into send messages

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

//composes function that when called undoes a cell if it matches a character
func (t * Terminal) sendUndoAtLocationConditional(row int,col int,match byte,matchFg bool){
	t.CustomFeed <- func(term *Terminal) {
		term.undoConditional(row,col,match,matchFg)
	}
}

func (t * Terminal) assoc(char byte,format * Context,fg byte){
	t.sendCharAssociation(char,&Recorded{
		Format: format,
		ShownSymbol: fg,
		BackgroundCode: char,
	})
}

//possibility of register happening after first character is sent
func (t * Terminal) handleRenders(){
	var custom func (t * Terminal)
	for true{
	 custom = <- t.CustomFeed//could send normally, let's make this performance critical though
		custom(t)
	}
}

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



