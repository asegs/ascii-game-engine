package main

import "fmt"

func hoboVisual() {
	uninitialized := true
	clientConfig, err := loadClientConfig("configs/client_network_setting.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	localState := &PlayerState{Pos: &Coord{Row: 0, Col: 0}}
	playerStates := make(map[int]interface{})
	gameMap := make([][]byte, mapHeight)
	for i := 0; i < mapHeight; i++ {
		gameMap[i] = make([]byte, mapWidth)
	}
	globalState := &GlobalState{Grid: &gameMap}
	onConnect := func(id int) {
		playerStates[id] = &PlayerState{Pos: &Coord{Row: 0, Col: 0}}
	}

	visualClient, input := graphicalClientWithInput("Hobo Encampment", 40, 20, true)
	visualClient.addBgSprite('#', "assets/sprites/rough.png")
	visualClient.addBgSprite('&', "assets/sprites/medium.png")
	visualClient.addBgSprite('%', "assets/sprites/light.png")
	visualClient.addBgSprite('.', "assets/sprites/basic_trail.png")
	visualClient.addBgSprite(' ', "assets/sprites/water.png")
	visualClient.addFgSprite('*', "assets/sprites/firefighter.png")

	window := createClientWindow(mapHeight, mapWidth, ' ', ' ', visualClient)
	zoning := initZones(mapHeight, mapWidth, input, visualClient)
	zone, err := zoning.createZone(0, 0, mapHeight, mapWidth, true)
	if err != nil {
		fmt.Println("Creating map error " + err.Error())
		return
	}
	_ = zoning.cursorEnterZone(zone, 0)
	disconnectHandler := func(id int) {
		disconnectedPos := playerStates[id].(*PlayerState).Pos
		window.sendUndoAtLocationConditional(disconnectedPos.Row, disconnectedPos.Col, '*', true)
		delete(playerStates, id)
	}
	client := newClient([]byte{127, 0, 0, 1}, &zone.Events, localState, playerStates, globalState, onConnect, disconnectHandler, clientConfig)
	client.addPlayersHandler("Pos", func(id int, oldState interface{}) {
		pos := playerStates[id].(*PlayerState).Pos
		oldPos := oldState.(*PlayerState).Pos
		window.sendPlaceFgCharAtCoordCondUndo('*', pos.Row, pos.Col, oldPos.Row, oldPos.Col, '*')
	})
	client.addGlobalHandler("Grid", func(oldState interface{}) {
		oldMap := oldState.(*GlobalState).Grid
		for row := 0; row < mapHeight; row++ {
			for col := 0; col < mapWidth; col++ {
				if (*oldMap)[row][col] != (*globalState.Grid)[row][col] {
					window.placeBgCharAtCoord((*globalState.Grid)[row][col], row, col, uninitialized)
				}
			}
		}
		uninitialized = false
	})
	client.listen()
	visualClient.show()
}
