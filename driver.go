package main

func main(){
	go HandleLog()
	terminal := createTerminal(20,40)
	input := initializeInput()
	redBlock := initContext().addRgbStyleBg(255,0,0).finish()
	clear := initContext().addSimpleStyle(0).finish()
	var dir byte

	for {
		dir = <- input.events
		row := 0
		col := 0
		switch dir {
		case MOVE_LEFT:
			col = -1
			break
		case MOVE_RIGHT:
			col = 1
			break
		case MOVE_DOWN:
			row = 1
			break
		case MOVE_UP:
			row = -1
			break
		}
		terminal.send(clear," ",0,0,false)
		terminal.send(redBlock,"*",row,col,false)
	}
}
