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

func (t * Terminal) placeAt(message string, row int, col int,txtLen int){
	t.moveTo(row,col)
	t.printRender(message,txtLen)

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
	t.placeAt(fmt.Sprintf(style.Format,text),row,col,len(text))
}

func (t * Terminal) writeStyleHere(style * Context,text string){
	t.printRender(fmt.Sprintf(style.Format,text),len(text))
}


func (t * Terminal) composeCharAssociation(char byte,recorded * Recorded) func(t * Terminal){
	return func(t *Terminal) {
		t.Associations[char] = recorded
	}
}


//possibility of register happening after first character is sent
func (t * Terminal) handleRenders(){
	var ctx ContextMessage
	var cstm func (t * Terminal)
	var char DataAt
	var character [1]byte
	for true{
		select {
			case ctx = <- t.Feed:
				ctx = <- t.Feed
				if ctx.IsCoord {
					t.moveTo(ctx.Row,ctx.Col)
				}else{
					t.moveTo(t.Row + ctx.Row,t.Col + ctx.Col)
				}

				t.writeStyleHere(ctx.Format,ctx.Body)

			case cstm = <- t.CustomFeed://could send normally, let's make this performance critical though
				cstm(t)
			case char = <- t.CharFeed:
				if format, ok := t.Associations[char.Char]; ok {
					if char.IsCoord{
						t.moveTo(char.Pos.Row,char.Pos.Col)
					}else{
						t.moveTo(t.Row + char.Pos.Row,t.Col + char.Pos.Col)
					}
					if t.Col >= width - 1{
						continue
					}
					t.Col ++
					character[0] = char.Char
					fmt.Printf(format.Format.Format,character)
					t.moveCursor(1,LEFT)

				}
		}
	}
}
