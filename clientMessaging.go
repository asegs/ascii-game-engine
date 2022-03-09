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

/**
A couple dummy structs for demonstrating how this works.
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

/**
Identifies the value of the field which has changed, and creates an update message containing that field.
 */
func toStateUpdate(state interface{}, key string, id int) * UpdateMessage {
	field := reflect.ValueOf(state).FieldByName(key)
	return &UpdateMessage{
		Id:    id,
		Key:   key,
		Value: field.Interface(),
	}
}

/**
Simply converts an UpdateMessage into a byte string.
 */
func (u * UpdateMessage) toBytes() []byte{
	output,_ := json.Marshal(u)
	return output
}

/**
Rebuilds an UpdateMessage from a byte string.
 */
func messageFromBytes (bytes []byte) * UpdateMessage {
	var update UpdateMessage
	_ = json.Unmarshal(bytes,&update)
	return &update

}

/**
Updates some state of an object from a given message.
 */
func updateStateFromMessage(state interface{},message * UpdateMessage) {
	//The interface value from the update message, the new value which the key has been changed to.
	field := reflect.Indirect(reflect.ValueOf(state)).FieldByName(message.Key)
	//If the updated field was a pointer
	if field.Kind() == reflect.Ptr {
		//Set the field to the actual value of that pointer
		field = reflect.Indirect(field)
	}
	//If the value was a struct, or if it was originally a pointer...
	if field.Kind() == reflect.Struct {
		//We do this because message.Value will be a map[string]interface{} if it is a struct.
		//Convert the message to a JSON string
		output, err := json.Marshal(message.Value)
		if err != nil {
			fmt.Println(err.Error())
		}
		//Instantiate a new object of the fields type
		newVersion := reflect.New(field.Type()).Interface()
		//Push the map of pairs into the new object for the proper fields
		err = json.Unmarshal(output, &newVersion)
		if err != nil {
			fmt.Println(err.Error())
		}
		//Sets the field equal to wherever the pointer for the field points.
		field.Set(reflect.Indirect(reflect.ValueOf(newVersion)))
	} else {
		//Sets the field directly to the message value.
		field.Set(reflect.ValueOf(message.Value))
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
	update := toStateUpdate(state,"LocPointer",0)
	packet := update.toBytes()
	received := messageFromBytes(packet)
	fmt.Println(localState.LocPointer)
	updateStateFromMessage(&localState,received)
	fmt.Println(time.Now().Sub(start))
	fmt.Println(state.LocPointer)
	fmt.Println(localState.LocPointer)
}

