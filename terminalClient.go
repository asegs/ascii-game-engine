package main

import (
	"fmt"
	"strings"
)

type TerminalClient struct {
	Window         *ClientWindow
	Row            int
	Col            int
	MultiMapLookup *MultiplexedLookup
	Height         int
	Width          int
	Stats          map[string]interface{}
}

//The char sequence to RESET the cursor style.
const RESET string = "\033[0m"

/**
A format string that contains a "%s" for the text to place in between style escape chars.
Can be compiled to update this format string from style modifiers unlimited times.
Format: The string that has the current "compiled" format, for example:
	"\033[38;2;200;12;181;m%s\033[0m"
Modifiers: The array of style modifiers, some seen in the example above, to apply when the Context is compiled.
*/
type Context struct {
	Format    string
	Modifiers []string
}

type Color struct {
	R int
	G int
	B int
}

type MultiplexedLookup struct {
	TopFgMap   map[byte]map[byte]*Context
	ColorFgMap map[byte]*Color
	ColorBgMap map[byte]*Color
}

func createMultiplexedLookup() *MultiplexedLookup {
	return &MultiplexedLookup{
		TopFgMap:   make(map[byte]map[byte]*Context),
		ColorFgMap: make(map[byte]*Color),
		ColorBgMap: make(map[byte]*Color),
	}
}

func (m *MultiplexedLookup) composeBasicContext(fg byte, bg byte) *Context {
	ctx := initContext()
	gotSomeStyle := false
	if color, ok := m.ColorFgMap[fg]; ok {
		ctx.addRgbStyleFg(color.R, color.G, color.B)
		gotSomeStyle = true
	}
	if color, ok := m.ColorBgMap[bg]; ok {
		ctx.addRgbStyleBg(color.R, color.G, color.B)
		gotSomeStyle = true
	}
	if !gotSomeStyle {
		ctx.addSimpleStyle(0)
	}
	ctx.compile()
	if _, ok := m.TopFgMap[fg]; !ok {
		m.TopFgMap[bg] = make(map[byte]*Context)
	}
	m.TopFgMap[fg][bg] = ctx
	return ctx
}

func (m *MultiplexedLookup) getContext(fg byte, bg byte) *Context {
	if innerLookup, ok := m.TopFgMap[fg]; ok {
		if ctx, stillOk := innerLookup[bg]; stillOk {
			return ctx
		}
		m.TopFgMap[fg] = make(map[byte]*Context)
	}
	return m.composeBasicContext(fg, bg)
}

func (m *MultiplexedLookup) addComposedStyle(fg byte, bg byte, context *Context) {
	m.TopFgMap[fg][bg] = context
}

func (m *MultiplexedLookup) addForegroundColor(char byte, r int, g int, b int) {
	m.ColorFgMap[char] = &Color{
		R: r,
		G: g,
		B: b,
	}
}

func (m *MultiplexedLookup) addBackgroundColor(char byte, r int, g int, b int) {
	m.ColorBgMap[char] = &Color{
		R: r,
		G: g,
		B: b,
	}
}

func terminalClientWithTerminalInput(height int, width int) (*TerminalClient, *NetworkedStdIn) {
	terminalClient := &TerminalClient{
		Window:         nil,
		Row:            0,
		Col:            0,
		MultiMapLookup: createMultiplexedLookup(),
		Height:         height,
		Width:          width,
		Stats:          make(map[string]interface{}),
	}
	input := initializeInput()
	return terminalClient, input
}

func (t *TerminalClient) Init(defaultFg byte, defaultBg byte, rows int, cols int) {
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			print(" ")
		}
		println()
	}
	t.Row = rows
	t.Col = 0
	t.moveTo(0, 0)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			t.DrawAt(defaultFg, defaultBg, row, col, true)
		}
	}
}

func (t *TerminalClient) DrawAt(fg byte, bg byte, row int, col int, bulk bool) {
	t.placeAt(fmt.Sprintf(t.MultiMapLookup.getContext(fg, bg).Format, string(fg)), row, col, 1)
}

func (t *TerminalClient) SetWindow(window *ClientWindow) {
	t.Window = window
}

/**
Given a Direction, or text character, moves the cursor n times in the given direction dir.
Changes the state of the terminal to match the new position.
*/
func (t *TerminalClient) moveCursor(n int, dir Direction) {
	fmt.Printf("\033[%d%c", n, dir)
	switch dir {
	case UP:
		t.Row -= n
		return
	case DOWN:
		t.Row += n
		return
	case LEFT:
		t.Col -= n
		return
	case RIGHT:
		t.Col += n
		return
	}
}

/**
Given a new pair of coordinates, calculates the difference in current position to the new position.
If the new position is out of bounds, sets it to the max value in that direction.
Uses moveCursor to update cursor position by going over the difference in the right direction.
*/
func (t *TerminalClient) moveTo(newRow int, newCol int) {
	if newRow >= t.Height {
		newRow = t.Height - 1
	}

	if newRow < 0 {
		newRow = 0
	}

	if newCol >= t.Width {
		newCol = t.Width - 1
	}

	if newCol < 0 {
		newCol = 0
	}
	if newRow-t.Row > 0 {
		t.moveCursor(newRow-t.Row, DOWN)
	} else if newRow-t.Row < 0 {
		t.moveCursor((newRow-t.Row)*-1, UP)
	}

	if newCol-t.Col > 0 {
		t.moveCursor(newCol-t.Col, RIGHT)
	} else if newCol-t.Col < 0 {
		t.moveCursor((newCol-t.Col)*-1, LEFT)
	}
}

/**
Takes a string of text, writes it to the end of the line and then cuts off the rest.
Updates the proper blocks of memory to track the current state.

One issue is the background code for each cell will never be anything other than ' '.
*/
func (t *TerminalClient) trimAndUpdateString(text string) string {
	over := len(text) + t.Col - (t.Window.Width - 1)
	if over > 0 {
		text = text[0 : len(text)-over]
	}
	for i := 0; i < len(text); i++ {
		t.Window.updateAtPos(t.Row, t.Col+i, text[i], true)
	}
	return text
}

/**
Writes a string to the terminal, shifts the cursor over back to the start after writing.
*/
func (t *TerminalClient) printRender(message string, txtLen int) {
	t.Col += txtLen
	print(message)
	t.moveCursor(txtLen, LEFT)
}

/**
Places a string at a given location using printRender.
Used after substituting message for format character and compiling Context.
*/
func (t *TerminalClient) placeAt(message string, row int, col int, txtLen int) {
	t.moveTo(row, col)
	t.printRender(message, txtLen)

}

/**
Creates a new Context.  Has no traits to begin with, just prints text.
*/
func initContext() *Context {
	return &Context{
		Format:    "%s",
		Modifiers: make([]string, 0),
	}
}

/**
Adds a basic style (preset color, flashing, etc.) to the modifiers array.
*/
func (ctx *Context) addSimpleStyle(styleConst int) *Context {
	ctx.Modifiers = append(ctx.Modifiers, fmt.Sprintf("%d;", styleConst))
	return ctx
}

/**
Adds a background color to the modifiers array.
*/
func (ctx *Context) addRgbStyleFg(r int, g int, b int) *Context {
	ctx.Modifiers = append(ctx.Modifiers, fmt.Sprintf("38;2;%d;%d;%d;", r, g, b))
	return ctx
}

/**
Adds a foreground color to the modifiers array.
*/
func (ctx *Context) addRgbStyleBg(r int, g int, b int) *Context {
	ctx.Modifiers = append(ctx.Modifiers, fmt.Sprintf("48;2;%d;%d;%d;", r, g, b))
	return ctx
}

/**
Removes the first RGB style from the modifiers array.
Often used for composing new styles.
Takes a boolean to remove the foreground style or background style.
*/
func (ctx *Context) removeRgbStyle(fg bool) *Context {
	key := "38;2"
	if !fg {
		key = "48;2"
	}
	for i, modifier := range ctx.Modifiers {
		if strings.Contains(modifier, key) {
			if i == len(ctx.Modifiers)-1 {
				ctx.Modifiers = ctx.Modifiers[0:i]
			} else {
				ctx.Modifiers = append(ctx.Modifiers[0:i], ctx.Modifiers[i+1:]...)
			}
			return ctx
		}
	}
	return ctx
}

/**
Makes a deep copy of a context.  Often used for composing new styles on the fly.
*/
func (ctx *Context) copyContext() *Context {
	newMods := make([]string, len(ctx.Modifiers))
	for i, modifier := range ctx.Modifiers {
		newMods[i] = modifier
	}
	return &Context{
		Format:    ctx.Format,
		Modifiers: newMods,
	}
}

/**
Returns the color modifier for a certain Context.
Either the foreground or background color based on the boolean.
If none exists, returns empty string.
*/
func (ctx *Context) getColorInfo(fg bool) string {
	key := "38;2"
	if !fg {
		key = "48;2"
	}
	for _, modifier := range ctx.Modifiers {
		if strings.Contains(modifier, key) {
			return modifier
		}
	}
	return ""
}

/**
Adds a raw style (typed out/string) to a given Context.
Often used to pass color info styles pulled directly from another Context.
*/
func (ctx *Context) addStyleRaw(modifier string) *Context {
	ctx.Modifiers = append(ctx.Modifiers, modifier)
	return ctx
}

/**
Builds the format string according to the current style modifiers.
First starts off with the proper escape code.
Next adds each style from the modifiers list.
Next deletes the final semicolon and replaces it with an 'm'.
Next adds an %s for use in printf.
Finally adds a RESET escape code to turn off the escape code for future writes.

Sets format equal to this string.
*/
func (ctx *Context) compile() *Context {
	newFmt := "\033["
	for _, modifier := range ctx.Modifiers {
		newFmt += modifier
	}
	newFmt = newFmt[0:len(newFmt)-1] + "m%s\033[0m"
	ctx.Format = newFmt
	return ctx
}

/**
Writes text with an associated style at a certain pair of coordinates.
*/
func (t *TerminalClient) writeStyleAt(style *Context, text string, row int, col int) {
	text = t.trimAndUpdateString(text)
	t.placeAt(fmt.Sprintf(style.Format, text), row, col, len(text))
}

/**
Writes just the styled text at a location, without loading into history.
Used for loading from saves.
*/
func (t *TerminalClient) writeStyleAtNoHistory(style *Context, text string, row int, col int) {
	t.placeAt(fmt.Sprintf(style.Format, text), row, col, len(text))
}

/**
Clears n tiles (replaces with nothing) at the given coordinates.  Doesn't update in memory, and not used.
*/
func (t *TerminalClient) wipeNTilesAt(tiles int, row int, col int) {
	t.moveTo(row, col)
	println(RESET)
	for tiles > 0 {
		fmt.Printf(" ")
		tiles--
	}
}

/**
Writes a certain character with no style at a particular location.
*/
//func (t * TerminalClient) placeCharRaw(char byte,row int,col int){
//	t.moveTo(row,col)
//	if t.Col >= t.Window.Width - 1{
//		return
//	}
//	t.Col ++
//	var character [1] byte
//	character[0] = char
//	fmt.Printf(t.Window.DataHistory[row][col].top().Format.Format,character)
//	t.moveCursor(1,LEFT)
//}

/**
Places a certain unassociated character with a provided format.
Requires the data to create a recording, including the format Context, as well as foreground and background values.
Adds the character to the history with a new Recorded value created from the aforementioned data.
*/
func (t *TerminalClient) placeCharFormat(char byte, row int, col int, format *Context, bgCode byte) {
	t.moveTo(row, col)
	if t.Col >= t.Window.Width-1 {
		return
	}
	t.Col++
	var character [1]byte
	character[0] = char
	fmt.Printf(format.Format, character)
	t.Window.updateAtPos(row, col, char, true)
	t.Window.updateAtPos(row, col, bgCode, false)
	t.moveCursor(1, LEFT)
}

/**
Composes and queues a function that places a character with a format at a specific position.
*/
func (t *TerminalClient) sendPlaceCharFormat(char byte, row int, col int, format *Context, code byte) {
	t.Window.CustomFeed <- func() {
		t.placeCharFormat(char, row, col, format, code)
	}
}

/**
Composes and queues a function prints text with a given style Context.
*/
func (t *TerminalClient) sendPrintStyleAtCoord(style *Context, row int, col int, text string) {
	t.Window.CustomFeed <- func() {
		t.writeStyleAt(style, text, row, col)
	}
}

/**
Only to be used on something where undoing it is not required.
*/
func (t *TerminalClient) sendRawFmtString(raw string, effectiveSize int, row int, col int) {
	t.Window.CustomFeed <- func() {
		t.moveTo(row, col)
		t.printRender(raw, effectiveSize)
	}
}

func repeatStringN(str string, length int) string {
	start := ""
	for i := 0; i < length; i++ {
		start += str
	}
	return start
}

func (t *TerminalClient) StatsToString() string {
	mapping := ""
	for stat, val := range t.Stats {
		mapping += fmt.Sprintf("%s: %v  ", stat, val)
	}
	return mapping
}

func (t *TerminalClient) RenderStats() {
	t.Window.CustomFeed <- func() {
		t.moveTo(t.Height, 0)
		t.moveCursor(1, DOWN)
		t.printRender(repeatStringN(" ", t.Width), t.Width)
		t.moveTo(t.Height, 0)
		t.moveCursor(1, DOWN)
		stats := t.StatsToString()
		if len(stats) > t.Width {
			stats = stats[0:t.Width]
		}
		t.printRender(stats, t.Width)
		t.moveCursor(t.Height, 0)
	}
}

func (t *TerminalClient) DrawStat(statName string, value interface{}) {
	t.Stats[statName] = value
	t.RenderStats()
}
