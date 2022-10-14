package main

import (
	"fmt"
	"math/rand"
	"time"
)

const mapWidth int = 60
const mapHeight int = 60

var undevelopedLevels = [...]byte{'%', '&', '#'}

var directions = [...]byte{MOVE_UP, MOVE_RIGHT, MOVE_DOWN, MOVE_LEFT}
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type Requirements struct {
	Logs  int
	Stone int
	Metal int
}

type Structure struct {
	Finished     bool
	Pos          *Coord
	Requirements []*Requirements
	Provided     []*Requirements
}

type PlayerState struct {
	Pos       *Coord
	Health    int
	Water     int
	Food      int
	LastSleep time.Time
	Logs      int
	Stone     int
	Metal     int
}

type GlobalState struct {
	Grid                  *[][]byte
	StoredLogs            int
	StoredStone           int
	StoredMetal           int
	HousingSpace          int
	Appeal                int
	Unassigned            int
	Loggers               int
	Miners                int
	Metalworkers          int
	Clearers              int
	LogsPerMinute         int
	StonePerMinute        int
	MetalPerMinute        int
	ImprovementsPerMinute int
	PeoplesHappiness      int
	QuarrySumLevel        int
	MetalworksSumLevel    int
	RoadUnlocked          int
	Unfinished            []*Structure
}

func newGlobalState(grid *[][]byte) *GlobalState {
	return &GlobalState{
		Grid:                  grid,
		StoredLogs:            0,
		StoredStone:           0,
		StoredMetal:           0,
		HousingSpace:          0,
		Appeal:                0,
		Unassigned:            0,
		Loggers:               0,
		Miners:                0,
		Metalworkers:          0,
		Clearers:              0,
		LogsPerMinute:         0,
		StonePerMinute:        0,
		MetalPerMinute:        0,
		ImprovementsPerMinute: 0,
		PeoplesHappiness:      0,
		QuarrySumLevel:        0,
		MetalworksSumLevel:    0,
		RoadUnlocked:          0,
		Unfinished:            make([]*Structure, 0),
	}
}

func RandInt(min int, max int) int {
	return min + r.Intn(max-min+1)
}

func inBounds(pos *Coord) bool {
	return pos.Row >= 0 && pos.Row < mapHeight && pos.Col >= 0 && pos.Col < mapWidth
}

func matches(grid *[][]byte, toMatch byte, pos *Coord) bool {
	return inBounds(pos) && (*grid)[pos.Row][pos.Col] == toMatch
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func applyToGrid(grid *[][]byte, apply func(grid *[][]byte, row int, col int)) *[][]byte {
	for row := 0; row < mapHeight; row++ {
		for col := 0; col < mapWidth; col++ {
			apply(grid, row, col)
		}
	}
	return grid
}

func countBorders(grid *[][]byte, match byte, pos *Coord, includeDiagonal bool) int {
	borders := 0
	if includeDiagonal {
		borders = b2i(matches(grid, match, &Coord{Row: pos.Row - 1, Col: pos.Col - 1})) + b2i(matches(grid, match, &Coord{Row: pos.Row + 1, Col: pos.Col - 1})) + b2i(matches(grid, match, &Coord{Row: pos.Row - 1, Col: pos.Col + 1})) + b2i(matches(grid, match, &Coord{Row: pos.Row + 1, Col: pos.Col + 1}))
	}
	borders += b2i(matches(grid, match, &Coord{Row: pos.Row - 1, Col: pos.Col})) + b2i(matches(grid, match, &Coord{Row: pos.Row, Col: pos.Col - 1})) + b2i(matches(grid, match, &Coord{Row: pos.Row, Col: pos.Col + 1})) + b2i(matches(grid, match, &Coord{Row: pos.Row + 1, Col: pos.Col}))
	return borders
}

func hasBorder(grid *[][]byte, match byte, pos *Coord, includeDiagonal bool) bool {
	return countBorders(grid, match, pos, includeDiagonal) > 0
}

func drawStreams(dryMap *[][]byte, initialWaters int, waterPercentage float64, traces int) *[][]byte {
	tiles := mapWidth * mapHeight
	//Spawning n water sources.
	for i := 0; i < initialWaters; i++ {
		place := RandInt(0, tiles-1)
		row := place / mapWidth
		col := place - row*mapWidth
		(*dryMap)[row][col] = ' '
	}
	//Drawing river paths.
	for i := 0; i < traces; i++ {
		//Drawing rightwards.
		for row := 0; row < mapHeight; row++ {
			for col := 0; col < mapWidth; col++ {
				if (*dryMap)[row][col] == 'a' && r.Float64() < waterPercentage && hasBorder(dryMap, ' ', &Coord{Row: row, Col: col}, true) {
					(*dryMap)[row][col] = ' '
				}
			}
		}
		//Drawing leftwards.
		for row := mapHeight - 1; row >= 0; row-- {
			for col := mapWidth - 1; col >= 0; col-- {
				if (*dryMap)[row][col] == 'a' && r.Float64() < waterPercentage && hasBorder(dryMap, ' ', &Coord{Row: row, Col: col}, true) {
					(*dryMap)[row][col] = ' '
				}
			}
		}
	}
	return dryMap

}

func erode(wetMap *[][]byte, cycles int, tolerance int) *[][]byte {
	for i := 0; i < cycles; i++ {
		applyToGrid(wetMap, func(grid *[][]byte, row int, col int) {
			borders := countBorders(grid, ' ', &Coord{Row: row, Col: col}, true)
			if borders > tolerance {
				(*grid)[row][col] = ' '
			}
		})
	}
	return wetMap
}

func drawLand(erodedMap *[][]byte, ruggedness float64) *[][]byte {
	ruggedness = ruggedness / 2
	applyToGrid(erodedMap, func(grid *[][]byte, row int, col int) {
		land := rand.Float64() / ruggedness
		if (*grid)[row][col] == 'a' {
			if land < 0.333 {
				(*grid)[row][col] = undevelopedLevels[2]
			} else if land < 0.666 {
				(*grid)[row][col] = undevelopedLevels[1]
			} else {
				(*grid)[row][col] = undevelopedLevels[0]
			}
		}
	})
	return erodedMap
}

func generateGrid() *[][]byte {
	//Allocating empty map buffer.
	m := make([][]byte, mapHeight)
	for i := 0; i < mapHeight; i++ {
		row := make([]byte, mapWidth)
		for j := 0; j < mapWidth; j++ {
			row[j] = 'a'
		}
		m[i] = row
	}
	//Drawing initial rivers using basic cellular automata.  Using some hardcoded values for now.
	mRef := drawStreams(&m, 3, 0.35, 3)
	//Eroding land to form lakes and dry up isolated ponds.
	mRef = erode(mRef, 5, 6)
	mRef = drawLand(mRef, 1)
	return mRef
}

func newPlayer(grid *[][]byte) *PlayerState {
	tiles := mapWidth * mapHeight
	player := &PlayerState{
		Pos:       nil,
		Health:    10,
		Water:     100,
		Food:      100,
		LastSleep: time.Now(),
		Logs:      0,
		Stone:     0,
		Metal:     0,
	}
	for true {
		placement := RandInt(0, tiles-1)
		row := int(placement / mapWidth)
		col := placement - row*mapWidth
		pos := &Coord{Row: row, Col: col}
		count := countBorders(grid, ' ', pos, true)
		if count <= 1 && (*grid)[row][col] != ' ' {
			player.Pos = pos
			break
		}
	}
	return player
}

func selectNew(direction byte, player *PlayerState) *Coord {
	if time.Now().Sub(player.LastSleep) > 10*time.Minute {
		direction = directions[RandInt(0, 3)]
	}
	if direction == MOVE_UP {
		return &Coord{Row: player.Pos.Row - 1, Col: player.Pos.Col}
	}
	if direction == MOVE_RIGHT {
		return &Coord{Row: player.Pos.Row, Col: player.Pos.Col + 1}
	}
	if direction == MOVE_DOWN {
		return &Coord{Row: player.Pos.Row + 1, Col: player.Pos.Col}
	}
	return &Coord{Row: player.Pos.Row, Col: player.Pos.Col - 1}
}

func move(direction byte, player *PlayerState, grid *[][]byte) {
	newCoord := selectNew(direction, player)
	if inBounds(newCoord) && (*grid)[newCoord.Row][newCoord.Col] != ' ' {
		player.Pos = newCoord
	}
}

func clearOnMove(lastPos *Coord, grid *[][]byte) bool {
	if (*grid)[lastPos.Row][lastPos.Col] == '#' {
		(*grid)[lastPos.Row][lastPos.Col] = '&'
		return true
	}
	if (*grid)[lastPos.Row][lastPos.Col] == '&' {
		(*grid)[lastPos.Row][lastPos.Col] = '%'
		return true
	}
	if (*grid)[lastPos.Row][lastPos.Col] == '%' {
		(*grid)[lastPos.Row][lastPos.Col] = '.'
		return true
	}
	return false
}

//Use a little goroutine and done channel to do this instead of blocking the main server thread
func waitForTerrain(lastPos *Coord, grid *[][]byte) int {
	if (*grid)[lastPos.Row][lastPos.Col] == '#' {
		return 5
	}
	if (*grid)[lastPos.Row][lastPos.Col] == '&' {
		return 3
	}
	if (*grid)[lastPos.Row][lastPos.Col] == '%' {
		return 1
	}
	return 0
}

func hardCastToState(playerState interface{}) *PlayerState {
	return playerState.(*PlayerState)
}

//No saving yet.
func savePlayerState(id int) {}

func serve() {
	serverConfig, err := loadServerConfig("configs/server_network_setting.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	playerStates := make(map[int]interface{})
	gameMap := generateGrid()
	globalState := newGlobalState(gameMap)

	server := newServerDefault(func(id int) {
		playerStates[id] = newPlayer(gameMap)
	}, savePlayerState, serverConfig, globalState, playerStates)

	handlers := server.newZoneHandlers("map")
	//When I tried to do this with a range, I had a bizarre issue with closures I think?
	handlers.addPlayerHandler(MOVE_UP, func(id int) {
		wait := waitForTerrain(hardCastToState(playerStates[id]).Pos, gameMap)
		go func() {
			time.Sleep(time.Duration(wait) * time.Second)
			if clearOnMove(hardCastToState(playerStates[id]).Pos, gameMap) {
				server.broadcastStateUpdate(globalState, GLOBAL_ID, true, "Grid")
			}
			move(MOVE_UP, hardCastToState(playerStates[id]), gameMap)
			server.broadcastStateUpdate(playerStates[id], id, true, "Pos")
		}()
	})
	handlers.addPlayerHandler(MOVE_RIGHT, func(id int) {
		wait := waitForTerrain(hardCastToState(playerStates[id]).Pos, gameMap)
		go func() {
			time.Sleep(time.Duration(wait) * time.Second)
			if clearOnMove(hardCastToState(playerStates[id]).Pos, gameMap) {
				server.broadcastStateUpdate(globalState, GLOBAL_ID, true, "Grid")
			}
			move(MOVE_RIGHT, hardCastToState(playerStates[id]), gameMap)
			server.broadcastStateUpdate(playerStates[id], id, true, "Pos")
		}()
	})
	handlers.addPlayerHandler(MOVE_DOWN, func(id int) {
		wait := waitForTerrain(hardCastToState(playerStates[id]).Pos, gameMap)
		go func() {
			time.Sleep(time.Duration(wait) * time.Second)
			if clearOnMove(hardCastToState(playerStates[id]).Pos, gameMap) {
				server.broadcastStateUpdate(globalState, GLOBAL_ID, true, "Grid")
			}
			move(MOVE_DOWN, hardCastToState(playerStates[id]), gameMap)
			server.broadcastStateUpdate(playerStates[id], id, true, "Pos")
		}()
	})
	handlers.addPlayerHandler(MOVE_LEFT, func(id int) {
		wait := waitForTerrain(hardCastToState(playerStates[id]).Pos, gameMap)
		go func() {
			time.Sleep(time.Duration(wait) * time.Second)
			if clearOnMove(hardCastToState(playerStates[id]).Pos, gameMap) {
				server.broadcastStateUpdate(globalState, GLOBAL_ID, true, "Grid")
			}
			move(MOVE_LEFT, hardCastToState(playerStates[id]), gameMap)
			server.broadcastStateUpdate(playerStates[id], id, true, "Pos")
		}()
	})
	server.start()
	for true {
		time.Sleep(1 * time.Second)
	}

}
