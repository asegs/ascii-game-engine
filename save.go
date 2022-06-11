package main
//
//import (
//	"encoding/json"
//	"fmt"
//)
//
//type State struct {
//	Data [][][]int
//}
//
//type Save struct {
//	History * State
//	Styles [] * Recorded
//}
//
///**
//If recorded is in provided list - record number in list, nothing more.
//If not, consider recorded erratas and store in file as JSON/something.
//*/
//
//func recordedEquals (r1 * Recorded, r2 * Recorded) bool {
//	return r1.Format.Format == r2.Format.Format &&
//		r1.BackgroundCode == r2.BackgroundCode &&
//		r1.ShownSymbol == r2.ShownSymbol
//}
//
//func recordedIndex (r * Recorded, styles [] * Recorded) int {
//	for i, recorded := range styles {
//		if recordedEquals(r,recorded) {
//			return i
//		}
//	}
//	return -1
//}
//
//func saveToString(s * Save) (error,string){
//	output,err := json.Marshal(s)
//	if err != nil {
//		return err,""
//	}
//	return nil,string(output)
//}
//
//func saveFromString (save string) (error, * Save) {
//	var s Save
//	err := json.Unmarshal([]byte(save),&s)
//	return err,&s
//}
//
//func (r * Recorded) toString() string{
//	return fmt.Sprintf("Foreground: %c, Background: %c, Text: %s",r.ShownSymbol,r.BackgroundCode,r.Format.Format)
//}
//
//func (t * Terminal) loadSave(save * Save) {
//	for y,row := range t.DataHistory {
//		for x,_ := range row {
//			records := make([] * Recorded,len(save.History.Data[y][x]))
//			for i := 0 ; i < len(records) ; i ++ {
//				records[i] = save.Styles[save.History.Data[y][x][i]]
//				t.DataHistory[y][x] = toHistory(records)
//			}
//		}
//	}
//}
//
//func (t * Terminal) drawInitialState(){
//	t.moveTo(0,0)
//	for y,row := range t.DataHistory {
//		for x, col := range row {
//			t.writeStyleAtNoHistory(col.top().Format,string(col.top().ShownSymbol),y,x)
//		}
//	}
//	t.moveTo(0,0)
//}
//
//func (t * Terminal) toState() (* State,[] * Recorded) {
//	styles := make([] * Recorded,0)
//	records := make([][][]int,t.Height)
//	for i := 0 ; i < t.Height ; i ++ {
//		records[i] = make([][]int,t.Width)
//	}
//	idx := -1
//	for y,row := range t.DataHistory {
//		for x,col := range row {
//			stackArr := col.toArr()
//			records[y][x] = make([] int, len(stackArr))
//			for i,record := range stackArr {
//				idx = recordedIndex(record,styles)
//				if idx == -1 {
//					idx = len(styles)
//					styles = append(styles,record)
//				}
//				records[y][x][i] = idx
//			}
//
//		}
//	}
//	return &State{Data: records} , styles
//}
//
//func (t * Terminal) save (filename string) error{
//	save := &Save{
//		History: nil,
//		Styles:  nil,
//	}
//	save.History,save.Styles = t.toState()
//	err,output := saveToString(save)
//	if err != nil {
//		return err
//	}
//	Write(filename,output)
//	return err
//}
//
//func (t * Terminal) load (filename string) error {
//	text,err := ReadToString(filename)
//	if err != nil {
//		return err
//	}
//	err,save := saveFromString(text)
//	if err != nil {
//		return err
//	}
//	t.loadSave(save)
//	t.drawInitialState()
//	t.parseMazeFromCurrent('1','2','3')
//	return nil
//}
//
//func saveData (filename string,gameState interface{})error{
//	output,err := json.Marshal(gameState)
//	if err != nil {
//		return err
//	}
//	Write(filename, string(output))
//	return err
//}
//
//func loadData(filename string, onto interface{}) (error,interface{}){
//	text,err := ReadToString(filename)
//	if err != nil {
//		return err,nil
//	}
//	err = json.Unmarshal([]byte(text),&onto)
//	return err,onto
//}
//
//
////func main()  {
////	mods := make([]string,2)
////	mods[0] = "one"
////	mods[1] = "two"
////	r := &Recorded{
////		Format:         &Context{
////			Format:    "test",
////			Modifiers: mods,
////		},
////		ShownSymbol:    'a',
////		BackgroundCode: 'x',
////	}
////	s := recordedToString(r)
////	_,r2 := recordedFromString(s)
////	fmt.Println(r2.ShownSymbol)
////}