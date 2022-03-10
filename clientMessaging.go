package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

var GLOBAL_ID int = -1
var LOCAL_ID int = 0

/**
Send key for state and new value with port/ID whenever state updates

Do this with reflect for now and if it is too slow after profiling do some code generation before compile

Standard game will track local state, other player's state, and world state

Possible message can effect any of these, other players will be done by:
-Get message
-If id is for other players, lookup in int->state map
-Apply state change to that state and run correct handler

-If id is for own self, run update as well, may keep track of unclosed messages to close them and apply, or do nothing if success

-If id is for game state, run update on game state, do this by id, like -1?

Also allow local state updates to come through, like position in a menu

 */

type CoordExample struct {
	X int
	Y int
}

type StateExample struct {
	Name string
	Loc CoordExample
	LocPointer * CoordExample
}

type UpdateMessage struct {
	Id int
	Keys []string
	Value string
}

func marshal(anything interface{}) []byte {
	output,_ := json.Marshal(anything)
	return output
}

func toStateUpdate(state interface{}, id int, keys ...string) * UpdateMessage {
	value := "{"
	for i,key := range keys {
		toJson := string(marshal(reflect.ValueOf(state).FieldByName(key).Interface()))
		value+="\""+key+"\":"+toJson
		if i < len(keys) - 1 {
			value += ","
		}
	}
	value+="}"
	return &UpdateMessage{
		Id:    id,
		Value: value,
		Keys: keys,
	}
}

func (u * UpdateMessage) toBytes() []byte{
	output,_ := json.Marshal(u)
	return output
}

func messageFromBytes (bytes []byte) * UpdateMessage {
	var update UpdateMessage
	_ = json.Unmarshal(bytes,&update)
	return &update

}

func updateStateFromMessage(state interface{},message * UpdateMessage) {
	_ = json.Unmarshal([]byte(message.Value),&state)
}

func (u * UpdateMessage) updateProperState(localState interface{},playerStates map[int]interface{},globalState interface{},localId int,globalId int) {
	switch u.Id {
	case localId:
		updateStateFromMessage(localState,u)
		break
	case globalId:
		updateStateFromMessage(globalState,u)
		break
	default:
		updateStateFromMessage(playerStates[u.Id],u)
	}
}

func (u * UpdateMessage) applyToStates(localState interface{},playerStates map[int]interface{},globalState interface{},localHandlers map[string]func(),playersHandlers map[string]func(int),globalHandlers map[string]func()){
	u.updateProperState(localState,playerStates,globalState,LOCAL_ID,GLOBAL_ID)
	for _,key := range u.Keys {
		switch u.Id {
		case LOCAL_ID:
			localHandlers[key]()
			break
		case GLOBAL_ID:
			globalHandlers[key]()
			break
		default:
			playersHandlers[key](u.Id)
		}
	}
}

func main()  {
	state := StateExample{
		Name: "Aaron",
		Loc:  CoordExample{
			X: 1,
			Y: 2,
		},
		LocPointer: &CoordExample{
			X: 8,
			Y: 9,
		},
	}
	localState := StateExample{
		Name: "Ronnie",
		Loc:  CoordExample{
			X: 2,
			Y: 3,
		},
		LocPointer: &CoordExample{
			X: 10,
			Y: 11,
		},
	}
	start := time.Now()
	update := toStateUpdate(state,0,"Name")
	packet := update.toBytes()
	received := messageFromBytes(packet)
	updateStateFromMessage(&localState,received)
	fmt.Println(time.Now().Sub(start))
	fmt.Println(localState)
	fmt.Println(localState.LocPointer)
}

