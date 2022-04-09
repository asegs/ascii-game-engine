package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const height int = 40
const width int = 100
const LOAD bool = true

type DemoState struct {
	FollowerPos * Coord
	PlayerPos * Coord
	OtherPlayersPos [] * Coord
	Maze [][] byte
}

func (terminal * Terminal) erasePath(p []*Coord){
	for i := 1;i<len(p) - 1;i++{
		terminal.sendUndoAtLocationConditional(p[i].Row,p[i].Col,'x',true)
	}
}
func (terminal * Terminal) drawPath(p []*Coord){
	for i := 1;i<len(p) - 1;i++{
		terminal.sendPlaceCharAtCoord('x',p[i].Row,p[i].Col)
	}
}

func composeNewContext (bg * Context,fg * Context) * Context {
	newCtx := bg.copyContext()
	newCtx.removeRgbStyle(true)
	newCtx.addStyleRaw(fg.getColorInfo(true))
	newCtx.compile()
	return newCtx
}

func (terminal * Terminal) drawFgOverBg(row int, col int, cursor *Context, oldX int, oldY int) {
	oldStyle := terminal.DataHistory[row][col].top().Format
	composedStyle := composeNewContext(oldStyle,cursor)
	terminal.sendPlaceCharFormat('*',row,col,composedStyle,'*')
	terminal.sendUndoAtLocationConditional(oldY,oldX,'*',true)

	for _,row := range terminal.DataHistory {
		for _,col := range row {
			LogString(fmt.Sprintf("%c",col.top().BackgroundCode))
		}
	}
}

func runNetworked () {
	var state DemoState
	go HandleLog()
	err := loadConfig("main_settings.txt",ClientNetworkConfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
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
		ShownSymbol: ' ',
		BackgroundCode: '0',
	},8)
	zoning := initZones(height,width,input,terminal)
	mapZone,err := zoning.createZone(0,0,height,width - 30,true)
	if err != nil {
		fmt.Println("creating map error: " + err.Error())
		return
	}
	err = zoning.cursorEnterZone(mapZone,0)
	if err != nil {
		fmt.Println("error entering zone: " + err.Error())
	}
	zoning.setDefaultZone(mapZone)
	faceZone,err := zoning.createZone(0,width - 30,15,20,false)
	if err != nil {
		fmt.Println("creating faces error: " + err.Error())
		return
	}
	if LOAD {
		// Use a reflection something for this,idk doesn't matter right now
		//_,data := loadData("saves/data.json",state)
		//state = data.(DemoState)
		d,_ := ReadToString("saves/data.json")
		_ = json.Unmarshal([]byte(d),&state)
		_ = terminal.load("saves/state.json")
	}else {
		maze := make([][] byte,mapZone.Height)
		for i := 0 ; i < mapZone.Height ; i ++ {
			maze[i] = make([] byte, mapZone.Width)
		}
		state = DemoState{
			FollowerPos:     &Coord{
				Row: 0,
				Col: 0,
			},
			PlayerPos:       &Coord{
				Row: 0,
				Col: 0,
			},
			OtherPlayersPos: make([] * Coord,0),
			Maze:            maze,
		}
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
	go follower(terminal,&state)
	var dir * NetworkedMsg
	mapZone.CursorMap[LOCAL_PORT] = state.PlayerPos
	for {
		dir = <- mapZone.Events
		if dir.From == LOCAL_PORT {
			network.broadcast(dir.Msg)
		}
		if dir.Msg == TAB {
			err := zoning.cursorEnterZone(mapZone,dir.From)
			if err != nil {
				//do something
			}
			continue
		}

		realX,realY := mapZone.getRealCoords(dir.From)
		if MOVE_UP <= dir.Msg && dir.Msg <= MOVE_LEFT {
			accepted := zoning.moveInDirection(dir.Msg,dir.From)
			if accepted {
				newX,newY := mapZone.getRealCoords(dir.From)
				terminal.drawFgOverBg(newY, newX, cursor, realX, realY)
				state.PlayerPos.Row = newY
				state.PlayerPos.Col = newX
			}
			continue
		}
		switch dir.Msg {
		case '1':
			if terminal.DataHistory[realY][realX].back(1).BackgroundCode == '1' {
				terminal.sendPlaceCharAtCoord('0', realY, realX)
			} else {
				terminal.sendPlaceCharAtCoord('1', realY, realX)
			}
			break
		case '2':
			if terminal.DataHistory[realY][realX].back(1).BackgroundCode == '2' {
				terminal.sendPlaceCharAtCoord('0', realY, realX)
			} else {
				terminal.sendPlaceCharAtCoord('2', realY, realX)
			}
			break
		case '3':
			if terminal.DataHistory[realY][realX].back(1).BackgroundCode == '3' {
				terminal.sendPlaceCharAtCoord('0', realY, realX)
			} else {
				terminal.sendPlaceCharAtCoord('3', realY, realX)
			}
			break
		case ENTER:
			if path != nil {
				terminal.erasePath(path)
				path = nil
			}
			maze, start, end := terminal.parseMazeFromCurrent('1', '2', '3')
			path = astar(maze, start, end)
			terminal.drawPath(path)
			break
		case BACKSLASH:
			if path != nil {
				terminal.erasePath(path)
				path = nil
			}
			break
		case BACKSPACE:
			for i := mapZone.Y; i < mapZone.Y+mapZone.Height; i++ {
				for b := mapZone.X; b < mapZone.X+mapZone.Width; b++ {
					terminal.sendPlaceCharAtCoord('0', i, b)
				}
			}
			break
		case '=':
			err := terminal.save("saves/state.json")
			if err != nil {
				LogString(err.Error())
			}
			err = saveData("saves/data.json",state)
			if err != nil {
				LogString(err.Error())
			}
			break
		}
	}
}

func getClosestCoords(coords [] * Coord,me * Coord) * Coord{
	if len(coords) == 0 {
		return nil
	}
	closestDistance := pythagDistance(coords[0],me)
	closestIdx := 0
	for i := 1 ; i < len(coords) ; i ++ {
		newDist := pythagDistance(coords[i],me)
		if newDist < closestDistance {
			closestDistance = newDist
			closestIdx = i
		}
	}
	return coords[closestIdx]
}


func follower (t * Terminal,state * DemoState) {
	path := make([] * Coord,0)
	for true {
		maze,_,_ := t.parseMazeFromCurrent('1', '2', '3')
		targets := t.getCoordsForCursor('*')
		target := getClosestCoords(targets,state.FollowerPos)
		if target == nil {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		path = astar(maze,state.FollowerPos,&Coord{
			Row: target.Row,
			Col: target.Col,
		})
		if len(path) > 2 {
			t.sendPlaceCharAtCoordCondUndo('?',path[1].Row,path[1].Col,state.FollowerPos.Row,state.FollowerPos.Col,'?',true)
			state.FollowerPos.Row = path[1].Row
			state.FollowerPos.Col = path[1].Col
		}
		time.Sleep(250 * time.Millisecond)


	}
}