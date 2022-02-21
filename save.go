package main

import (
	"encoding/json"
)

/**
If recorded is in provided list - record number in list, nothing more.
If not, consider recorded erratas and store in file as JSON/something.
*/

func recordedEquals (r1 * Recorded, r2 * Recorded) bool {
	return r1.Format.Format == r2.Format.Format &&
		r1.BackgroundCode == r2.BackgroundCode &&
		r1.ShownSymbol == r2.ShownSymbol
}

func recordedIndex (r * Recorded, styles [] * Recorded) int {
	for i, recorded := range styles {
		if recordedEquals(r,recorded) {
			return i
		}
	}
	return -1
}

func recordedToString(r * Recorded) string{
	output,err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(output)
}

func recordedFromString (record string) (error, * Recorded) {
	var r Recorded
	err := json.Unmarshal([]byte(record),&r)
	return err,&r
}


func (t * Terminal) saveState (filename string,styles [] * Recorded) error{

}

func (t * Terminal) loadState (filename string, styles [] * Recorded) error {

}

//func main()  {
//	mods := make([]string,2)
//	mods[0] = "one"
//	mods[1] = "two"
//	r := &Recorded{
//		Format:         &Context{
//			Format:    "test",
//			Modifiers: mods,
//		},
//		ShownSymbol:    'a',
//		BackgroundCode: 'x',
//	}
//	s := recordedToString(r)
//	_,r2 := recordedFromString(s)
//	fmt.Println(r2.ShownSymbol)
//}