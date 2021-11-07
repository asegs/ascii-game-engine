package main

const height int = 20
const width int = 40

func main(){
	go HandleLog()
	terminal := createTerminal(height,width)
	input := initializeInput()
	redBlock := initContext().addRgbStyleBg(255,0,0).finish()
	clear := initContext().addSimpleStyle(0).finish()
	var dir byte
	pos := Coord{
		Row: 0,
		Col: 0,
	}

	data := make([][] rune, height)
	for i := 0;i<height;i++{
		row := make([] rune], width)
		for b:=0;b<width;b++{
			row[i][b] = '0'
		}
	}

	for {
		dir = <- input.events
		switch dir {
		case MOVE_LEFT:
			if pos.Col > 0{
				pos.Col --
			}
			break
		case MOVE_RIGHT:
			if pos.Col < width - 1{
				pos.Col ++
			}
			break
		case MOVE_DOWN:
			if pos.Row < height - 1{
				pos.Row ++
			}
			break
		case MOVE_UP:
			if pos.Row > 0{
				pos.Row --
			}
			break
		}
		terminal.send(clear," ",0,0,false)
		terminal.send(redBlock,"*",pos.Row,pos.Col,true)
	}
}
