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
	return z.Y + z.CursorY,z.X + z.CursorX
}

func (z * Zoning) getValidCursorMove (x int, y int) (bool,int,int) {
	z.waitUntilZoneLoaded()
	realX := x
	realY := y
	if x < z.CursorZone.X {
		realX = z.CursorZone.X
	} else if x >= z.CursorZone.X + z.CursorZone.Width {
		realX = z.CursorZone.X + z.CursorZone.Width - 1
	}
	if y < z.CursorZone.Y {
		realY = z.CursorZone.Y
	} else if y >= z.CursorZone.Y + z.CursorZone.Height {
		realY = z.CursorZone.Y + z.CursorZone.Height - 1
	}
	if realX == x && realY == y {
		return true,x,y
	}
	return false,realX,realY
}

func (z * Zoning) waitUntilZoneLoaded () {
	for z.CursorZone == nil {
		time.Sleep(10 * time.Millisecond)
	}
}
