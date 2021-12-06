package main

import (
	"fmt"
	"strings"
)

const height int = 20
const width int = 40
func (terminal * Terminal) erasePath(p []*Coord){
	for i := 1;i<len(p) - 1;i++{
		terminal.sendUndoAtLocationConditional(p[i].Row,p[i].Col,'x')
	}
}
func (terminal * Terminal) drawPath(p []*Coord){
	for i := 1;i<len(p) - 1;i++{
		terminal.sendPlaceCharAtCoord('x',p[i].Row,p[i].Col)
	}
}

func getNthOccurrence (str string,substr string,n int) int{
	deleted := 0
	for i := 0; i < n ; i++ {
		occurrenceIdx := strings.Index(str,substr)
		if occurrenceIdx == -1 {
			return -1
		}
		deleted += occurrenceIdx + len(substr)
		str = str[occurrenceIdx + len(substr):]
		if i == n-1 {
			return deleted
		}
	}
	//if it ever gets here...you did something insane (indexed by 1 here, not 0)
	return -1
}

func composeNewContext (bg * Context,fg * Context) * Context {
	bgIdx := strings.Index(bg.Format,"[48")
	fgStartIdx := strings.Index(fg.Format,"[38;2;") + 1
	fgEndIdx := getNthOccurrence(fg.Format,";",5) + 1
	return &Context{Format: bg.Format[0:bgIdx] + fg.Format[fgStartIdx:fgEndIdx] + bg.Format[bgIdx:]}
}

func (terminal * Terminal) drawFgOverBg (row int,col int,zone * Zone,cursor * Context,oldX int,oldY int){
	trueX,trueY := zone.getRealNewCoords(row,col)
	oldStyle := terminal.DataHistory[trueY][trueX][terminal.Depth - 1].Format
	composedStyle := composeNewContext(oldStyle,cursor)
	terminal.sendPlaceCharFormat('*',row,col,composedStyle,'*')
	terminal.sendUndoAtLocationConditional(oldY,oldX,'*')
}


func main(){
	input := initializeInput()
	cursor := initContext().addRgbStyleFg(255,0,0).finish()
	redBlock := initContext().addRgbStyleBg(255,0,0).finish()
	blackBlock := initContext().addRgbStyleBg(0,0,0).finish()
	greenBlock := initContext().addRgbStyleBg(0,255,0).finish()
	blueBlock := initContext().addRgbStyleBg(0,0,255).finish()
	clear := initContext().addSimpleStyle(0).finish()
	terminal := createTerminal(height,width,&Recorded{
		Format: clear,
		data: ' ',
		code: '0',
	},4)
	zoning := initZones(height,width,input)
	mapZone,err := zoning.createZone(0,0,height,20,true)
	if err != nil {
		fmt.Println("creating map error: " + err.Error())
		return
	}
	err = zoning.cursorEnterZone(mapZone)
	if err != nil {
		fmt.Println("error entering zone: " + err.Error())
	}
	faceZone,err := zoning.createZone(0,20,10,20,false)
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
	var dir byte
	var path []*Coord
	path = nil
	terminal.assoc('0',clear,' ')
	terminal.assoc('1',blackBlock,' ')
	terminal.assoc('2',greenBlock,' ')
	terminal.assoc('3',blueBlock,' ')
	terminal.assoc('*',cursor,'*')
	terminal.assoc('x',redBlock,' ')
	for {
		dir = <- mapZone.Events
		realX,realY := mapZone.getRealCoords()
		if 128 <= dir && dir <= 131 {
			accepted := zoning.moveInDirection(dir)
			if accepted {
				newX,newY := mapZone.getRealCoords()
				terminal.drawFgOverBg(newY,newX,mapZone,cursor,realX,realY)
				//terminal.sendPlaceCharAtCoordCondUndo('*',newY,newX,realY,realX,'*')
			}
			continue
		}
		switch dir {
		case '1':
			if terminal.DataHistory[realY][realX][terminal.Depth - 2].code == '1'{
				terminal.sendPlaceCharAtCoord('0',realY,realX)
			}else{
				terminal.sendPlaceCharAtCoord('1',realY,realX)
			}
			break
		case '2':
			if terminal.DataHistory[realY][realX][terminal.Depth - 2].code == '2'{
				terminal.sendPlaceCharAtCoord('0',realY,realX)
			}else{
				terminal.sendPlaceCharAtCoord('2',realY,realX)
			}
			break
		case '3':
			if terminal.DataHistory[realY][realX][terminal.Depth - 2].code == '3'{
				terminal.sendPlaceCharAtCoord('0',realY,realX)
			}else{
				terminal.sendPlaceCharAtCoord('3',realY,realX)
			}
			break
		case ENTER:
			if path != nil{
				terminal.erasePath(path)
				path = nil
			}
			maze,start,end := terminal.parseMazeFromCurrent('1','0','2','3')
			path = astar(maze,start,end)
			terminal.drawPath(path)
			break
		case BACKSLASH:
			if path != nil{
				terminal.erasePath(path)
				path = nil
			}
			break
		case BACKSPACE:
			for i := mapZone.Y;i<mapZone.Y + mapZone.Height;i++{
				for b := mapZone.X;b<mapZone.X + mapZone.Width;b++{
					terminal.sendPlaceCharAtCoord('0',i,b)
				}
			}
			break
		}
	}
}
