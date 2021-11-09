package main

const height int = 20
const width int = 40

func main(){
	terminal := createTerminal(height,width)
	input := initializeInput()
	cursor := initContext().addRgbStyleFg(255,0,0).finish()
	redBlock := initContext().addRgbStyleBg(255,0,0).finish()
	blackBlock := initContext().addRgbStyleBg(0,0,0).finish()
	greenBlock := initContext().addRgbStyleBg(0,255,0).finish()
	blueBlock := initContext().addRgbStyleBg(0,0,255).finish()
	clear := initContext().addSimpleStyle(0).finish()
	var dir byte

	opType := 0
	rowChange := 0
	colChange := 0
	var path []*Coord

	for {
		rowChange = 0
		colChange = 0
		dir = <- input.events
		switch dir {
		case MOVE_LEFT:
			terminal.sendPlaceCharAtShift('*',0,-1)
			terminal.undoConditional()
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
		case ENTER:
			maze,start,end := parseMazeFromChars(data,'1','0','2','3')
			path = astar(maze,start,end)
			for i := 1;i<len(path) - 1;i++{
				terminal.send(redBlock," ",path[i].Row,path[i].Col,true)
			}
			opType = 0
			break
		case BACKSLASH:
			for i := 1;i<len(path) - 1;i++{
				terminal.send(clear," ",path[i].Row,path[i].Col,true)
			}
			opType = 0
			break
		case BACKSPACE:
			for i := 0;i<height;i++{
				for b := 0;b<width;b++{
					data[i][b] = '0'
					terminal.send(clear," ",i,b,true)
				}
			}
			opType = 0
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
			terminal.send(cursor,"*",pos.Row,pos.Col,true)
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
