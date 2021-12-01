package main

type Zoning struct {
	Zones [] * Zone
	Height int
	Width int
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

func (z * Zoning) createObject (Y int, X int, Height int, Width int, CursorAllowed bool) error {
	//check if zone mapping is good, if shape is within other zone
	return nil
}

