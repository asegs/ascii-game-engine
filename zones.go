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
	return &Zoning{
		Zones:  make([] * Zone, 0),
		Height: height,
		Width:  width,
		CursorZone: nil,
		Input: input,
	}
}

func (z * Zoning) zonesIntersect (a * Zone, b * Zone) bool {
	return (a.X <= b.X + b.Width) && (a.X + a.Width >= b.X) && (a.Y <= b.Y + b.Height) && (a.Y + a.Height >= b.Y)
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
	for true {
		for z.CursorZone == nil {
			time.Sleep(10 * time.Millisecond)
		}
		z.CursorZone.Events <- <- z.Input.events
	}
}

