package main

import (
	"fmt"
)

type Direction rune
const (
	LEFT Direction = 'D'
	RIGHT = 'C'
	DOWN = 'B'
	UP = 'A'
)

const MAX_MESSAGES int = 1000

const RESET string = "\033[0m"

type Context struct {
	Format string
}

type Recorded struct {
	Format * Context
	data byte
	code byte
}

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

func createTerminal(height int,width int,defaultRecorded * Recorded,history int)*Terminal{
	stored := make([][][] * Recorded,height)
	for i:=0;i<height;i++{
		sRow := make([][] * Recorded,width)
		for b := 0;b<width;b++{
			records := make([] * Recorded, history)
			for n := 0;n < history;n++{
				records[n] = defaultRecorded
			}
			sRow[b] = records
		}
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
	for i := 0;i<height;i++ {
		for b := 0;b<width;b++ {
			print(" ")
		}
		println()
	}
	terminal.moveTo(0,0)
	go terminal.handleRenders()
	return terminal
}

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

func (t * Terminal) wipeNTilesAt(tiles int, row int, col int){
	t.moveTo(row,col)
	println(RESET)
	for tiles > 0 {
		fmt.Printf(" ")
		tiles --
	}
}

func (t * Terminal) printRender(message string,txtLen int){
	t.Col += txtLen
	print(message)
	t.moveCursor(txtLen,LEFT)
}

func (t * Terminal) trimAndUpdateString(style * Context, text string) string{
	over := len(text) + t.Col - (t.Width - 1)
	if over > 0{
		text = text[0:len(text) - over]
	}
	for i := 0;i<len(text);i++{
		t.updateAtPos(t.Row,t.Col + i,&Recorded{
			Format: style,
			data:   text[i],
			code: text[i],
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
	character[0] = t.DataHistory[row][col][t.Depth - 1].data
	fmt.Printf(t.DataHistory[row][col][t.Depth - 1].Format.Format,character)
	t.moveCursor(1,LEFT)

}



func initContext () * Context {
	return &Context{Format: "\033["}
}

func (ctx * Context) addSimpleStyle(styleConst int) * Context{
	ctx.Format = ctx.Format + fmt.Sprintf("%d;",styleConst)
	return ctx
}

func (ctx * Context) addRgbStyleFg(r int,g int,b int) * Context{
	ctx.Format = ctx.Format + fmt.Sprintf("38;2;%d;%d;%d;",r,g,b)
	return ctx
}

func (ctx * Context) addRgbStyleBg(r int, g int,b int) * Context{
	ctx.Format = ctx.Format + fmt.Sprintf("48;2;%d;%d;%d;",r,g,b)
	return ctx
}

func (ctx * Context) finish() * Context {
	ctx.Format = ctx.Format[0:len(ctx.Format) - 1] + "m%s\033[0m"
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

func (t * Terminal) undoConditional(row int,col int,match byte){
	if t.DataHistory[row][col][t.Depth - 1].code == match{
		t.undoAtPos(row,col)
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
		character[0] = format.data
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

func (t * Terminal) placeCharFormat(char byte,row int,col int,format * Context,code byte){
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
		data:   char,
		code:   code,
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

func (t * Terminal) sendPlaceCharAtCoordCondUndo(char byte,row int,col int,lastRow int,lastCol int,match byte) {
	t.CustomFeed <- func(term *Terminal) {
		if format, ok := term.Associations[char]; ok {
			term.undoConditional(lastRow,lastCol,match)
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
func (t * Terminal) sendUndoAtLocationConditional(row int,col int,match byte){
	t.CustomFeed <- func(term *Terminal) {
		term.undoConditional(row,col,match)
	}
}

func (t * Terminal) assoc(char byte,format * Context,txt byte){
	t.sendCharAssociation(char,&Recorded{
		Format: format,
		data:   txt,
		code: char,
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
