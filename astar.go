package main

import (
	"fmt"
	"math/rand"
	"time"
)

type TileType int

const (
	FREE TileType = iota
	WALL
	START
	END
)

const EROSIONS int = 10
const MIN_TO_ERODE int= 2
const MAX_TO_ERODE int = 4
const DIAG bool = false

var source = rand.NewSource(time.Now().UnixNano())
var random = rand.New(source)

type Coord struct {
	Row int
	Col int
}

type Node struct {
	PathParent * Node
	Arrival int
	Remaining int
	Pos * Coord
	Left * Node
	Right * Node
	GraphParent * Node
}

type PriorityQueue struct {
	 Head * Node
}

type Tile struct {
	Pos * Coord
	Type TileType
	Visited bool
}

var terms4Axis [4] * Coord = [4]*Coord{
	{
		Row: 0,
		Col: -1,
	},
	{
		Row: -1,
		Col: 0,
	},
	{
		Row: 0,
		Col: 1,
	},
	{
		Row: 1,
		Col: 0,
	},
}

var terms8Axis [8] * Coord = [8] * Coord{
	{
		Row: 0,
		Col: -1,
	},
	{
		Row: -1,
		Col: 0,
	},
	{
		Row: 0,
		Col: 1,
	},
	{
		Row: 1,
		Col: 0,
	},
	{
		Row: -1,
		Col: -1,
	},
	{
		Row: -1,
		Col: 1,
	},
	{
		Row: 1,
		Col: 1,
	},
	{
		Row: 1,
		Col: -1,
	},
}

func (n * Node) estimate()int{
	return n.Arrival + n.Remaining
}

func (c * Coord) matches (c1 * Coord) bool {
	return c.Row == c1.Row && c.Col == c1.Col
}


//logically speaking, shouldn't need to return anything
func insertNodeHelper(parent * Node,node * Node){
	if parent.Pos.matches(node.Pos){
		if node.estimate() < parent.estimate(){
			parent.PathParent = node.PathParent
			parent.Arrival = node.Arrival
		}
		return
	}
	if node.estimate() >= parent.estimate(){
		if parent.Right == nil{
			node.GraphParent = parent
			parent.Right = node
			return
		}
		insertNodeHelper(parent.Right, node)
	}else{
		if parent.Left == nil{
			node.GraphParent = parent
			parent.Left = node
			return
		}
		insertNodeHelper(parent.Left, node)
	}
}

func (queue * PriorityQueue) insert(node * Node){
	if queue.Head == nil {
		queue.Head = node
	}else {
		insertNodeHelper(queue.Head, node)
	}
}

func takeClosestHelper(parent * Node) * Node {
	if parent.Left == nil {
		if parent.Right != nil {
			parentRoot := parent.GraphParent
			parentRoot.Left = parent.Right
			parent.Right.GraphParent = parentRoot
		}else{
			parent.GraphParent.Left = nil
		}
		return parent
	}
	return takeClosestHelper(parent.Left)
}

func (queue * PriorityQueue) pop () * Node {
	if queue.Head == nil {
		return queue.Head
	}
	if queue.Head.Left == nil {
		toReturn := queue.Head
		if queue.Head.Right != nil {
			queue.Head = queue.Head.Right
		}else{
			queue.Head = nil
		}
		return toReturn
	}
	return takeClosestHelper(queue.Head)
}

func tileGood(maze [][] * Tile, pos * Coord)bool{
	return 0 <= pos.Row && pos.Row < len(maze) && 0 <= pos.Col && pos.Col < len(maze[0]) && !maze[pos.Row][pos.Col].Visited
}

func getCoordsForPair(pos * Coord, mod * Coord)* Coord{
	return &Coord{
		Row: pos.Row + mod.Row,
		Col: pos.Col + mod.Col,
	}
}

func bioticErode(maze  [][] * Tile ){
	for row,wholeRow := range maze{
		for col,_ := range wholeRow {
			surroundingWalls := 0
			for _,term := range terms8Axis{
				if !tileGood(maze,getCoordsForPair(&Coord{Row: row, Col: col,},term)){
					surroundingWalls++
				}
			}
			if surroundingWalls > MAX_TO_ERODE {
				maze[row][col].Type = WALL
				maze[row][col].Visited = true
			}else if surroundingWalls < MIN_TO_ERODE {
				maze[row][col].Type = FREE
				maze[row][col].Visited = false
			}
		}
	}
}

func generateMaze(width int,height int,freq float64)([][] * Tile,*Coord,*Coord) {
	maze := make([][] * Tile,height)
	for i := 0;i<height;i++{
		maze[i] = make([] * Tile,width)
		for b := 0;b<width;b++{
			if random.Float64() < freq{
				maze[i][b] = &Tile{
					Pos:     &Coord{
						Row: i,
						Col: b,
					},
					Type:    WALL,
					Visited: true,
				}
			}else{
				maze[i][b] = &Tile{
					Pos:     &Coord{
						Row: i,
						Col: b,
					},
					Type:    FREE,
					Visited: false,
				}
			}
		}
	}

	for i := 0;i<EROSIONS;i++{
		bioticErode(maze)
	}

	start := &Coord{
		Row: random.Intn(height-1),
		Col: random.Intn(width - 1),
	}
	maze[start.Row][start.Col].Type = START
	maze[start.Row][start.Col].Visited = false

	end := &Coord{
		Row: random.Intn(height - 1),
		Col: random.Intn(width - 1),
	}
	maze[end.Row][end.Col].Type = END
	maze[end.Row][end.Col].Visited = false

	return maze,start,end
}

func parseMazeFromChars(data [][] rune,wall rune,free rune,start rune,end rune)([][] * Tile,*Coord,*Coord){
	height := len(data)
	width := len(data[0])
	maze := make([][] * Tile, height)
	startCoord := &Coord{
		Row: 0,
		Col: 0,
	}

	endCoord := &Coord{
		Row: 0,
		Col: 0,
	}
	for i := 0;i<height;i++{
		row := make([] * Tile, width)
		maze[i] = row
		for b := 0;b<width;b++{
			maze[i][b] = &Tile{
				Pos:     &Coord{
					Row: i,
					Col: b,
				},
				Type:    FREE,
				Visited: false,
			}
				switch data[i][b] {
				case wall:
					maze[i][b].Type = WALL
					maze[i][b].Visited = true
					break
				case start:
					maze[i][b].Type = START
					maze[i][b].Visited = false
					startCoord.Row = i
					startCoord.Col = b
					break
				case end:
					maze[i][b].Type = END
					maze[i][b].Visited = false
					endCoord.Row = i
					endCoord.Col = b
					break
				}
		}
	}
	return maze,startCoord,endCoord
}

func square(n int)int{
	return n*n
}

func pythagDistance(c1 * Coord,c2 * Coord)int{
	return square(c1.Row - c2.Row) + square(c1.Col - c2.Col)
}

func getAdjacentValidTiles(maze [][] * Tile,pos * Coord) [] * Coord  {
	resultSize := 4
	if DIAG{
		resultSize = 8
	}
	adjacents := make([] * Coord,resultSize)
	for i,term := range terms8Axis[0:resultSize]{
		newCoord := getCoordsForPair(pos,term)
		if tileGood(maze,newCoord){
			adjacents[i] = newCoord
		}else{
			adjacents[i] = nil
		}
	}
	return adjacents
}

func unwrapPath(end * Node)  [] * Coord {
	path := make([] * Coord,end.Arrival + 1)
	for end != nil {
		path[end.Arrival] = &Coord{
			Row: end.Pos.Row,
			Col: end.Pos.Col,
		}
		end = end.PathParent
	}
	return path
}

func astar(maze [][] * Tile,start * Coord,end * Coord) [] * Coord {
	st := time.Now()
	s := &Node{
		PathParent:  nil,
		Arrival:     0,
		Remaining:   pythagDistance(start,end),
		Pos:         start,
		Left:        nil,
		Right:       nil,
		GraphParent: nil,
	}
	queue := PriorityQueue{Head: s}
	var position * Node
	var mazeData * Tile
	maxAdj := 4
	if DIAG{
		maxAdj = 8
	}
	adjacent := make([] * Coord,maxAdj)
	for queue.Head != nil {
		position = queue.pop()
		mazeData = maze[position.Pos.Row][position.Pos.Col]
		if mazeData.Visited{
			continue
		}
		if position.Pos.matches(end){
			ft := time.Now()
			fmt.Println(ft.Sub(st))
			return unwrapPath(position)
		}
		mazeData.Visited = true
		adjacent = getAdjacentValidTiles(maze,position.Pos)
		for _,tile := range adjacent{
			if tile != nil{
				queue.insert(&Node{
					PathParent:  position,
					Arrival:     position.Arrival + 1,
					Remaining:   pythagDistance(tile,end),
					Pos:         tile,
					Left:        nil,
					Right:       nil,
					GraphParent: nil,
				})
			}
		}
	}
	ft := time.Now()
	fmt.Println(ft.Sub(st))
	return nil
}

func printRed(txt string){
	fmt.Printf("\033[48;2;255;0;0m%s\033[0m",txt)
}

func printBlue(txt string){
	fmt.Printf("\033[48;2;0;208;233m%s\033[0m",txt)
}

func printGreen(txt string){
	fmt.Printf("\033[48;2;9;152;13m%s\033[0m",txt)
}

func printBlack(txt string){
	fmt.Printf("\033[48;2;0;0;0m%s\033[0m",txt)
}

func printNormal(txt string){
	fmt.Printf("\033[48;2;255;255;255m%s\033[0m",txt)
}

func inPath(c * Coord,path [] * Coord) bool{
	for _,item := range path{
		if item.Row == c.Row && item.Col == c.Col{
			return true
		}
	}
	return false
}

func display(maze [][] * Tile,path [] * Coord){
	for _,wholeRow := range maze {
		for _,item := range wholeRow {
			if item.Type == START{
				printBlue(" ")
			}else if item.Type == END {
				printGreen(" ")
			}else if item.Type == WALL {
				printBlack(" ")
			}else if inPath(item.Pos,path){
				printRed(" ")
			}else{
				printNormal(" ")
			}
		}
		println()
	}
}

//func main(){
//	maze,start,end := generateMaze(180,40,0.3)
//	pth := 	astar(maze,start,end)
//	display(maze,pth)
//
//}
