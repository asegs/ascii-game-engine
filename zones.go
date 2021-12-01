package main

import "errors"

type Zoning struct {
	Zones [] * Zone
	Height int
	Width int
	CursorY int
	CursorX int
}

type Zone struct {
	Y int
	X int
	Height int
	Width int
	CursorAllowed bool
}

func initZones (height int,width int) * Zoning{
	return &Zoning{
		Zones:  make([] * Zone, 0),
		Height: height,
		Width:  width,
	}
}

func (z * Zoning) zonesIntersect (a * Zone, b * Zone) bool {
	return (a.X <= b.X + b.Width) && (a.X + a.Width >= b.X) && (a.Y <= b.Y + b.Height) && (a.Y + a.Height >= b.Y)
}

func (z * Zoning) createObject (Y int, X int, Height int, Width int, CursorAllowed bool) error {
	newZone := &Zone{
		Y:             Y,
		X:             X,
		Height:        Height,
		Width:         Width,
		CursorAllowed: CursorAllowed,
	}
	for _,zone := range z.Zones {
		if z.zonesIntersect(zone,newZone){
			//specify which zone
			return errors.New("zone intersects with preexisting zone")
		}
	}
	z.Zones = append(z.Zones,newZone)
	return nil
}

