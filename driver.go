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

	row := 0
	col := 0

	terminal.sendCharAssociation('0',&Recorded{
		Format: clear,
		data:   ' ',
		code: '0',
	})
	terminal.sendCharAssociation('1',&Recorded{
		Format: blackBlock,
		data:   ' ',
		code: '1',
	})
	terminal.sendCharAssociation('2',&Recorded{
		Format: greenBlock,
		data:   ' ',
		code: '2',
	})
	terminal.sendCharAssociation('3',&Recorded{
		Format: blueBlock,
		data:   ' ',
		code: '3',
	})
	terminal.sendCharAssociation('*',&Recorded{
		Format: cursor,
		data:   '*',
		code: '*',
	})
	terminal.sendCharAssociation('x',&Recorded{
		Format: redBlock,
		data:   ' ',
		code:   'x',
	})

	for {
		rowChange = 0
		colChange = 0
		dir = <- input.events
		switch dir {
		case MOVE_LEFT:
			terminal.sendPlaceCharAtCoordCondUndo('*',row,col - 1,row,col,'*')
			if col > 0{
				col--
			}
			opType = 1
			break
		case MOVE_RIGHT:
			terminal.sendPlaceCharAtCoordCondUndo('*',row,col+1,row,col,'*')
			if col < terminal.Width - 1{
				col++
			}
			opType = 1
			break
		case MOVE_DOWN:
			terminal.sendPlaceCharAtCoordCondUndo('*',row+1,col,row,col,'*')
			if row < terminal.Height -1{
				row++
			}
			opType = 1
			break
		case MOVE_UP:
			terminal.sendPlaceCharAtCoordCondUndo('*',row-1,col,row,col,'*')
			if row > 0{
				row --
			}
			opType = 1
			break
		case '1':
			if terminal.CurrentData[terminal.Row][terminal.Col].code == '0'{
				terminal.sendPlaceCharAtCoord('1',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}
			opType = 2
			break
		case '2':
			if terminal.CurrentData[terminal.Row][terminal.Col].code == '0'{
				terminal.sendPlaceCharAtCoord('2',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}
			opType = 3
			break
		case '3':
			if terminal.CurrentData[terminal.Row][terminal.Col].code == '0'{
				terminal.sendPlaceCharAtCoord('3',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}
			opType = 4
			break
		case ENTER:

			maze,start,end := terminal.parseMazeFromCurrent('1','0','2','3')
			if len(path) == 0{
				for _,coord := range path{
					//just call an undo
					terminal.sendPlaceCharAtCoord()
				}
			}
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
