package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	// Enable loading of PNG files
	"image/color"
	_ "image/png"

	"founderio.net/eljam/elcar"
	"founderio.net/eljam/paths"
	"github.com/BurntSushi/toml"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const spriteFolder = "resources/sprites"

func main() {
	pixelgl.Run(run)
}

func loadPicture(filename string) (*pixel.PictureData, error) {
	path := filepath.Join(spriteFolder, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

var (
	fontAtlas *text.Atlas

	carHoodSprite     *pixel.Sprite
	componentBGSprite *pixel.Sprite

	componentEmpty   *pixel.Sprite
	componentUnknown *pixel.Sprite

	componentSprites map[string]*pixel.Sprite
	propSprites      map[string]*pixel.Sprite

	spritePinIn      *pixel.Sprite
	spritePinOut     *pixel.Sprite
	spriteChipPort   *pixel.Sprite
	spriteSensorPort *pixel.Sprite
)

var (
	componentList []string
)

var (
	car   *elcar.Car
	world *elcar.World
)

var (
	hoodScale float64 = 3

	connectingFromState int
	connectingFromID    int
	connectingFromPort  int

	selectingComponent string

	dragRectStartPoint pixel.Vec

	menu = MenuClosed
)

const (
	NotConnecting int = iota
	ConnectingFromInput
	ConnectingFromOutput
)

const (
	MenuClosed int = iota
	MenuHood
	MenuMain
	MenuSave
	MenuLoad
)

func run() {

	_, err := toml.DecodeFile(filepath.Join("resources", "world.toml"), &world)
	if err != nil {
		panic(err)
	}
	// Safeguard against wacky maths
	if world.Scale <= 0.1 {
		world.Scale = 3
	}

	cfg := pixelgl.WindowConfig{
		Title:  "Electronics Jam",
		Bounds: pixel.R(0, 0, world.Size.X*world.Scale, world.Size.Y*world.Scale),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	_, err = toml.DecodeFile(filepath.Join("resources", "definitions.toml"), &elcar.Definitions)
	if err != nil {
		panic(err)
	}

	_, err = toml.DecodeFile(filepath.Join("resources", "sprites.toml"), &elcar.SpriteDefinitions)
	if err != nil {
		panic(err)
	}

	for typeName, def := range elcar.Definitions.Components {
		if def.Usable {
			componentList = append(componentList, typeName)
		}
	}
	sort.Strings(componentList)

	fontAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// Common + UI Sprites
	carPic, err := loadPicture("car.png")
	if err != nil {
		panic(err)
	}
	carSprite := pixel.NewSprite(carPic, carPic.Bounds())

	carHoodPic, err := loadPicture("car_circuits.png")
	if err != nil {
		panic(err)
	}
	carHoodSprite = pixel.NewSprite(carHoodPic, carHoodPic.Bounds())

	componentBGPic, err := loadPicture("component_bg.png")
	if err != nil {
		panic(err)
	}
	componentBGSprite = pixel.NewSprite(componentBGPic, componentBGPic.Bounds())

	worldPic, err := loadPicture(world.BackgroundSprite)
	if err != nil {
		panic(err)
	}
	worldSprite := pixel.NewSprite(worldPic, worldPic.Bounds())

	// Prop Sprites
	propSprites = make(map[string]*pixel.Sprite)
	propSpriteSheet, err := loadPicture("props.png")
	if err != nil {
		panic(err)
	}
	for name, def := range elcar.SpriteDefinitions.Props {
		propSprites[name] = pixel.NewSprite(propSpriteSheet, pixel.Rect{
			Min: def.Start,
			Max: def.Start.Add(def.Size),
		})
	}

	// Component Sprites
	componentSprites = make(map[string]*pixel.Sprite)
	componentSpriteSheet, err := loadPicture("components.png")
	if err != nil {
		panic(err)
	}
	for name, def := range elcar.SpriteDefinitions.Components {
		componentSprites[name] = pixel.NewSprite(componentSpriteSheet, pixel.Rect{
			Min: def.Start,
			Max: def.Start.Add(def.Size),
		})
	}

	// PCB Sprites
	pcbSpriteSheet, err := loadPicture("pcb.png")
	if err != nil {
		panic(err)
	}
	spritePinIn = pixel.NewSprite(pcbSpriteSheet, pixel.R(0, 0, 6, 6))
	spritePinOut = pixel.NewSprite(pcbSpriteSheet, pixel.R(0, 6, 6, 6+6))
	componentEmpty = pixel.NewSprite(pcbSpriteSheet, pixel.R(6, 0, 6+14, 18))
	componentUnknown = pixel.NewSprite(componentSpriteSheet, pixel.R(20, 0, 20+14, 18))
	spriteSensorPort = pixel.NewSprite(pcbSpriteSheet, pixel.R(34, 0, 34+14, 18))
	spriteChipPort = pixel.NewSprite(pcbSpriteSheet, pixel.R(48, 0, 48+14, 18))

	// Initialize a default car
	car = &elcar.Car{
		Position: pixel.V(210, 204),
		Rotation: math.Pi,
		Speed:    0,
	}

	for idx, port := range elcar.Definitions.Ports {
		if port.Prefill != "" {
			car.AddComponent(idx, port.Prefill)
		}
	}

	imd := imdraw.New(nil)

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.JustPressed(pixelgl.KeyTab) {
			if menu == MenuClosed {
				menu = MenuHood
			} else if menu == MenuHood {
				menu = MenuClosed
			}
		}
		if win.JustPressed(pixelgl.KeyEscape) {
			switch menu {
			case MenuLoad:
				fallthrough
			case MenuSave:
				fallthrough
			case MenuClosed:
				menu = MenuMain

			case MenuMain:
				fallthrough
			case MenuHood:
				menu = MenuClosed

			}
		}

		car.Update(dt, worldPic, world)

		win.Clear(colornames.Gainsboro)

		worldSprite.Draw(win, pixel.IM.Scaled(pixel.ZV, world.Scale).Moved(win.Bounds().Center()))

		// World Walls
		for _, o := range world.Walls {
			imd.Clear()
			imd.Color = colornames.Gray
			imd.Push(o.Pos.Scaled(world.Scale), o.Pos.Add(o.Size).Scaled(world.Scale))
			imd.Rectangle(2)
			imd.Draw(win)
		}

		// Props
		for _, o := range world.Props {
			sprite, ok := propSprites[o.Name]
			if !ok {
				sprite = spriteChipPort
			}

			sprite.Draw(win, pixel.IM.Moved(sprite.Frame().Size().Scaled(0.5)).Moved(o.Pos).Scaled(pixel.ZV, world.Scale))

			imd.Clear()
			imd.Color = colornames.Lightblue
			imd.Push(o.Pos.Scaled(world.Scale), o.Pos.Add(sprite.Frame().Size()).Scaled(world.Scale))
			imd.Rectangle(2)
			imd.Draw(win)
		}

		mat := pixel.IM.Rotated(pixel.ZV, -car.Rotation)
		mat = mat.Moved(car.Position)
		mat = mat.Scaled(pixel.ZV, world.Scale)
		carSprite.Draw(win, mat)

		for _, debug := range car.DebugPoints {
			imd.Clear()

			imd.Color = colornames.Magenta
			imd.EndShape = imdraw.SharpEndShape
			imd.Push(car.Position.Scaled(world.Scale), debug.Scaled(world.Scale))
			imd.Line(1)
			imd.Draw(win)
		}

		for _, debug := range car.DebugLines {
			imd.Clear()

			imd.Color = colornames.Magenta
			imd.EndShape = imdraw.SharpEndShape
			imd.Push(debug.A.Scaled(world.Scale), debug.B.Scaled(world.Scale))
			imd.Line(1)
			imd.Draw(win)
		}

		switch menu {
		case MenuClosed:
			runLevelEditTools(win, dt)
			if drawMenuButton(win, fontAtlas, "Open Hood [Tab]", pixel.R(0, 0, 350, 50)) {
				menu = MenuHood
			}

		case MenuHood:
			drawHood(win, dt)
			drawComponentSelector(win, dt)
			if drawMenuButton(win, fontAtlas, "Close Hood [Tab]", pixel.R(0, 256*hoodScale, 350, 256*hoodScale+50)) {
				menu = MenuHood
			}

		case MenuMain:
			drawMainMenu(win, dt)

		case MenuLoad:
			drawLoadMenu(win, dt)

		case MenuSave:
			drawSaveMenu(win, dt)
		}

		win.Update()
	}
}

func rectAround(center, size pixel.Vec) pixel.Rect {
	half := size.Scaled(0.5)
	return pixel.Rect{
		Min: center.Sub(half),
		Max: center.Add(half),
	}
}

func drawMainMenu(win *pixelgl.Window, dt float64) {
	buttonSize := pixel.V(450, 50)

	if drawMenuButton(win, fontAtlas, "Save Car", rectAround(win.Bounds().Center().Add(pixel.V(0, 150)), buttonSize)) {
		menu = MenuSave
		loadSaveEntries()
	}
	if drawMenuButton(win, fontAtlas, "Load Car", rectAround(win.Bounds().Center().Add(pixel.V(0, 50)), buttonSize)) {
		menu = MenuLoad
		loadSaveEntries()
	}
	if drawMenuButton(win, fontAtlas, "Exit", rectAround(win.Bounds().Center().Add(pixel.V(0, -150)), buttonSize)) {
		win.SetClosed(true)
	}
}

func drawLoadMenu(win *pixelgl.Window, dt float64) {
	buttonSize := pixel.V(450, 50)

	drawMenuButton(win, fontAtlas, "Load Car", rectAround(win.Bounds().Center().Add(pixel.V(0, 150)), buttonSize))
	if drawMenuButton(win, fontAtlas, "<", rectAround(win.Bounds().Center().Add(pixel.V(-275, 150)), pixel.V(50, 50))) {
		menu = MenuMain
	}

	if saveLoadError != "" {
		drawError(win, fontAtlas, saveLoadError, win.Bounds().Center().Add(pixel.V(0, 220)))
	}

	for i, entry := range saveEntries {
		buttonRect := rectAround(win.Bounds().Center().Add(pixel.V(0, float64(50-i*100))), buttonSize)
		var buttonText string
		if entry.Used {
			buttonText = entry.Created.Format("2006-01-02 15:04:05")
		} else {
			buttonText = "[Empty]"
		}

		if drawMenuButton(win, fontAtlas, buttonText, buttonRect) {
			err := loadCar(i)
			if err == nil {
				saveLoadError = ""
				menu = MenuHood
			} else {
				saveLoadError = err.Error()
			}
		}
	}
}

func drawSaveMenu(win *pixelgl.Window, dt float64) {
	buttonSize := pixel.V(450, 50)

	drawMenuButton(win, fontAtlas, "Save Car", rectAround(win.Bounds().Center().Add(pixel.V(0, 150)), buttonSize))
	if drawMenuButton(win, fontAtlas, "<", rectAround(win.Bounds().Center().Add(pixel.V(-275, 150)), pixel.V(50, 50))) {
		menu = MenuMain
	}

	if saveLoadError != "" {
		drawError(win, fontAtlas, saveLoadError, win.Bounds().Center().Add(pixel.V(0, 220)))
	}

	for i, entry := range saveEntries {
		buttonRect := rectAround(win.Bounds().Center().Add(pixel.V(0, float64(50-i*100))), buttonSize)
		var buttonText string
		if entry.Used {
			buttonText = entry.Created.Format("2006-01-02 15:04:05")
		} else {
			buttonText = "[Empty]"
		}

		if drawMenuButton(win, fontAtlas, buttonText, buttonRect) {
			err := saveCar(i)
			if err == nil {
				loadSaveEntries()
				saveLoadError = ""
			} else {
				saveLoadError = err.Error()
			}
		}
	}
}

func loadCar(slot int) error {
	filename := getSaveFileName(slot)
	return car.Load(filename)
}

func saveCar(slot int) error {
	filename := getSaveFileName(slot)
	return car.Save(filename)
}

var saveEntries [5]SaveEntry
var saveLoadError string

type SaveEntry struct {
	Used    bool
	Created time.Time
}

func getSaveFileName(slot int) string {
	return filepath.Join(paths.GetDataPath(), fmt.Sprintf("save_%d.toml", slot))
}

func loadSaveEntries() {
	saveLoadError = ""
	for i := range saveEntries {
		filename := getSaveFileName(i)

		info, err := os.Stat(filename)
		if err == nil && !info.IsDir() {
			saveEntries[i] = SaveEntry{
				Used:    true,
				Created: info.ModTime(),
			}
		} else {
			saveEntries[i] = SaveEntry{}
		}
	}
}

func drawError(win *pixelgl.Window, atlas *text.Atlas, errorText string, location pixel.Vec) {
	textScale := float64(1.5)

	textDraw := text.New(location, atlas)
	textDraw.Color = colornames.Red
	textDraw.Dot.X -= textDraw.BoundsOf(errorText).W() / 2
	textDraw.Dot.Y -= textDraw.LineHeight / textScale
	textDraw.WriteString(errorText)
	textDraw.Draw(win, pixel.IM.Scaled(location, textScale))
}

func drawText(win *pixelgl.Window, atlas *text.Atlas, content string, location pixel.Vec) {
	textScale := float64(1.5)

	textDraw := text.New(location, atlas)
	textDraw.Color = colornames.White
	textDraw.WriteString(content)
	textDraw.Draw(win, pixel.IM.Scaled(location, textScale))
}

func drawMenuButton(win *pixelgl.Window, atlas *text.Atlas, buttonText string, bounds pixel.Rect) bool {
	textScale := float64(3)

	imd := imdraw.New(nil)

	imd.Color = colornames.Goldenrod
	imd.EndShape = imdraw.RoundEndShape
	imd.Push(bounds.Min, bounds.Max)
	imd.Rectangle(4)
	imd.Draw(win)

	textDraw := text.New(bounds.Center(), atlas)
	textDraw.Color = colornames.Goldenrod
	textDraw.Dot.X -= textDraw.BoundsOf(buttonText).W() / 2
	textDraw.Dot.Y -= textDraw.LineHeight / textScale
	textDraw.WriteString(buttonText)
	textDraw.Draw(win, pixel.IM.Scaled(bounds.Center(), textScale))

	return win.JustReleased(pixelgl.MouseButtonLeft) &&
		bounds.Contains(win.MousePosition())
}

func runLevelEditTools(win *pixelgl.Window, dt float64) {
	if win.JustPressed(pixelgl.MouseButtonLeft) {
		dragRectStartPoint = win.MousePosition().Scaled(1 / world.Scale)
	}
	if win.JustReleased(pixelgl.MouseButtonLeft) {
		dragRectEndPoint := win.MousePosition().Scaled(1 / world.Scale)
		rect := pixel.R(dragRectStartPoint.X, dragRectStartPoint.Y, dragRectEndPoint.X, dragRectEndPoint.Y)
		fmt.Printf("[[Walls]]\nPos = { X = %4.2f, Y = %4.2f }\nSize = { X = %4.2f, Y = %4.2f }\n",
			rect.Min.X, rect.Min.Y, rect.Size().X, rect.Size().Y)
	}
}

func drawHood(win *pixelgl.Window, dt float64) {
	carHoodSprite.Draw(win, pixel.IM.Moved(carHoodSprite.Frame().Center()).Scaled(pixel.ZV, hoodScale))

	imd := imdraw.New(nil)

	for idx, port := range elcar.Definitions.Ports {
		var sprite *pixel.Sprite
		switch port.PortKind {
		case elcar.PortKindChip:
			sprite = spriteChipPort
		case elcar.PortKindSensor:
			sprite = spriteSensorPort
		}

		if sprite != nil {
			sprite.Draw(win, pixel.IM.Moved(port.HoodPosition).Scaled(pixel.ZV, hoodScale))
		}

		component := car.GetComponent(idx)

		if component.TypeName == "" {
			continue
		}

		componentDef, ok := elcar.Definitions.Components[component.TypeName]
		if !ok {
			continue
		}

		for _, pin := range componentDef.InputPins {
			imd.Clear()
			imd.Color = colornames.Darkolivegreen
			imd.EndShape = imdraw.RoundEndShape
			imd.Push(port.HoodPosition.Scaled(hoodScale), port.HoodPosition.Add(pin.Position).Scaled(hoodScale))
			imd.Line(5)
			imd.Draw(win)

			spritePinIn.Draw(win, pixel.IM.Moved(port.HoodPosition).Moved(pin.Position).Scaled(pixel.ZV, hoodScale))
		}

		for _, pin := range componentDef.OutputPins {
			imd.Clear()
			imd.Color = colornames.Darkolivegreen
			imd.EndShape = imdraw.RoundEndShape
			imd.Push(port.HoodPosition.Scaled(hoodScale), port.HoodPosition.Add(pin.Position).Scaled(hoodScale))
			imd.Line(5)
			imd.Draw(win)

			spritePinOut.Draw(win, pixel.IM.Moved(port.HoodPosition).Moved(pin.Position).Scaled(pixel.ZV, hoodScale))
		}

		sprite = nil
		sprite, ok = componentSprites[component.TypeName]
		if !ok {
			sprite = componentUnknown
		}

		if sprite != nil {
			sprite.Draw(win, pixel.IM.Moved(port.HoodPosition).Scaled(pixel.ZV, hoodScale))
		}

		drawComponentConnections(win, idx)

		if component.State != nil {
			debug := component.State.GetDebugState()
			if debug != "" {
				basicTxt := text.New(port.HoodPosition.Add(pixel.V(0, -8)).Scaled(hoodScale), fontAtlas)
				fmt.Fprintln(basicTxt, debug)
				basicTxt.Draw(win, pixel.IM)
			}
		}

	}

	if win.JustReleased(pixelgl.MouseButtonRight) {
		connectingFromState = NotConnecting
	}
	mouseJustReleased := win.JustReleased(pixelgl.MouseButtonLeft)

	pos := win.MousePosition()
	// Adjust to hood GUI scale
	pos = pos.Scaled(1 / hoodScale)

	for idx, port := range elcar.Definitions.Ports {

		if idx >= elcar.ComponentAny {

			rect := pixel.R(port.HoodPosition.X-7, port.HoodPosition.Y-9, port.HoodPosition.X+7, port.HoodPosition.Y+9)
			if rect.Contains(pos) {

				var tint color.Color
				if selectingComponent != "" && !elcar.IsComponentAllowedInSlot(idx, selectingComponent) {
					tint = color.RGBA{R: 200, A: 40}
				} else {
					tint = color.Alpha{A: 70}
				}
				componentEmpty.DrawColorMask(win, pixel.IM.Moved(port.HoodPosition).Scaled(pixel.ZV, hoodScale), tint)

				// Change component
				if mouseJustReleased {
					if connectingFromState != NotConnecting {
						connectingFromState = NotConnecting
					} else if selectingComponent != "" {
						if elcar.IsComponentAllowedInSlot(idx, selectingComponent) {
							car.AddComponent(idx, selectingComponent)
							selectingComponent = ""
						}
					} else {
						car.RemoveComponent(idx)
					}
				}
			}

		}

		component := car.GetComponent(idx)
		if component.TypeName == "" {
			continue
		}

		componentDef, ok := elcar.Definitions.Components[component.TypeName]
		if !ok {
			continue
		}

		for i, pin := range componentDef.InputPins {
			pinPos := pin.Position.Add(port.HoodPosition).Scaled(hoodScale)
			if math.Abs(win.MousePosition().To(pinPos).Len()) < 10 {

				imd.Clear()
				imd.Color = colornames.Red
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(pinPos)
				imd.Circle(10, 2)
				imd.Draw(win)

				if mouseJustReleased {
					if connectingFromState == ConnectingFromOutput {
						car.ConnectPorts(connectingFromID, connectingFromPort, idx, i)
						connectingFromState = NotConnecting
					} else {
						connectingFromState = ConnectingFromInput
						connectingFromID = idx
						connectingFromPort = i
					}
				}
			}

			if connectingFromState == ConnectingFromInput &&
				connectingFromID == idx && connectingFromPort == i {

				imd.Clear()
				imd.Color = colornames.Blueviolet
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(pinPos, win.MousePosition())
				imd.Line(5)
				imd.Draw(win)
			}
		}

		for i, pin := range componentDef.OutputPins {
			pinPos := pin.Position.Add(port.HoodPosition).Scaled(hoodScale)
			if math.Abs(win.MousePosition().To(pinPos).Len()) < 10 {

				imd.Clear()
				imd.Color = colornames.Red
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(pinPos)
				imd.Circle(10, 2)
				imd.Draw(win)

				if mouseJustReleased {
					if connectingFromState == ConnectingFromInput {
						car.ConnectPorts(idx, i, connectingFromID, connectingFromPort)
						connectingFromState = NotConnecting
					} else {
						connectingFromState = ConnectingFromOutput
						connectingFromID = idx
						connectingFromPort = i
					}
				}
			}

			if connectingFromState == ConnectingFromOutput &&
				connectingFromID == idx && connectingFromPort == i {

				imd.Clear()
				imd.Color = colornames.Blueviolet
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(pinPos, win.MousePosition())
				imd.Line(5)
				imd.Draw(win)
			}
		}
	}
}

func drawComponentSelector(win *pixelgl.Window, dt float64) {
	componentBGSprite.Draw(win, pixel.IM.Moved(pixel.V(carHoodSprite.Frame().W(), 0)).Moved(componentBGSprite.Frame().Center()).Scaled(pixel.ZV, hoodScale))

	basePos := pixel.V(carHoodSprite.Frame().H()+20, carHoodSprite.Frame().W()+-20)

	top := 0.0

	for _, typeName := range componentList {
		def := elcar.Definitions.Components[typeName]
		moveDown := 20.0

		singlePos := pixel.V(0, top)
		rectCenter := basePos.Add(singlePos)

		sprite := componentSprites[typeName]
		if sprite == nil {
			sprite = componentUnknown
		}
		sprite.Draw(win, pixel.IM.Moved(rectCenter).Scaled(pixel.ZV, hoodScale))

		rect := pixel.Rect{
			Min: rectCenter.Sub(componentEmpty.Frame().Size().Scaled(0.5)).Scaled(hoodScale),
			Max: rectCenter.Add(componentEmpty.Frame().Size().Scaled(0.5)).Scaled(hoodScale),
		}

		if rect.Contains(win.MousePosition()) {

			componentEmpty.DrawColorMask(win, pixel.IM.Moved(rectCenter).Scaled(pixel.ZV, hoodScale), color.Alpha{A: 70})

			// Change selected component
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				selectingComponent = typeName
			}
		}

		drawText(win, fontAtlas, def.Name, basePos.Add(singlePos).Add(pixel.V(14, 4)).Scaled(hoodScale))
		desc := strings.Split(def.Description, "\n")
		for i, line := range desc {
			drawText(win, fontAtlas, line, basePos.Add(singlePos).Add(pixel.V(14, float64(-4+i*-5))).Scaled(hoodScale))
		}
		if len(desc) > 1 {
			moveDown += float64(len(desc)-1) * 5
		}
		top -= moveDown
	}

	if win.JustPressed(pixelgl.MouseButtonRight) {
		selectingComponent = ""
	}

	if selectingComponent != "" {
		sprite := componentSprites[selectingComponent]
		if sprite == nil {
			sprite = componentUnknown
		}
		sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, hoodScale).Moved(win.MousePosition()).Moved(pixel.V(16, -16).Scaled(hoodScale)))
	}
}

func drawComponentConnections(win *pixelgl.Window, id int) {
	comp := car.GetComponent(id)
	if len(comp.ConnectedOutputs) == 0 {
		return
	}
	if id < 0 || id >= len(elcar.Definitions.Ports) {
		return
	}

	pos := elcar.Definitions.Ports[id].HoodPosition

	for outPin, conn := range comp.ConnectedOutputs {

		if conn.ID < 0 || conn.ID >= len(elcar.Definitions.Ports) {
			continue
		}

		targetComponent := car.GetComponent(conn.ID)
		if targetComponent.State == nil {
			continue
		}

		pinOffsetOut := elcar.GetOutPinPosition(comp.TypeName, outPin)

		targetPos := elcar.Definitions.Ports[conn.ID].HoodPosition

		pinOffsetIn := elcar.GetInPinPosition(targetComponent.TypeName, conn.Pin)

		imd := imdraw.New(nil)
		imd.Color = colornames.Red
		imd.EndShape = imdraw.RoundEndShape
		imd.Push(pos.Add(pinOffsetOut).Scaled(hoodScale), targetPos.Add(pinOffsetIn).Scaled(hoodScale))
		imd.Line(5)
		imd.Draw(win)
	}
}
