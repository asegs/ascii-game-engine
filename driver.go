package main
//
//import (
//	"fmt"
//	"time"
//)
//
//const height int = 40
//const width int = 100
//func (terminal * Terminal) erasePath(p []*Coord){
//	for i := 1;i<len(p) - 1;i++{
//		terminal.sendUndoAtLocationConditional(p[i].Row,p[i].Col,'x')
//	}
//}
//func (terminal * Terminal) drawPath(p []*Coord){
//	for i := 1;i<len(p) - 1;i++{
//		terminal.sendPlaceCharAtCoord('x',p[i].Row,p[i].Col)
//	}
//}
//
//
//func composeNewContext (bg * Context,fg * Context) * Context {
//	newCtx := bg.copyContext()
//	newCtx.removeRgbStyle(true)
//	newCtx.addStyleRaw(fg.getColorInfo(true))
//	newCtx.compile()
//	return newCtx
//}
//
//func (terminal * Terminal) drawFgOverBg(row int, col int, cursor *Context, oldX int, oldY int) {
//	oldStyle := terminal.DataHistory[row][col][terminal.Depth - 1].Format
//	composedStyle := composeNewContext(oldStyle,cursor)
//	terminal.sendPlaceCharFormat('*',row,col,composedStyle,'*')
//	terminal.sendUndoAtLocationConditional(oldY,oldX,'*')
//}
//
//
//func main(){
//	input := initializeInput()
//	cursor := initContext().addRgbStyleFg(255,0,0).compile()
//	redBlock := initContext().addRgbStyleBg(255,0,0).compile()
//	blackBlock := initContext().addRgbStyleBg(0,0,0).compile()
//	greenBlock := initContext().addRgbStyleBg(0,255,0).compile()
//	blueBlock := initContext().addRgbStyleBg(0,0,255).compile()
//	hunter := initContext().addRgbStyleFg(255,255,255).compile()
//	clear := initContext().addSimpleStyle(0).compile()
//	terminal := createTerminal(height,width,&Recorded{
//		Format: clear,
//		data: ' ',
//		code: '0',
//	},4)
//	zoning := initZones(height,width,input)
//	mapZone,err := zoning.createZone(0,0,height,width - 20,true)
//	if err != nil {
//		fmt.Println("creating map error: " + err.Error())
//		return
//	}
//	err = zoning.cursorEnterZone(mapZone)
//	if err != nil {
//		fmt.Println("error entering zone: " + err.Error())
//	}
//	faceZone,err := zoning.createZone(0,width - 20,10,20,false)
//	if err != nil {
//		fmt.Println("creating faces error: " + err.Error())
//		return
//	}
//	faces := make([] string,2)
//	faces[0] = "assets/faces/smile_face.txt"
//	faces[1] = "assets/faces/open_mouth.txt"
//	exps := make([] string,2)
//	exps[0] = "smile"
//	exps[1] = "open"
//	face := buildFace(exps,faces,"guy")
//	go terminal.cycleExpressions(face,exps,600,-1,faceZone)
//	var dir byte
//	var path []*Coord
//	path = nil
//	terminal.assoc('0',clear,' ')
//	terminal.assoc('1',blackBlock,' ')
//	terminal.assoc('2',greenBlock,' ')
//	terminal.assoc('3',blueBlock,' ')
//	terminal.assoc('*',cursor,'*')
//	terminal.assoc('x',redBlock,' ')
//	terminal.assoc('?',hunter,'?')
//
//	//go follower(terminal)
//	polo := make(chan * Coord,1000)
//	go marcoPolo(terminal,polo,'0')
//	for {
//		dir = <- mapZone.Events
//		realX,realY := mapZone.getRealCoords()
//		if 128 <= dir && dir <= 131 {
//			accepted := zoning.moveInDirection(dir)
//			if accepted {
//				newX,newY := mapZone.getRealCoords()
//				terminal.drawFgOverBg(newY, newX, cursor, realX, realY)
//			}
//			continue
//		}
//		switch dir {
//		case '1':
//			if terminal.DataHistory[realY][realX][terminal.Depth - 2].code == '1'{
//				terminal.sendPlaceCharAtCoord('0',realY,realX)
//			}else{
//				terminal.sendPlaceCharAtCoord('1',realY,realX)
//			}
//			break
//		case '2':
//			if terminal.DataHistory[realY][realX][terminal.Depth - 2].code == '2'{
//				terminal.sendPlaceCharAtCoord('0',realY,realX)
//			}else{
//				terminal.sendPlaceCharAtCoord('2',realY,realX)
//			}
//			break
//		case '3':
//			if terminal.DataHistory[realY][realX][terminal.Depth - 2].code == '3'{
//				terminal.sendPlaceCharAtCoord('0',realY,realX)
//			}else{
//				terminal.sendPlaceCharAtCoord('3',realY,realX)
//			}
//			break
//		case ENTER:
//			if path != nil{
//				terminal.erasePath(path)
//				path = nil
//			}
//			maze,start,end := terminal.parseMazeFromCurrent('1','0','2','3')
//			path = astar(maze,start,end)
//			terminal.drawPath(path)
//			break
//		case BACKSLASH:
//			if path != nil{
//				terminal.erasePath(path)
//				path = nil
//			}
//			break
//		case BACKSPACE:
//			for i := mapZone.Y;i<mapZone.Y + mapZone.Height;i++{
//				for b := mapZone.X;b<mapZone.X + mapZone.Width;b++{
//					terminal.sendPlaceCharAtCoord('0',i,b)
//				}
//			}
//			break
//		case 'e':
//			polo <- &Coord{
//				Row: realY,
//				Col: realX,
//			}
//		}
//	}
//}
//
//
//
//func marcoPolo (t * Terminal,polo chan * Coord,validTerritory byte) {
//	row := 0
//	col := 0
//	path := make([] * Coord,0)
//	target := &Coord{}
//	target = nil
//	newTarget := false
//	for true {
//		if len(polo) > 0 {
//			target = <- polo
//			newTarget = true
//		}
//
//		if target == nil {
//			time.Sleep(250 * time.Millisecond)
//			continue
//		}
//		//check if moving onto invalid
//		if len(path) >= 1 {
//			contentAtMove := t.DataHistory[path[0].Row][path[0].Col][t.Depth - 1]
//			if contentAtMove.code != validTerritory || newTarget{
//				if newTarget {
//					newTarget = false
//				}
//				maze,_,_ := t.parseMazeFromCurrent('1','0','2','3')
//				path = astar(maze,&Coord{
//					Row: row,
//					Col: col,
//				},&Coord{
//					Row: target.Row,
//					Col: target.Col,
//				})
//				if len(path) > 1 {
//					path = path[1:]
//				}
//			}
//		}
//		if len(path) >= 2 {
//			t.sendPlaceCharAtCoordCondUndo('?',path[0].Row,path[0].Col,row,col,'?')
//			row = path[0].Row
//			col = path[0].Col
//			path = path[1:]
//		}else {
//			maze,_,_ := t.parseMazeFromCurrent('1','0','2','3')
//			path = astar(maze,&Coord{
//				Row: row,
//				Col: col,
//			},&Coord{
//				Row: target.Row,
//				Col: target.Col,
//			})
//			if len(path) > 1 {
//				path = path[1:]
//			}
//		}
//		time.Sleep(250 * time.Millisecond)
//
//
//	}
//}
//
//
//func follower (t * Terminal) {
//	row := 0
//	col := 0
//	path := make([] * Coord,0)
//	for true {
//		maze,_,_ := t.parseMazeFromCurrent('1','0','2','3')
//		target := t.getCoordsForCursor('*')
//		if target == nil {
//			time.Sleep(250 * time.Millisecond)
//			continue
//		}
//		path = astar(maze,&Coord{
//			Row: row,
//			Col: col,
//		},&Coord{
//			Row: target.Row,
//			Col: target.Col,
//		})
//		if len(path) > 2 {
//			t.sendPlaceCharAtCoordCondUndo('?',path[1].Row,path[1].Col,row,col,'?')
//			row = path[1].Row
//			col = path[1].Col
//		}
//		time.Sleep(250 * time.Millisecond)
//
//
//	}
//}