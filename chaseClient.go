package main

import (
	"fmt"
	"time"
)

func render () {
	go HandleLog()
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
	cursor := initContext().addRgbStyleFg(255,0,0).compile()
	redBlock := initContext().addRgbStyleBg(255,0,0).compile()
	blackBlock := initContext().addRgbStyleBg(0,0,0).compile()
	greenBlock := initContext().addRgbStyleBg(0,255,0).compile()
	blueBlock := initContext().addRgbStyleBg(0,0,255).compile()
	hunter := initContext().addRgbStyleFg(255,255,255).compile()
	clear := initContext().addSimpleStyle(0).compile()
	terminal := createTerminal(mapHeight, mapWidth, &Recorded{
		Format:         clear,
		ShownSymbol:    ' ',
		BackgroundCode: '0',
	})
	terminal.assoc('0',clear,' ')
	terminal.assoc('1',blackBlock,' ')
	terminal.assoc('2',greenBlock,' ')
	terminal.assoc('3',blueBlock,' ')
	terminal.assoc('*',cursor,'*')
	terminal.assoc('x',redBlock,' ')
	terminal.assoc('?',hunter,'?')
	zoning := initZones(mapHeight,mapWidth,input,terminal)
	zone,err := zoning.createZone(0,0,mapHeight,mapWidth,true)
	if err != nil {
		fmt.Println("creating map error: " + err.Error())
		return
	}
	_ = zoning.cursorEnterZone(zone,0)

	client := newClient([]byte{192,168,0,225},input,localState,playerStates,globalState,onConnect,clientConfig)
	client.addLocalHandler("Pos", func() {
		zone.sendPlaceCharFormat('*',localState.Pos.Row,localState.Pos.Col,cursor,'*')
	})
	client.addPlayersHandler("Pos", func(id int) {
		pos := playerStates[id].(PlayerState).Pos
		zone.sendPlaceCharFormat('*',pos.Row,pos.Col,cursor,'*')
	})

	client.listen()
	for true {
		time.Sleep(1 * time.Second)
	}

}