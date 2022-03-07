package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

/**
Send key for state and new value with port/ID whenever state updates

Do this with reflect for now and if it is too slow after profiling do some code generation before compile
 */

func toStateUpdate(state interface{}, key string, id int) string {
	field := reflect.ValueOf(state).FieldByName(key)
	toJson,_ := json.Marshal(field.Interface())
	return fmt.Sprintf("%d,%s,%v",id,key,string(toJson))
}


