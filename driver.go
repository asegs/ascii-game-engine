package main

const height int = 20
const width int = 40

func main(){
	go HandleLog()
	terminal := createTerminal(height,width)
	input := initializeInput()
	redBlock := initContext().addRgbStyleBg(255,0,0).finish()
	blackBlock := initContext().addRgbStyleBg(0,0,0).finish()
	greenBlock := initContext().addRgbStyleBg(0,255,0).finish()
	blueBlock := initContext().addRgbStyleBg(0,0,255).finish()
	clear := initContext().addSimpleStyle(0).finish()
	var dir byte
	pos := Coord{
		Row: 0,
		Col: 0,
	}

	data := make([][] rune, height)
	for i := 0;i<height;i++{
		row := make([] rune, width)
		for b:=0;b<width;b++{
			row[b] = '0'
		}
		data[i] = row
	}

	opType := 0
	rowChange := 0
	colChange := 0

	for {
		rowChange = 0
		colChange = 0
		dir = <- input.events
		switch dir {
		case MOVE_LEFT:
			if pos.Col > 0{
				pos.Col --
				colChange = -1
			}
			opType = 1
			break
		case MOVE_RIGHT:
			if pos.Col < width - 1{
				pos.Col ++
				colChange = 1
			}
			opType = 1
			break
		case MOVE_DOWN:
			if pos.Row < height - 1{
				pos.Row ++
				rowChange = 1
			}
			opType = 1
			break
		case MOVE_UP:
			if pos.Row > 0{
				pos.Row --
				rowChange = -1
			}
			opType = 1
			break
		case '1':
			if data[pos.Row][pos.Col] == '1'{
				data[pos.Row][pos.Col] = '0'
			}else {
				data[pos.Row][pos.Col] = '1'
			}
			opType = 2
			break
		case '2':
			if data[pos.Row][pos.Col] == '2'{
				data[pos.Row][pos.Col] = '0'
			}else {
				data[pos.Row][pos.Col] = '2'
			}
			opType = 3
			break
		case '3':
			if data[pos.Row][pos.Col] == '3'{
				data[pos.Row][pos.Col] = '0'
			}else {
				data[pos.Row][pos.Col] = '3'
			}
			opType = 4
			break
		}
		switch opType {
		case 1:
			prevRow := pos.Row - rowChange
			prevCol := pos.Col - colChange
			prev := data[prevRow][prevCol]
			switch prev {
			case '0':
				terminal.send(clear," ",prevRow,prevCol,true)
				break
			case '1':
				terminal.send(blackBlock," ",prevRow,prevCol,true)
				break
			case '2':
				terminal.send(greenBlock," ",prevRow,prevCol,true)
				break
			case '3':
				terminal.send(blueBlock," ",prevRow,prevCol,true)
			}
			terminal.send(redBlock,"*",pos.Row,pos.Col,true)
			break
		case 2:
			if data[pos.Row][pos.Col] == '0'{
				terminal.send(clear," ",pos.Row,pos.Col,true)
			}else{
				terminal.send(blackBlock," ",pos.Row,pos.Col,true)
			}
			break
		case 3:
			if data[pos.Row][pos.Col] == '0'{
				terminal.send(clear," ",pos.Row,pos.Col,true)
			}else{
				terminal.send(greenBlock," ",pos.Row,pos.Col,true)
			}
			break
		case 4:
			if data[pos.Row][pos.Col] == '0'{
				terminal.send(clear," ",pos.Row,pos.Col,true)
			}else{
				terminal.send(blueBlock," ",pos.Row,pos.Col,true)
			}
			break
		}

	}
}
