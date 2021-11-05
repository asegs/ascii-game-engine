package main

import (
	"fmt"
)

type Direction rune
const (
	LEFT Direction = 'D'
	RIGHT Direction = 'C'
	DOWN Direction = 'B'
	UP Direction = 'A'
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
}

type Terminal struct {
	Row int
	Col int
	Height int
	Width int
	Feed chan ContextMessage
}

func createTerminal(height int,width int)*Terminal{
	feed := make(chan ContextMessage,MAX_MESSAGES)
	terminal := &Terminal{
		Row:  0,
		Col:  height + 1,
		Feed: feed,
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

func (t * Terminal) send (context * Context,body string,row int,col int){
	t.Feed <- ContextMessage{
		Format: context,
		Body:   body,
		Row:    row,
		Col:    col,
	}
}

func (t * Terminal) moveCursor(n int,dir Direction){
	fmt.Printf("\033[%d%c",n,dir)
	if dir == UP{

	}
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
	println(message)
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
func (t * Terminal) handleRenders(){
	var ctx ContextMessage
	for true{
		ctx = <- t.Feed
		t.writeStyleAt(ctx.Format,ctx.Body,ctx.Row,ctx.Col)
	}
}

func main(){
	t := createTerminal(10,40)
	c := initContext().addRgbStyleBg(255,0,0).finish()
	t.send(c,"Hello",4,4)
	for true{

	}
}
