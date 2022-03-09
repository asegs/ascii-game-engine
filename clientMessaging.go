package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

/**
Send key for state and new value with port/ID whenever state updates

Do this with reflect for now and if it is too slow after profiling do some code generation before compile

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

