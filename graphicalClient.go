package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image"
	"image/color"
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

var resX = 2560
var resY = 1440

type CanvasWrapper struct {
	widget.BaseWidget
	image        *canvas.Raster
	zoom         int
	offset       *Coord
	rect         *image.Rectangle
	unmovedImage *image.RGBA
	viewingImage *image.RGBA
	stockImage   *image.RGBA
	charOffset   *Coord
	dragCounter  int
}

func (c *CanvasWrapper) Scrolled(scroll *fyne.ScrollEvent) {
	if scroll.Scrolled.DY > 0 {
		c.zoom++
	} else if c.zoom > 1 {
		c.zoom--
	}
	c.image.Resize(fyne.NewSize(float32(c.zoom*resX), float32(c.zoom*resY)))
	c.Refresh()
}

func (c *CanvasWrapper) Dragged(event *fyne.DragEvent) {
	divisor := 1
	if c.zoom > 1 {
		divisor = c.zoom / 2
	}
	c.offset.Col += -1 * int(event.Dragged.DX) / divisor
	c.offset.Row += -1 * int(event.Dragged.DY) / divisor
	if c.dragCounter%100 == 0 {
		draw.Draw(c.viewingImage, *c.rect, c.stockImage, image.Point{0, 0}, draw.Over)
		draw.Draw(c.viewingImage, *c.rect, c.unmovedImage, image.Point{
			X: c.offset.Col,
			Y: c.offset.Row,
		}, draw.Over)
		c.Refresh()
	}
	c.dragCounter++
}

func CloneImage(src image.RGBA) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, &src, b.Min, draw.Src)
	return dst
}

func (c *CanvasWrapper) DragEnd() {
	draw.Draw(c.viewingImage, *c.rect, c.stockImage, image.Point{0, 0}, draw.Over)
	draw.Draw(c.viewingImage, *c.rect, c.unmovedImage, image.Point{
		X: c.offset.Col,
		Y: c.offset.Row,
	}, draw.Over)
	c.Refresh()
	c.dragCounter = 0
}

func (c *CanvasWrapper) CreateRenderer() fyne.WidgetRenderer {
	return &rasterWidgetRender{raster: c}
}

type rasterWidgetRender struct {
	raster *CanvasWrapper
}

func (r *rasterWidgetRender) Layout(size fyne.Size) {
	r.raster.image.Resize(size)
}

func (r *rasterWidgetRender) MinSize() fyne.Size {
	return fyne.Size{1, 1}
}

func (r *rasterWidgetRender) Refresh() {

	canvas.Refresh(r.raster)
}

func (r *rasterWidgetRender) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *rasterWidgetRender) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster.image}
}

func (r *rasterWidgetRender) Destroy() {
}

type GraphicalClient struct {
	Window         *ClientWindow
	Sprites        *MultiplexedSpriteLookup
	WindowName     string
	InputPipe      *chan byte
	ViewingImage   *image.RGBA
	Rect           *image.Rectangle
	Canvas         *CanvasWrapper
	StdImageWidth  int
	StdImageHeight int
	RenderWindow   fyne.Window
	IsIsometric    bool
	Stats          map[string]interface{}
	Revealed       bool
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
		Stats:          make(map[string]interface{}),
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

func (i *GraphicalClient) Init(defaultFg byte, defaultBg byte, rows int, cols int) {
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

	stockImage := image.NewRGBA(r)
	black := color.RGBA{48, 48, 48, 0xff}
	for x := 0; x < resX; x++ {
		for y := 0; y < resY; y++ {
			stockImage.Set(x, y, black)
		}
	}
	widgetWrapper := CanvasWrapper{
		image: canvasToWrite,
		zoom:  1,
		offset: &Coord{
			Row: 0,
			Col: 0,
		},
		charOffset:   &Coord{-650, 0},
		rect:         &r,
		viewingImage: rgbaImage,
		stockImage:   stockImage,
		unmovedImage: CloneImage(*rgbaImage),
	}
	widgetWrapper.ExtendBaseWidget(&widgetWrapper)
	imageWindow.SetContent(&widgetWrapper)
	imageWindow.Resize(fyne.NewSize(float32(resX), float32(resY)))
	i.ViewingImage = rgbaImage
	i.Rect = &r
	i.Canvas = &widgetWrapper

	for col := cols; col >= 0; col-- {
		for row := 0; row < rows; row++ {
			i.DrawAt(defaultFg, defaultBg, row, col, true)
		}
	}
	i.RenderWindow = imageWindow
	imageWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		*i.InputPipe <- keycodeMap(event.Name)
	})
}

func (i *GraphicalClient) DrawAt(fg byte, bg byte, row int, col int, bulk bool) {
	bgSprite := i.Sprites.getBgSprite(bg)
	fgSprite := i.Sprites.getFgSprite(fg)
	if bgSprite != nil {
		i.drawAtCoord(i.Canvas.unmovedImage, bgSprite, col, row, i.Rect)
	}
	if fgSprite != nil {
		i.drawAtCoord(i.Canvas.unmovedImage, fgSprite, col, row, i.Rect)
	}
	if i.Revealed && !bulk {
		draw.Draw(i.Canvas.viewingImage, *i.Canvas.rect, i.Canvas.unmovedImage, image.Point{
			X: i.Canvas.offset.Col,
			Y: i.Canvas.offset.Row,
		}, draw.Over)
	}
	canvas.Refresh(i.Canvas)
}

func (i *GraphicalClient) permutePoint(point image.Point) image.Point {
	if !i.IsIsometric {
		return point
	}

	return image.Point{
		X: -1*(point.X*i.StdImageWidth) + (point.X-point.Y)*int(float64(i.StdImageWidth)*0.5),
		Y: -1*(point.Y*i.StdImageHeight) + (point.X+point.Y)*int(float64(i.StdImageHeight)*0.5) - 650,
	}
}

func offsetForImage(point image.Point, tile image.Image) image.Point {
	return image.Point{X: point.X, Y: point.Y + tile.Bounds().Dy()}
}

func (i *GraphicalClient) drawAtCoord(onto *image.RGBA, from *image.Image, x int, y int, bounds *image.Rectangle) {
	draw.Draw(onto, *bounds, *from, offsetForImage(i.permutePoint(image.Point{x, y + 1}), *from), draw.Over)
}

func (i *GraphicalClient) show() {
	i.Revealed = true
	i.RenderWindow.ShowAndRun()
}

func (i *GraphicalClient) RenderStats() {

}

func (i *GraphicalClient) DrawStat(statName string, value interface{}) {

}
