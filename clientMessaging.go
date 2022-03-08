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

Fails with struct key, if struct -> to string, json unmarshal
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
	field := reflect.Indirect(reflect.ValueOf(state)).FieldByName(message.Key)
	fmt.Println(field.Kind())
	if field.Kind() == reflect.Struct {
		output,err := json.Marshal(message.Value)
		if err != nil {
			fmt.Println(err.Error())
		}
		newVersion := reflect.New(field.Type()).Interface()
		err = json.Unmarshal(output, &newVersion)
		if err != nil {
			fmt.Println(err.Error())
		}
		field.Set(reflect.Indirect(reflect.ValueOf(newVersion)))
	}else if field.Kind() == reflect.Ptr{
		//handle pointer case
	} else {
		field.Set(reflect.ValueOf(message.Value))
	}

	fmt.Println(field)
	fmt.Println(field.Type())
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
	update := toStateUpdate(state,"LocPointer",0)
	packet := update.toBytes()
	received := messageFromBytes(packet)
	updateStateFromMessage(&localState,received)
	fmt.Println(time.Now().Sub(start))
	fmt.Println(localState)
}

