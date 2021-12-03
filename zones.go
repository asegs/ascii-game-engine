package main

import (
	"errors"
	"time"
)

type Zoning struct {
	Zones [] * Zone
	Height int
	Width int
	CursorZone * Zone
	Input * StdIn
}

type Zone struct {
	Y int
	X int
	Height int
	Width int
	CursorAllowed bool
	Events chan byte
	CursorY int
	CursorX int
}

func initZones (height int,width int, input * StdIn) * Zoning{
	z := &Zoning{
		Zones:  make([] * Zone, 0),
		Height: height,
		Width:  width,
		CursorZone: nil,
		Input: input,
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
		CursorY: 0,
		CursorX: 0,
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

func (z * Zoning) cursorEnterZone(zone * Zone) error {
	if zone == nil {
		return errors.New("bad zone")
	}
	if !zone.CursorAllowed {
		return errors.New("cursor not allowed in zone")
	}
	z.CursorZone = zone
	return nil
}

func (z * Zoning) pipeToZone () {
	z.waitUntilZoneLoaded()
	for true {
		z.CursorZone.Events <- <- z.Input.events
	}
}

func (z * Zone) getRealCoords () (int,int) {
	return z.X + z.CursorX,z.Y + z.CursorY
}

func (z * Zone) getRealNewCoords (x int,y int) (int,int) {
	return z.X + x,z.Y +y
}

func (z * Zoning) getValidCursorMove (x int, y int) (bool,int,int) {
	z.waitUntilZoneLoaded()
	realX,realY := z.CursorZone.getRealNewCoords(x,y)
	forcedX,forcedY := realX,realY
	if realX < z.CursorZone.X {
		forcedX = z.CursorZone.X
	} else if x >= z.CursorZone.X + z.CursorZone.Width {
		forcedX = z.CursorZone.X + z.CursorZone.Width - 1
	}
	if y < z.CursorZone.Y {
		forcedY = z.CursorZone.Y
	} else if y >= z.CursorZone.Y + z.CursorZone.Height {
		forcedY = z.CursorZone.Y + z.CursorZone.Height - 1
	}
	if realX == forcedX && realY == forcedY {
		return true,x,y
	}
	return false,forcedX - z.CursorZone.X,forcedY - z.CursorZone.Y
}

func (z * Zoning) waitUntilZoneLoaded () {
	for z.CursorZone == nil {
		time.Sleep(10 * time.Millisecond)
	}
}

func (z * Zone) moveToCoord (x int, y int) {
	z.CursorX = x
	z.CursorY = y
}

func (z * Zoning) moveInDirection (direction byte) bool {
	moveAccepted,newX,newY := false,0,0
	switch direction {
	case MOVE_LEFT:
		moveAccepted,newX,newY = z.getValidCursorMove(z.CursorZone.CursorX - 1,z.CursorZone.CursorY)
		break
	case MOVE_RIGHT:
		moveAccepted,newX,newY = z.getValidCursorMove(z.CursorZone.CursorX + 1,z.CursorZone.CursorY)
		break
	case MOVE_UP:
		moveAccepted,newX,newY = z.getValidCursorMove(z.CursorZone.CursorX,z.CursorZone.CursorY-1)
		break
	case MOVE_DOWN:
		moveAccepted,newX,newY = z.getValidCursorMove(z.CursorZone.CursorX,z.CursorZone.CursorY+1)
		break
	}
	z.CursorZone.CursorX = newX
	z.CursorZone.CursorY = newY
	return moveAccepted
}
