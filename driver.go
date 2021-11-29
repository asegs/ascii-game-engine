package main
const height int = 20
const width int = 40
func (terminal * Terminal) erasePath(p []*Coord){
	for i := 1;i<len(p) - 1;i++{
		terminal.sendUndoAtLocationConditional(p[i].Row,p[i].Col,'x')
	}
}
func (terminal * Terminal) drawPath(p []*Coord){
	for i := 1;i<len(p) - 1;i++{
		terminal.sendPlaceCharAtCoord('x',p[i].Row,p[i].Col)
	}
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
	},4)
	faces := make([] string,2)
	faces[0] = "assets/faces/simple_face.txt"
	faces[1] = "assets/faces/altered_face.txt"
	exps := make([] string,2)
	exps[0] = "normal"
	exps[1] = "white"
	face := buildFace(exps,faces,"guy")
	go terminal.cycleExpressions(face,exps,200,-1)
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
			if col < terminal.Width - 2{
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
			if terminal.DataHistory[row][col][terminal.Depth - 2].code == '1'{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('1',row,col)
			}
			break
		case '2':
			if terminal.DataHistory[row][col][terminal.Depth - 2].code == '2'{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('2',row,col)
			}
			break
		case '3':
			if terminal.DataHistory[row][col][terminal.Depth - 2].code == '3'{
				terminal.sendPlaceCharAtCoord('0',row,col)
			}else{
				terminal.sendPlaceCharAtCoord('3',row,col)
			}
			break
		case ENTER:
			if path != nil{
				terminal.erasePath(path)
				path = nil
			}
			maze,start,end := terminal.parseMazeFromCurrent('1','0','2','3')
			path = astar(maze,start,end)
			terminal.drawPath(path)
			break
		case BACKSLASH:
			if path != nil{
				terminal.erasePath(path)
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
