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
}

type UpdateMessage struct {
	Id int
	Key string
	Value interface{}
}

func toStateUpdate(state interface{}, key string, id int) * UpdateMessage {
	field := reflect.ValueOf(state).FieldByName(key)
	return &UpdateMessage{
		Id:    id,
		Key:   key,
		Value: field.Interface(),
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
	reflect.Indirect(reflect.ValueOf(state)).FieldByName(message.Key).Set(reflect.ValueOf(message.Value))
}

func main()  {
	state := StateExample{
		Name: "Aaron",
		Loc:  CoordExample{
			X: 1,
			Y: 2,
		},
	}
	localState := StateExample{
		Name: "Ronnie",
		Loc:  CoordExample{
			X: 2,
			Y: 3,
		},
	}
	start := time.Now()
	update := toStateUpdate(state,"Name",0)
	packet := update.toBytes()
	received := messageFromBytes(packet)
	updateStateFromMessage(&localState,received)
	fmt.Println(time.Now().Sub(start))
}

