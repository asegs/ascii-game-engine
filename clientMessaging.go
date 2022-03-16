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

type StatePair struct {
	Key string
	Json string
}

type StateExample struct {
	Name string
	Loc CoordExample
	LocPointer * CoordExample
}

type UpdateMessage struct {
	Id int
	Pairs [] StatePair
}

func marshal(anything interface{}) []byte {
	output,_ := json.Marshal(anything)
	return output
}

func newStateUpdate(id int) * UpdateMessage {
	return &UpdateMessage{
		Id:    id,
		Pairs: make([] StatePair,0),
	}
}

func (u * UpdateMessage) append(state interface{}, keys ...string) * UpdateMessage {
	for _,key := range keys {
		u.Pairs = append(u.Pairs,StatePair{
			Key:  key,
			Json: `{"` + key + `":` + string(marshal(reflect.ValueOf(state).FieldByName(key).Interface())) + "}",
		})
	}
	return u
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

func updateStateFromJson(state interface{},data string) {
	_ = json.Unmarshal([]byte(data),&state)
}

func keyInState (key string, state interface{}) bool{
	return reflect.ValueOf(state).FieldByName(key).IsValid()
}

func (p StatePair) performCustomFunction(customs map[string]func(string)) {
	customs[p.Key](p.Json)
}


func (u * UpdateMessage) applyToStates(localState interface{},playerStates map[int]interface{},globalState interface{},localHandlers map[string]func(),playersHandlers map[string]func(int),globalHandlers map[string]func(),customHandlers map[string]func(string2 string)){
	for _,pair := range u.Pairs {
		switch u.Id {
		case LOCAL_ID:
			if keyInState(pair.Key,localState) {
				updateStateFromJson(&localState,pair.Json)
				localHandlers[pair.Key]()
			} else {
				pair.performCustomFunction(customHandlers)
			}
			break
		case GLOBAL_ID:
			if keyInState(pair.Key,globalState) {
				updateStateFromJson(&globalState,pair.Json)
				globalHandlers[pair.Key]()
			} else {
				pair.performCustomFunction(customHandlers)
			}
			break
		default:
			playerState := playerStates[u.Id]
			if keyInState(pair.Key,playerState) {
				updateStateFromJson(&playerState,pair.Json)
				playersHandlers[pair.Key](u.Id)
			} else {
				pair.performCustomFunction(customHandlers)
			}
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
	update := newStateUpdate(0).append(state,"Name","Loc","LocPointer")
	packet := update.toBytes()
	received := messageFromBytes(packet)
	for _,pair := range received.Pairs {
		updateStateFromJson(&localState,pair.Json)
	}
	fmt.Println(time.Now().Sub(start))
	fmt.Println(localState)
	fmt.Println(localState.LocPointer)
}

