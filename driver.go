package main

func main(){
	terminal := createTerminal(20,40)
	input := initializeInput()
	redBlock := initContext().addRgbStyleBg(255,0,0).finish()
	var dir byte
	row := 0
	col := 0
	for {
		dir = <- input.events
		switch dir {
		case MOVE_LEFT:
			col --
		case MOVE_RIGHT:
			col++
			break
		case MOVE_DOWN:
			row++
			break
		case MOVE_UP:
			row--
			break
		}
		terminal.send(redBlock," ",row,col,true)
	}
}
