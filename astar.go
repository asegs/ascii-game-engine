package main

import (
	"math/rand"
	"time"
)

type TileType int

//Defines what the types of a maze tile can be.
const (
	FREE TileType = iota
	WALL
	START
	END
)

//Used for generating biotic maps, how many erosion cycles are made.
//Move to map gen package
const EROSIONS int = 10
//If a tile has less than MIN_TO_ERODE neighbors, it is cleared.
const MIN_TO_ERODE int= 2
//If a tile has more than MAX_TO_ERODE neighbors, it is filled.
const MAX_TO_ERODE int = 4

//If the path can include diagonal movements.
const DIAG bool = false

//The source of map generation randomness.
var source = rand.NewSource(time.Now().UnixNano())
//The random object used.
var random = rand.New(source)

//A holder for a map coordinate.
type Coord struct {
	Row int
	Col int
}

/**
A PriorityQueue Node that tracks a lot of state:
PathParent: The Node on the map that led to this Node.
Arrival: How many steps from the start Node it was to this Node.
Remaining: The heuristic estimate of how many more steps will be needed to reach the end Node.
Pos: The Coord position of the Node on the map.
Left: The Node to the left on the PriorityQueue tree.
Right: The Node to the right on the PriorityQueue tree.
GraphParent: The parent Node on the PriorityQueue tree of this Node.
 */
type Node struct {
	PathParent * Node
	Arrival int
	Remaining int
	Pos * Coord
	Left * Node
	Right * Node
	GraphParent * Node
}

//Just a wrapper for a Head Node.
type PriorityQueue struct {
	 Head * Node
}

/**
A representation of a map tile:
Pos: The coordinates of the Tile.
Type: The TileType (FREE,WALL,START,END) of the Tile.
Visited: If the Tile has already been visited.
 */
type Tile struct {
	Pos * Coord
	Type TileType
	Visited bool
}

//The possible moves that a path could take given diagonal movement.
//The first four are only lateral moves.
var terms8Axis = [8] * Coord{
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

//Estimates the remaining distance to the end from a given Tile.
func (n * Node) estimate()int{
	return square(n.Arrival) + n.Remaining
}

//Checks to see if two coordinates are in the same position.
func (c * Coord) matches (c1 * Coord) bool {
	return c.Row == c1.Row && c.Col == c1.Col
}

/**
Inserts a Tile Node into the PriorityQueue recursively.
parent: the Node to insert the new Node under.
node: the Node to insert into the PriorityQueue.
 */
func insertNodeHelper(parent * Node,node * Node){
	//If the parent is the child node:
	if parent.Pos.matches(node.Pos){
		//If the node was found in a faster way than last time:
		if node.Arrival < parent.Arrival{
			//Set the path that the parent came from to whatever the child came from.
			parent.PathParent = node.PathParent
			//Set the arrival time at the parent to the faster one of the child.
			parent.Arrival = node.Arrival
		}
		return
	}
	//If the child is farther from the end than parent:
	if node.estimate() >= parent.estimate(){
		//If the parent has no right Node, add the child.
		if parent.Right == nil{
			node.GraphParent = parent
			parent.Right = node
			return
		}
		//Else if it does, insert it under the right Node recursively.
		insertNodeHelper(parent.Right, node)
	//If the child is closer to the end than the parent:
	} else{
		//If the parent has no left Node, add the child.
		if parent.Left == nil{
			node.GraphParent = parent
			parent.Left = node
			return
		}
		//Else if it does, insert it under the left Node recursively.
		insertNodeHelper(parent.Left, node)
	}
}

/**
Inserts a Node into a PriorityQueue and handles the empty queue edge case.
 */
func (queue * PriorityQueue) insert(node * Node){
	//If the queue has nothing in it, set the head to the child Node.
	if queue.Head == nil {
		queue.Head = node
	//If the queue has a head, insert the child Node under the head.
	}else {
		insertNodeHelper(queue.Head, node)
	}
}

/**
Takes the Node with the shortest estimate under the parent.
 */
func takeClosestHelper(parent * Node) * Node {
	//If the parent has no left Node:
	if parent.Left == nil {
		//If the parent has a right Node, connects it to it's parent's left Node and removes itself.
		if parent.Right != nil {
			parentRoot := parent.GraphParent
			parentRoot.Left = parent.Right
			parent.Right.GraphParent = parentRoot
		//If the parent has no right node, just sets its own parent's left to nil and removes itself.
		}else{
			parent.GraphParent.Left = nil
		}
		return parent
	}
	//If left is not nil, recursively call this function again with the left node.
	return takeClosestHelper(parent.Left)
}


/**
Removes the Node with the least distance (arrival + estimated) to the end from a PriorityQueue.  Returns it.
Also handles some edge cases.
 */
func (queue * PriorityQueue) pop () * Node {
	//If the queue is empty, return nil
	if queue.Head == nil {
		return queue.Head
	}
	//If the queue has no left, return the head after setting the head's right as the new head.
	if queue.Head.Left == nil {
		toReturn := queue.Head
		queue.Head = queue.Head.Right
		return toReturn
	}
	//If the queue has a left, get the farthest left element recursively.
	return takeClosestHelper(queue.Head)
}

/**
Checks if a given tile is valid for maze exploration.
First checks if it is within the given bounds (x and y).
Next checks if the tile has already been visited.
 */
func tileGood(maze [][] * Tile, pos * Coord)bool{
	return 0 <= pos.Row && pos.Row < len(maze) && 0 <= pos.Col && pos.Col < len(maze[0]) && !maze[pos.Row][pos.Col].Visited
}

/**
Applies a coordinate shift to a pair of coordinates and returns a new Coord.

For example: (12,15),(1,0) returns (13,15).
 */
func getCoordsForPair(pos * Coord, mod * Coord)* Coord{
	return &Coord{
		Row: pos.Row + mod.Row,
		Col: pos.Col + mod.Col,
	}
}

/**
Applies simple biotic rules to soften edges of the maze.
Coordinates surrounded with mostly free space turn into free space.
Coordinates surrounded with mostly walls turn into walls.

Tiles bordering out of bounds spaces are counted as bordering walls there.
Doesn't validate for start/end, these would be added later.  Could check!
 */
func bioticErode(maze  [][] * Tile ){
	for row,wholeRow := range maze{
		for col := range wholeRow {
			surroundingWalls := 0
			for _,term := range terms8Axis{
				if !tileGood(maze,getCoordsForPair(&Coord{Row: row, Col: col},term)){
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

/**
Randomly generates a maze of a certain size with a wall frequency.
Returns the maze as well as the start and end positions.
 */
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

func (t * Terminal) toString()string{
	str := ""
	for _,row := range t.DataHistory {
		sub := make([]byte,t.Width)
		//s1 := ""
		for b,col := range row {
			sub[b] = col.top().BackgroundCode
			//s1 += string(int(col[t.Depth - 1].BackgroundCode))
		}
		str += string(sub) + "\n"
	}
	return str
}

/**
Reads the terminal space and returns a parsed map.
Takes the symbol that represents a wall, the start, and the end.
Returns a maze array (2d) and the start/end coordinates.
 */
func (t * Terminal) parseMazeFromCurrent(wall byte, start byte, end byte) ([][]*Tile, *Coord, *Coord) {
	//The dimensions of the terminals memory
	height := len(t.DataHistory)
	width := len(t.DataHistory[0]) - 1
	maze := make([][] * Tile, height)
	//Create default start and end coords
	startCoord := &Coord{
		Row: 0,
		Col: 0,
	}
	endCoord := &Coord{
		Row: 0,
		Col: 0,
	}
	discoveredWalls := 0
	//Reads through each tile in the array and switches on the background code to set the maze
	for i := 0;i<height;i++{
		row := make([] * Tile, width)
		maze[i] = row
		for b := 0;b<width;b++{
			//Default is free space.
			maze[i][b] = &Tile{
				Pos:     &Coord{
					Row: i,
					Col: b,
				},
				Type:    FREE,
				Visited: false,
			}
			switch t.DataHistory[i][b].top().BackgroundCode {
			case wall:
				maze[i][b].Type = WALL
				maze[i][b].Visited = true
				discoveredWalls++
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

/**
Very similar to above function, but it takes the array of characters as an argument.  Might refactor and combine these two.
 */
func parseMazeFromChars(data [][]rune, wall rune, start rune, end rune) ([][]*Tile, *Coord, *Coord) {
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

//Quickly squares two numbers.
func square(n int)int{
	return n*n
}

/**
Finds the non square rooted distance between two Coordinates.
 */
func pythagDistance(c1 * Coord,c2 * Coord)int{
	return square(c1.Row - c2.Row) + square(c1.Col - c2.Col)
}

/**
For a given maze and coordinate, gets an array of all adjacent coordinates.
If a coordinate is invalid, fills its array slot with nil.
 */
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

/**
Given an end Node with a path parent leading backwards down the correct path,
returns an array of Nodes composing the path.
 */
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

/**
Performs the A* pathfinding algorithm on a maze of Tiles.  Returns the path as an array.
 */
func astar(maze [][] * Tile,start * Coord,end * Coord) [] * Coord {
	//Creates the start node.
	s := &Node{
		PathParent:  nil,
		Arrival:     0,
		Remaining:   pythagDistance(start,end),
		Pos:         start,
		Left:        nil,
		Right:       nil,
		GraphParent: nil,
	}
	//Creates the PriorityQueue with the start Node as the head.
	queue := PriorityQueue{Head: s}
	//Defines a variable to track current position.
	var position * Node
	//Defines a variable to track the current maze data.
	var mazeData * Tile
	//Creates a reused array of coords the size of maxAdj.
	var adjacent [] * Coord
	//While the queue has Nodes in it:
	for queue.Head != nil {
		//Get the closest node as position.
		position = queue.pop()
		//Get the details about that node as mazeData.
		mazeData = maze[position.Pos.Row][position.Pos.Col]
		//If this Node has already been explored, skip it and let it die.
		//This happens when two tiles adjacent to this one have both been considered.
		if mazeData.Visited{
			continue
		}
		//If the new Node is the end Node:
		if position.Pos.matches(end){
			//Return the path that landed at the end Node.
			return unwrapPath(position)
		}
		//Marks the tile as visited.
		mazeData.Visited = true
		//Gets all adjacent spots to the current position.
		adjacent = getAdjacentValidTiles(maze,position.Pos)
		for _,tile := range adjacent{
			if tile != nil{
				//For all valid adjacent Nodes, insert them into the PriorityQueue.
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
	return nil
}
