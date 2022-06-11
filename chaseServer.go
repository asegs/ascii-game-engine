package main

import (
	"fmt"
	"time"
)

//Intentionally blank.
func doNotSavePlayerState (id int) {}

type PlayerState struct {
	Pos * Coord
}

type GlobalState struct {
	Pos * Coord
}

const mapWidth int = 20
const mapHeight int = 20

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

func serve () {
	serverConfig,err := loadServerConfig("configs/server_network_setting.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	playerPositions := make(map[int] interface{})
	globalState := GlobalState{Pos: &Coord{
		Row: 0,
		Col: 0,
	}}
	server := newServerDefault(func(id int) {
		playerPositions[id] = &PlayerState{Pos: &Coord{
			Row: 0,
			Col: 0,
		}}
	},doNotSavePlayerState,serverConfig,globalState,playerPositions)
	handlers := server.newZoneHandlers("map")
	handlers.addPlayerHandler(MOVE_LEFT, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		currentPos.Pos.Col = subtractOrCap(currentPos.Pos.Col)
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})
	handlers.addPlayerHandler(MOVE_RIGHT, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		currentPos.Pos.Col = addOrCap(currentPos.Pos.Col,mapWidth)
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})
	handlers.addPlayerHandler(MOVE_UP, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		currentPos.Pos.Row = subtractOrCap(currentPos.Pos.Row)
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})
	handlers.addPlayerHandler(MOVE_DOWN, func(id int) {
		currentPos := directCastToState(playerPositions[id])
		currentPos.Pos.Row = addOrCap(currentPos.Pos.Row, mapHeight)
		server.broadcastStateUpdate(playerPositions[id],id,true,"Pos")
	})

	server.start()
	for true {
		time.Sleep(1 * time.Second)
	}

}
