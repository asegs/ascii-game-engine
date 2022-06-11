package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"image"
	"image/draw"
	"image/png"
	"os"
)

var resX = 1920
var resY = 1080

type IsometricClient struct {
	Window * ClientWindow
	Sprites * MultiplexedSpriteLookup
	WindowName string
	InputPipe * chan byte
	ViewingImage *image.RGBA
	Rect * image.Rectangle
	Canvas * canvas.Raster
	StdImageWidth int
	StdImageHeight int
}

func isometricClientWithInput (windowName string, spriteWidth int, spriteHeight int) (* IsometricClient, * NetworkedStdIn) {
	pipe := make(chan byte, MAX_MESSAGES)
	iso := &IsometricClient{
		Window:         nil,
		Sprites:        createSpriteLookup(),
		WindowName:     windowName,
		InputPipe:      &pipe,
		ViewingImage:   nil,
		Rect:           nil,
		Canvas: 		nil,
		StdImageWidth:  spriteWidth,
		StdImageHeight: spriteHeight,
	}
	return iso, &NetworkedStdIn{events: pipe}
}


type MultiplexedSpriteLookup struct {
	FgSprites map[byte] * image.Image
	BgSprites map[byte] * image.Image
 }

 func createSpriteLookup () * MultiplexedSpriteLookup {
	 return &MultiplexedSpriteLookup{
		 FgSprites: make(map[byte] * image.Image),
		 BgSprites: make(map[byte] * image.Image),
	 }
 }

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	return img, err
}

func (i * IsometricClient) addFgSprite (char byte, filepath string) {
	img,err := getImageFromFilePath(filepath)
	if err != nil {
		LogString("No image found at: " + filepath)
		LogString(err.Error())
		return
	}
	i.Sprites.FgSprites[char] = &img
}

func (i * IsometricClient) addBgSprite (char byte, filepath string) {
	img,err := getImageFromFilePath(filepath)
	if err != nil {
		LogString("No image found at: " + filepath)
		LogString(err.Error())
		return
	}
	i.Sprites.BgSprites[char] = &img
}

func (m * MultiplexedSpriteLookup) getFgSprite (char byte) * image.Image {
	if sprite,ok := m.FgSprites[char] ; ok {
		return sprite
	}
	return nil
}

func (m * MultiplexedSpriteLookup) getBgSprite (char byte) * image.Image {
	if sprite,ok := m.BgSprites[char] ; ok {
		return sprite
	}
	return nil
}


 func (i * IsometricClient) SetWindow (window * ClientWindow) {
	 i.Window = window
 }

 func (i * IsometricClient) Init (pair * TilePair, rows int, cols int) {
	 topLeft := image.Point{0, 0}
	 bottomRight := image.Point{resX, resY}

	 rgbaImage := image.NewRGBA(image.Rectangle{topLeft, bottomRight})

	 mainApp := app.New()
	 imageWindow := mainApp.NewWindow(i.WindowName)

	 r := image.Rectangle{
		 Min: image.Point{0,0},
		 Max: image.Point{resX,resY},
	 }
	 canvasToWrite := canvas.NewRasterFromImage(rgbaImage)
	 imageWindow.SetContent(canvasToWrite)
	 imageWindow.Resize(fyne.NewSize(float32(resX), float32(resY)))
	 i.ViewingImage = rgbaImage
	 i.Rect = &r
	 i.Canvas = canvasToWrite

	 for col := cols ; col >= 0 ; col -- {
		 for row := 0 ; row < rows ; row ++ {
			 i.DrawAt(pair, row, col)
		 }
	 }

	 go imageWindow.ShowAndRun()
 }

 func (i * IsometricClient) DrawAt (pair * TilePair, row int, col int) {
	 bgSprite := i.Sprites.getBgSprite(pair.BackgroundCode)
	 fgSprite := i.Sprites.getFgSprite(pair.ShownSymbol)
	 if bgSprite != nil {
		 i.drawAtCoord(i.ViewingImage,bgSprite,col, row, i.Rect)
	 }
	 if fgSprite != nil {
		 i.drawAtCoord(i.ViewingImage,fgSprite,col, row, i.Rect)
	 }
	 canvas.Refresh(i.Canvas)
 }

func (i * IsometricClient) permutePoint (point image.Point) image.Point {
	offsetX := -100
	offsetY := -200

	return image.Point{
		X: -1 * (point.X * i.StdImageWidth) + (point.X - point.Y) * int(float64(i.StdImageWidth) * 0.5) + offsetX,
		Y: -1 * (point.Y * i.StdImageHeight) + (point.X + point.Y) * int(float64(i.StdImageHeight) * 0.5 ) + offsetY,
	}
}

func  (i * IsometricClient) drawAtCoord (onto  *image.RGBA, from * image.Image, x int, y int, bounds * image.Rectangle) {
	draw.Draw(onto,* bounds, * from,i.permutePoint(image.Point{x,y}),draw.Over)
}





