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

var keycodeMapping = map[fyne.KeyName]byte{
	"Right": MOVE_RIGHT,
	"Left":  MOVE_LEFT,
	"Up":    MOVE_UP,
	"Down":  MOVE_DOWN,
}

func keycodeMap(pressed fyne.KeyName) byte {
	if key, ok := keycodeMapping[pressed]; ok {
		return key
	}
	return pressed[0]
}

var resX = 1920
var resY = 1080

type GraphicalClient struct {
	Window         *ClientWindow
	Sprites        *MultiplexedSpriteLookup
	WindowName     string
	InputPipe      *chan byte
	ViewingImage   *image.RGBA
	Rect           *image.Rectangle
	Canvas         *canvas.Raster
	StdImageWidth  int
	StdImageHeight int
	RenderWindow   fyne.Window
	IsIsometric    bool
}

func graphicalClientWithInput(windowName string, spriteWidth int, spriteHeight int, isometric bool) (*GraphicalClient, *NetworkedStdIn) {
	pipe := make(chan byte, MAX_MESSAGES)
	gCl := &GraphicalClient{
		Window:         nil,
		Sprites:        createSpriteLookup(),
		WindowName:     windowName,
		InputPipe:      &pipe,
		ViewingImage:   nil,
		Rect:           nil,
		Canvas:         nil,
		StdImageWidth:  spriteWidth,
		StdImageHeight: spriteHeight,
		IsIsometric:    isometric,
	}
	return gCl, &NetworkedStdIn{events: pipe}
}

type MultiplexedSpriteLookup struct {
	FgSprites map[byte]*image.Image
	BgSprites map[byte]*image.Image
}

func createSpriteLookup() *MultiplexedSpriteLookup {
	return &MultiplexedSpriteLookup{
		FgSprites: make(map[byte]*image.Image),
		BgSprites: make(map[byte]*image.Image),
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

func (i *GraphicalClient) addFgSprite(char byte, filepath string) {
	img, err := getImageFromFilePath(filepath)
	if err != nil {
		LogString("No image found at: " + filepath)
		LogString(err.Error())
		return
	}
	i.Sprites.FgSprites[char] = &img
}

func (i *GraphicalClient) addBgSprite(char byte, filepath string) {
	img, err := getImageFromFilePath(filepath)
	if err != nil {
		LogString("No image found at: " + filepath)
		LogString(err.Error())
		return
	}
	i.Sprites.BgSprites[char] = &img
}

func (m *MultiplexedSpriteLookup) getFgSprite(char byte) *image.Image {
	if sprite, ok := m.FgSprites[char]; ok {
		return sprite
	}
	return nil
}

func (m *MultiplexedSpriteLookup) getBgSprite(char byte) *image.Image {
	if sprite, ok := m.BgSprites[char]; ok {
		return sprite
	}
	return nil
}

func (i *GraphicalClient) SetWindow(window *ClientWindow) {
	i.Window = window
}

func (i *GraphicalClient) Init(pair *TilePair, rows int, cols int) {
	topLeft := image.Point{0, 0}
	bottomRight := image.Point{resX, resY}

	rgbaImage := image.NewRGBA(image.Rectangle{topLeft, bottomRight})

	mainApp := app.New()
	imageWindow := mainApp.NewWindow(i.WindowName)

	r := image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{resX, resY},
	}
	canvasToWrite := canvas.NewRasterFromImage(rgbaImage)
	imageWindow.SetContent(canvasToWrite)
	imageWindow.Resize(fyne.NewSize(float32(resX), float32(resY)))
	i.ViewingImage = rgbaImage
	i.Rect = &r
	i.Canvas = canvasToWrite

	for col := cols; col >= 0; col-- {
		for row := 0; row < rows; row++ {
			i.DrawAt(pair, row, col)
		}
	}
	i.RenderWindow = imageWindow
	imageWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		*i.InputPipe <- keycodeMap(event.Name)
	})
}

func (i *GraphicalClient) DrawAt(pair *TilePair, row int, col int) {
	bgSprite := i.Sprites.getBgSprite(pair.BackgroundCode)
	fgSprite := i.Sprites.getFgSprite(pair.ShownSymbol)
	if bgSprite != nil {
		i.drawAtCoord(i.ViewingImage, bgSprite, col, row, i.Rect)
	}
	if fgSprite != nil {
		i.drawAtCoord(i.ViewingImage, fgSprite, col, row, i.Rect)
	}
	canvas.Refresh(i.Canvas)
}

func (i *GraphicalClient) permutePoint(point image.Point) image.Point {
	offsetX := -100
	offsetY := -500
	if i.IsIsometric {
		return point
	}

	return image.Point{
		X: -1*(point.X*i.StdImageWidth) + (point.X-point.Y)*int(float64(i.StdImageWidth)*0.5) + offsetX,
		Y: -1*(point.Y*i.StdImageHeight) + (point.X+point.Y)*int(float64(i.StdImageHeight)*0.5) + offsetY,
	}
}

func offsetForImage(point image.Point, tile image.Image) image.Point {
	return image.Point{X: point.X, Y: point.Y + tile.Bounds().Dy()}
}

func (i *GraphicalClient) drawAtCoord(onto *image.RGBA, from *image.Image, x int, y int, bounds *image.Rectangle) {
	draw.Draw(onto, *bounds, *from, offsetForImage(i.permutePoint(image.Point{x, y + 1}), *from), draw.Over)
}

func (i *GraphicalClient) show() {
	i.RenderWindow.ShowAndRun()
}
