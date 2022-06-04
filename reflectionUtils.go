package main

import "reflect"

func directIfPointer (i interface{}) reflect.Value {
	valueOf := reflect.ValueOf(i)
	for valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}
	return valueOf
}
