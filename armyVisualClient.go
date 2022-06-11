package main

import (
	"encoding/json"
	"fmt"
)

type FireWrapper struct {
	Fire FirePacket
}

func armyVisual () {
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
	gameMap := make([][] int, mapHeight)
	for i := 0 ; i < mapHeight ; i ++ {
		gameMap[i] = make([] int, mapWidth)
	}
	globalState := &GlobalState{Pos: &Coord{
		Row: 0,
		Col: 0,
	},
	Map: gameMap,
	}
	onConnect := func(id int) {
		playerStates[id] = &PlayerState{Pos: &Coord{
			Row: 0,
			Col: 0,
		}}
	}

	visualClient, input := isometricClientWithInput("Army Men",40,20)
	visualClient.addBgSprite('0', "assets/sprites/grass.png")
	visualClient.addFgSprite('*',"assets/sprites/firefighter.png")
	visualClient.addFgSprite('Q',"assets/sprites/fire.png")
	visualClient.addFgSprite('R',"assets/sprites/redking.png")
	visualClient.addFgSprite('B',"assets/sprites/blueking.png")
	window := createClientWindow(mapHeight, mapWidth, &TilePair{
		ShownSymbol:    ' ',
		BackgroundCode: '0',
	},visualClient)
	zoning := initZones(mapHeight,mapWidth,input,visualClient)
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
	client.addGlobalHandler("Map", func(oldState interface{}) {
		oldMap := oldState.(* GlobalState).Map
		for row := 0 ; row < mapHeight ; row ++ {
			for col := 0 ; col < mapWidth ; col ++ {
				if oldMap[row][col] != globalState.Map[row][col] {
					window.placeFgCharAtCoord(byte(globalState.Map[row][col]),row,col)
				}
			}
		}
	})
	client.addCustomHandler("Fire", func(s string) {
		var firePacket FireWrapper
		er := json.Unmarshal([]byte(s), &firePacket)
		if er != nil {
			LogString(er.Error())
			return
		}
		if firePacket.Fire.OnFire {
			window.placeFgCharAtCoord('Q',firePacket.Fire.Pos.Row,firePacket.Fire.Pos.Col)
		}else {
			window.placeFgCharAtCoord('0',firePacket.Fire.Pos.Row,firePacket.Fire.Pos.Col)
		}
	})

	client.listen()
	visualClient.show()

}
