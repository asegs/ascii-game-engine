package main

import "fmt"

const height int = 40
const width int = 100

func composeNewContext (bg * Context,fg * Context) * Context {
	newCtx := bg.copyContext()
	newCtx.removeRgbStyle(true)
	newCtx.addStyleRaw(fg.getColorInfo(true))
	newCtx.compile()
	return newCtx
}

func (terminal * Terminal) drawFgOverBg(row int, col int, cursor *Context, oldX int, oldY int) {
	oldStyle := terminal.DataHistory[row][col][terminal.Depth - 1].Format
	composedStyle := composeNewContext(oldStyle,cursor)
	terminal.sendPlaceCharFormat('*',row,col,composedStyle,'*')
	terminal.sendUndoAtLocationConditional(oldY,oldX,'*')
}

func main () {
	input := initializeInput()
	network,err := initNetwork(10001,input)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cursor := initContext().addRgbStyleFg(255,0,0).compile()
	redBlock := initContext().addRgbStyleBg(255,0,0).compile()
	blackBlock := initContext().addRgbStyleBg(0,0,0).compile()
	greenBlock := initContext().addRgbStyleBg(0,255,0).compile()
	blueBlock := initContext().addRgbStyleBg(0,0,255).compile()
	hunter := initContext().addRgbStyleFg(255,255,255).compile()
	clear := initContext().addSimpleStyle(0).compile()
	terminal := createTerminal(height,width,&Recorded{
		Format: clear,
		data: ' ',
		code: '0',
	},4)
	zoning := initZones(height,width,input)
	mapZone,err := zoning.createZone(0,0,height,width - 20,true)
	if err != nil {
		fmt.Println("creating map error: " + err.Error())
		return
	}
	err = zoning.cursorEnterZone(mapZone,0)
	if err != nil {
		fmt.Println("error entering zone: " + err.Error())
	}
	faceZone,err := zoning.createZone(0,width - 20,10,20,false)
	if err != nil {
		fmt.Println("creating faces error: " + err.Error())
		return
	}
	faces := make([] string,2)
	faces[0] = "assets/faces/smile_face.txt"
	faces[1] = "assets/faces/open_mouth.txt"
	exps := make([] string,2)
	exps[0] = "smile"
	exps[1] = "open"
	face := buildFace(exps,faces,"guy")
	go terminal.cycleExpressions(face,exps,600,-1,faceZone)
	var path []*Coord
	path = nil
	terminal.assoc('0',clear,' ')
	terminal.assoc('1',blackBlock,' ')
	terminal.assoc('2',greenBlock,' ')
	terminal.assoc('3',blueBlock,' ')
	terminal.assoc('*',cursor,'*')
	terminal.assoc('x',redBlock,' ')
	terminal.assoc('?',hunter,'?')

	var dir * NetworkedMsg
	for {
		dir = <- mapZone.Events
		if dir.From == LOCAL_PORT {
			network.broadcast(dir.Msg)
		}
	}
}