package main

import (
	"fmt"
	"time"
)

const (
	None int = '0'
	Fire	= 'Q'
	RedKing	= 'R'
	BlueKing = 'B'
)

//Intentionally blank.
func doNotSavePlayerState (id int) {}

type PlayerState struct {
	Pos * Coord
}

type GlobalState struct {
	Pos * Coord
	Map [] [] int
}

type FirePacket struct {
	Pos * Coord
	OnFire bool
}

const mapWidth int = 40
const mapHeight int = 40

func subtractOrCap (val int) int {
	if val < 1 {
		return val
	}
	return val - 1
}

func addOrCap (val int, max int) int {
	if val >= max - 2{
		return val
	}
	return val + 1
}

func directCastToState (i interface{}) * PlayerState {
	return i.(* PlayerState)
}

func enteringFire (m [][] int, pos * Coord) bool {
	return m[pos.Row][pos.Col] == Fire
}

func serve () {
	serverConfig,err := loadServerConfig("configs/server_network_setting.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	playerPositions := make(map[int] interface{})
	gameMap := make([][] int, mapHeight)
	for i := 0 ; i < mapHeight ; i ++ {
		row := make([] int, mapWidth)
		for x := 0; x < mapWidth ; x ++ {
			row[x] = '0'
		}
		gameMap[i] = row

	}

	gameMap[0][0] = RedKing

	gameMap[37][37] = BlueKing

	globalState := GlobalState{Pos: &Coord{
		Row: 0,
		Col: 0,
	},
	Map: gameMap,
	}
	server := newServerDefault(func(id int) {
		playerPositions[id] = &PlayerState{Pos: &Coord{
			Row: 0,
			Col: 0,
		}}
	},doNotSavePlayerState,serverConfig,globalState,playerPositions)
	handlers := server.newZoneHandlers("map")
	handlers.addPlayerHandler(MOVE_LEFT, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		newCol := subtractOrCap(currentPos.Pos.Col)
		if !enteringFire(gameMap, &Coord{
			Row: currentPos.Pos.Row,
			Col: newCol,
		}) {
			currentPos.Pos.Col = newCol
		}
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})
	handlers.addPlayerHandler(MOVE_RIGHT, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		newCol := addOrCap(currentPos.Pos.Col,mapWidth)
		if !enteringFire(gameMap, &Coord{
			Row: currentPos.Pos.Row,
			Col: newCol,
		}) {
			currentPos.Pos.Col = newCol
		}
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})
	handlers.addPlayerHandler(MOVE_UP, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		newRow := subtractOrCap(currentPos.Pos.Row)
		if !enteringFire(gameMap, &Coord{
			Row: newRow,
			Col: currentPos.Pos.Col,
		}) {
			currentPos.Pos.Row = newRow
		}
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})
	handlers.addPlayerHandler(MOVE_DOWN, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		newRow := addOrCap(currentPos.Pos.Row, mapHeight)
		if !enteringFire(gameMap, &Coord{
			Row: newRow,
			Col: currentPos.Pos.Col,
		}) {
			currentPos.Pos.Row = newRow
		}
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})
	handlers.addPlayerHandler('Q', func(id int) {
		currentPos := directCastToState(playerPositions[id])
		if globalState.Map[currentPos.Pos.Row][currentPos.Pos.Col] == Fire {
			globalState.Map[currentPos.Pos.Row][currentPos.Pos.Col] = None
		} else {
			globalState.Map[currentPos.Pos.Row][currentPos.Pos.Col] = Fire
		}
		server.broadcastCustomPair("Fire", &FirePacket{
			Pos:    currentPos.Pos,
			OnFire: globalState.Map[currentPos.Pos.Row][currentPos.Pos.Col] == Fire,
		},GLOBAL_ID,true)
	})

	server.start()
	for true {
		time.Sleep(1 * time.Second)
	}

}
