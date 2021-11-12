package main
const height int = 20
const width int = 40



func (terminal * Terminal) assoc(char byte,format * Context,txt byte){
	terminal.sendCharAssociation(char,&Recorded{
		Format: format,
		data:   txt,
		code: char,
	})
}

func main(){


	input := initializeInput()
	cursor := initContext().addRgbStyleFg(255,0,0).finish()
	redBlock := initContext().addRgbStyleBg(255,0,0).finish()
	blackBlock := initContext().addRgbStyleBg(0,0,0).finish()
	greenBlock := initContext().addRgbStyleBg(0,255,0).finish()
	blueBlock := initContext().addRgbStyleBg(0,0,255).finish()
	clear := initContext().addSimpleStyle(0).finish()

	terminal := createTerminal(height,width,&Recorded{
		Format: clear,
		data: ' ',
		code: '0',
	})

	var dir byte

	var path []*Coord
	path = nil

	row := 0
	col := 0

	terminal.assoc('0',clear,' ')
	terminal.assoc('1',blackBlock,' ')
	terminal.assoc('2',greenBlock,' ')
	terminal.assoc('3',blueBlock,' ')
	terminal.assoc('*',cursor,'*')
	terminal.assoc('x',redBlock,' ')

	for {
		dir = <- input.events
		switch dir {
		case MOVE_LEFT:
			if col > 0{
				col--
				terminal.sendPlaceCharAtCoordCondUndo('*',row,col,row,col+1,'*')
			}
			break
		case MOVE_RIGHT:
			if col < terminal.Width - 1{
				col++
				terminal.sendPlaceCharAtCoordCondUndo('*',row,col,row,col-1,'*')
			}
			break
		case MOVE_DOWN:
			if row < terminal.Height -1{
				row++
				terminal.sendPlaceCharAtCoordCondUndo('*',row,col,row-1,col,'*')
			}
			break
		case MOVE_UP:
			if row > 0{
				row--
				terminal.sendPlaceCharAtCoordCondUndo('*',row,col,row+1,col,'*')
			}
			break
		case '1':
			if terminal.StoredData[row][col].code == '0'{
				terminal.sendPlaceCharAtCoord('1',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}
			break
		case '2':
			if terminal.StoredData[row][col].code == '0'{
				terminal.sendPlaceCharAtCoord('2',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}
			break
		case '3':
			if terminal.StoredData[row][col].code == '0'{
				terminal.sendPlaceCharAtCoord('3',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}
			break
		case ENTER:


			if path != nil{
				for _,coord := range path{
					terminal.sendUndoAtLocationConditional(coord.Row,coord.Col,'x')
				}
				path = nil
			}
			maze,start,end := terminal.parseMazeFromCurrent('1','0','2','3')
			path = astar(maze,start,end)
			for _,coord := range path{
				terminal.sendPlaceCharAtCoord('x',coord.Row,coord.Col)
			}
			break
		case BACKSLASH:
			if path != nil{
				for _,coord := range path{
					terminal.sendUndoAtLocationConditional(coord.Row,coord.Col,'x')
				}
				path = nil
			}
			break
		case BACKSPACE:
			for i := 0;i<height;i++{
				for b := 0;b<width;b++{
					terminal.sendPlaceCharAtCoord('0',i,b)
				}
			}
			break

		}

	}
}
