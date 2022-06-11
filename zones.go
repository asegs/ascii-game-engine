package main

import (
	"errors"
	"time"
)

type Zoning struct {
	Zones [] * Zone
	Height int
	Width int
	CursorZoneMap map[int] * Zone
	Input * NetworkedStdIn
	Default * Zone
	Terminal AbstractClient
}

type Zone struct {
	Y int
	X int
	Height int
	Width int
	CursorAllowed bool
	Events chan byte
	CursorMap map[int] * Coord
	Parent * Zoning
}

func initZones (height int,width int, input * NetworkedStdIn, term AbstractClient) * Zoning{
	z := &Zoning{
		Zones:  make([] * Zone, 0),
		Height: height,
		Width:  width,
		CursorZoneMap: make(map[int] * Zone),
		Input: input,
		Default: nil,
		Terminal: term,
	}
	go z.pipeToZone()
	return z
}

func (z * Zoning) zonesIntersect (a * Zone, b * Zone) bool {
	return (a.X < b.X + b.Width) && (a.X + a.Width > b.X) && (a.Y < b.Y + b.Height) && (a.Y + a.Height > b.Y)
}

func (z * Zoning) createZone (Y int, X int, Height int, Width int, CursorAllowed bool) (* Zone , error) {
	newZone := &Zone{
		Y:             Y,
		X:             X,
		Height:        Height,
		Width:         Width,
		CursorAllowed: CursorAllowed,
		Events: make(chan byte,1000),
		CursorMap: make(map[int] * Coord),
		Parent: z,
	}
	if Y + Height > z.Height || Y < 0 || X + Width > z.Width || X < 0 {
		return nil,errors.New("zone does not fit into terminal")
	}
	for _,zone := range z.Zones {
		if z.zonesIntersect(zone,newZone){
			//specify which zone
			return nil,errors.New("zone intersects with preexisting zone")
		}
	}
	z.Zones = append(z.Zones,newZone)
	return newZone,nil
}



func (z * Zoning) cursorEnterZone(zone * Zone,port int) error {
	if zone == nil {
		return errors.New("bad zone")
	}
	if !zone.CursorAllowed {
		return errors.New("cursor not allowed in zone")
	}
	z.CursorZoneMap[port] = zone
	return nil
}

func (z * Zoning) pipeToZone () {
	var e byte
	for true {
		e = <- z.Input.events
		z.waitUntilZoneLoaded(LOCAL_ID)
		z.CursorZoneMap[LOCAL_ID].Events <- e
	}
}

func (z * Zone) getRealCoords (port int) (int,int) {
	var loc *Coord
	if val, ok := z.CursorMap[port]; ok {
		loc = val
	}else {
		z.CursorMap[port] = &Coord{
			Row: 0,
			Col: 0,
		}
		loc = z.CursorMap[port]
	}
	return z.X + loc.Col,z.Y + loc.Row
}

func (z * Zone) getRealNewCoords (x int,y int) (int,int) {
	return z.X + x,z.Y +y
}

func (z * Zoning) getValidCursorMove (x int, y int,port int) (bool,int,int) {
	z.waitUntilZoneLoaded(port)
	zone := z.CursorZoneMap[port]
	realX,realY := zone.getRealNewCoords(x,y)
	forcedX,forcedY := realX,realY
	if realX < zone.X {
		forcedX = zone.X
	} else if x >= zone.X + zone.Width {
		forcedX = zone.X + zone.Width - 1
	}
	if y < zone.Y {
		forcedY = zone.Y
	} else if y >= zone.Y + zone.Height {
		forcedY = zone.Y + zone.Height - 1
	}
	if realX == forcedX && realY == forcedY {
		return true,x,y
	}
	return false,forcedX - zone.X,forcedY - zone.Y
}

func (z * Zoning) waitUntilZoneLoaded (port int) {

	for true{
		if _, ok := z.CursorZoneMap[port]; ok || z.Default != nil {
			if !ok {
				z.CursorZoneMap[port] = z.Default
			}
			return
		}else{
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (z * Zone) moveToCoord (x int, y int,port int) {
	z.CursorMap[port].Col = x
	z.CursorMap[port].Row = y
}

func (z * Zoning) moveInDirection (direction byte,port int) bool {
	z.waitUntilZoneLoaded(port)
	moveAccepted,newX,newY := false,0,0
	var loc * Coord
	zone := z.CursorZoneMap[port]
	if val, ok := zone.CursorMap[port]; ok {
		loc = val
	}else {
		zone.CursorMap[port] = &Coord{
			Row: 0,
			Col: 0,
		}
		loc = zone.CursorMap[port]
	}
	switch direction {
		case MOVE_LEFT:
			moveAccepted,newX,newY = z.getValidCursorMove(loc.Col - 1,loc.Row,port)
			break
		case MOVE_RIGHT:
			moveAccepted,newX,newY = z.getValidCursorMove(loc.Col + 1,loc.Row,port)
			break
		case MOVE_UP:
			moveAccepted,newX,newY = z.getValidCursorMove(loc.Col,loc.Row-1,port)
			break
		case MOVE_DOWN:
			moveAccepted,newX,newY = z.getValidCursorMove(loc.Col,loc.Row+1,port)
			break
	}
	loc.Col = newX
	loc.Row = newY
	return moveAccepted
}

func (z * Zoning) setDefaultZone (zone * Zone) {
	z.Default = zone
}

//Wrapper functions for zones, no direct terminal control

//func (z * Zone) sendPlaceCharFormat(char byte, row int, col int, format *Context, code byte){
//	nCol,nRow := z.getRealNewCoords(col,row)
//	z.Parent.Terminal.sendPlaceCharFormat(char,nRow,nCol,format,code)
//}
//
//func (z * Zone) sendCharAssociation(char byte,recorded * Recorded) {
//	z.Parent.Terminal.sendCharAssociation(char,recorded)
//}
//
//func (z * Zone) sendPrintStyleAtCoord(style * Context,row int,col int,text string){
//	nCol,nRow := z.getRealNewCoords(col,row)
//	z.Parent.Terminal.sendPrintStyleAtCoord(style,nRow,nCol,text)
//}
//
//func (z * Zone) sendPlaceCharAtCoord(char byte,row int,col int) {
//	nCol,nRow := z.getRealNewCoords(col,row)
//	z.Parent.Terminal.sendPlaceCharAtCoord(char,nRow,nCol)
//}
//
//func (z * Zone) sendPlaceCharAtCoordCondUndo(char byte,row int,col int,lastRow int,lastCol int,match byte,matchFg bool) {
//	nCol,nRow := z.getRealNewCoords(col,row)
//	nLastCol,nLastRow := z.getRealNewCoords(lastCol,lastRow)
//	z.Parent.Terminal.sendPlaceCharAtCoordCondUndo(char,nRow,nCol,nLastRow,nLastCol,match,matchFg)
//}
//
//func (z * Zone) sendUndoAtLocationConditional(row int,col int,match byte,matchFg bool) {
//	nCol,nRow := z.getRealNewCoords(col,row)
//	z.Parent.Terminal.sendUndoAtLocationConditional(nRow,nCol,match,matchFg)
//}
//
//func (z * Zone) sendRawFmtString(raw string,effectiveSize int, row int, col int) {
//	nCol,nRow := z.getRealNewCoords(col,row)
//	z.Parent.Terminal.sendRawFmtString(raw,effectiveSize,nRow,nCol)
//}

