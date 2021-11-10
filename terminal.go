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


type ContextMessage struct {
	Format * Context
	Body string
	Row int
	Col int
	IsCoord bool
}

type Recorded struct {
	Format * Context
	data byte
}

type DataAt struct {
	Pos * Coord
	Char byte
	IsCoord bool
}

type Terminal struct {
	Row int
	Col int
	Height int
	Width int
	Feed chan ContextMessage
	CharFeed chan DataAt
	CustomFeed chan func(terminal * Terminal)
	Associations map[byte] * Recorded
	StoredData [][] * Recorded
	CurrentData [][] * Recorded
}

func createTerminal(height int,width int)*Terminal{
	stored := make([][] * Recorded,height)
	current := make([][] * Recorded,height)
	for i:=0;i<height;i++{
		sRow := make([] * Recorded,width)
		cRow := make([] * Recorded,width)
		stored[i] = sRow
		current[i] = cRow
	}
	terminal := &Terminal{
		Row:  height,
		Col:  0,
		Feed: make(chan ContextMessage,MAX_MESSAGES),
		CharFeed: make(chan DataAt,MAX_MESSAGES),
		CustomFeed: make(chan func(terminal * Terminal),MAX_MESSAGES),
		Associations: make(map[byte]*Recorded),
		StoredData: stored,
		CurrentData: current,
		Height:height,
		Width:width,
	}
	for i := 0;i<height;i++ {
		for b := 0;b<width;b++ {
			print(" ")
		}
		println()
	}
	go terminal.handleRenders()
	return terminal
}

func (t * Terminal) send (context * Context,body string,row int,col int,isCoord bool){
	t.Feed <- ContextMessage{
		Format: context,
		Body:   body,
		Row:    row,
		Col:    col,
		IsCoord: isCoord,
	}
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
		})
	}
	return text
}

func (t * Terminal) placeAt(message string, row int, col int,txtLen int){
	t.moveTo(row,col)
	t.printRender(message,txtLen)

}

func (t * Terminal) updateAtPos(row int,col int,record * Recorded){
	t.StoredData[row][col] = t.CurrentData[row][col]
	t.CurrentData[row][col] = record
}

func (t * Terminal) undoAtPos(row int,col int){
	t.CurrentData[row][col] = t.StoredData[row][col]
	t.moveTo(row,col)
	if t.Col >= width - 1{
		return
	}
	t.Col ++
	var character [1] byte
	character[0] = t.CurrentData[row][col].data
	fmt.Printf(t.CurrentData[row][col].Format.Format,character)
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
	if t.CurrentData[row][col].data == match{
		t.undoAtPos(row,col)
	}
}

/*
Composition functions
 */

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

//sends a function that moves over and up/down some amount and prints text with a style when called
func (t * Terminal) sendPrintStyleAtShift(style * Context,rShift int,cShift int,text string){
	t.CustomFeed <- func(term *Terminal) {
		term.moveTo(term.Row + rShift,term.Col + cShift)
		term.writeStyleHere(style,text)
	}
}

//sends a function that when called places a key character at a coordinate
func (t * Terminal) sendPlaceCharAtCoord(char byte,row int,col int) {
	t.CustomFeed <- func(term *Terminal) {
		if format, ok := term.Associations[char]; ok {
			term.moveTo(row,col)
			if term.Col >= width - 1{
				return
			}
			term.Col ++
			var character [1] byte
			character[0] = char
			fmt.Printf(format.Format.Format,character)
			term.updateAtPos(row,col,&Recorded{
				Format: format.Format,
				data:   char,
			})
			term.moveCursor(1,LEFT)
		}
	}
}
//bake undo into send messages
//sends a function that when called places a key character at a over, up/down shifted location
func (t * Terminal) sendPlaceCharAtShift(char byte,rShift int,cShift int) {
	t.CustomFeed <- func(term *Terminal) {
		if format, ok := term.Associations[char]; ok {
			term.moveTo(term.Row + rShift,term.Col + cShift)
			if term.Col >= width - 1{
				return
			}
			term.Col ++
			var character [1] byte
			character[0] = char
			fmt.Printf(format.Format.Format,character)
			term.moveCursor(1,LEFT)
		}
	}
}

func (t * Terminal) sendPlaceCharAtShiftWithCondUndo(char byte,rShift int,cShift int,match byte) {
	t.CustomFeed <- func(term *Terminal) {
		if format, ok := term.Associations[char]; ok {
			term.undoConditional(term.Row,term.Col,match)
			term.moveTo(term.Row + rShift,term.Col + cShift)
			if term.Col >= width - 1{
				return
			}
			term.Col ++
			var character [1] byte
			character[0] = char
			fmt.Printf(format.Format.Format,character)
			term.moveCursor(1,LEFT)

		}
	}
}

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
			term.updateAtPos(row,col,&Recorded{
				Format: format.Format,
				data:   char,
			})
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

//possibility of register happening after first character is sent
func (t * Terminal) handleRenders(){
	var custom func (t * Terminal)
	for true{
	 custom = <- t.CustomFeed//could send normally, let's make this performance critical though
		custom(t)
	}
}
