package main

import (
	"fmt"
	"time"
)

func render () {
	clientConfig,err := loadClientConfig("configs/client_network_setting.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	localState := &PlayerState{Pos: &Coord{
		Row: 0,
		Col: 0,
	}}
	playerStates := make(map[int] interface{})
	globalState := &GlobalState{Pos: &Coord{
		Row: 0,
		Col: 0,
	}}
	onConnect := func(id int) {
		playerStates[id] = &PlayerState{Pos: &Coord{
			Row: 0,
			Col: 0,
		}}
	}

	input := initializeInput()
	terminalClient, input := terminalClientWithTerminalInput()
	window := createClientWindow(mapHeight, mapWidth, &TilePair{
		ShownSymbol:    ' ',
		BackgroundCode: '0',
	},terminalClient)
	terminalClient.MultiMapLookup.addForegroundColor('*',255,0,0)
	zoning := initZones(mapHeight,mapWidth,input,terminalClient)
	zone,err := zoning.createZone(0,0,mapHeight,mapWidth,true)
	if err != nil {
		fmt.Println("creating map error: " + err.Error())
		return
	}
	_ = zoning.cursorEnterZone(zone,0)
	disconnectHandler := func(id int) {
		disconnectedPos := playerStates[id].(* PlayerState).Pos
		window.sendUndoAtLocationConditional(disconnectedPos.Row,disconnectedPos.Col,'*',true)
		delete(playerStates,id)
	}
	client := newClient([]byte{127,0,0,1},&zone.Events,localState,playerStates,globalState,onConnect,disconnectHandler,clientConfig)
	client.addLocalHandler("Pos", func(oldState interface{}) {
		oldPos := oldState.(* PlayerState).Pos
		window.sendPlaceFgCharAtCoordCondUndo('*',localState.Pos.Row,localState.Pos.Col,oldPos.Row,oldPos.Col,'*',true)
	})
	client.addPlayersHandler("Pos", func(id int, oldState interface{}) {
		pos := playerStates[id].(* PlayerState).Pos
		oldPos := oldState.(* PlayerState).Pos
		window.sendPlaceFgCharAtCoordCondUndo('*',pos.Row,pos.Col,oldPos.Row,oldPos.Col,'*',true)
	})

	client.listen()
	for true {
		time.Sleep(1 * time.Second)
	}

}