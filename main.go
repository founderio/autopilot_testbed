package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"time"

	// Enable loading of PNG files
	"image/color"
	_ "image/png"

	"founderio.net/eljam/elcar"
	"founderio.net/eljam/world"
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

func loadPicture(filename string) (pixel.Picture, error) {
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

	carHoodSprite *pixel.Sprite

	componentEmpty   *pixel.Sprite
	componentUnknown *pixel.Sprite

	componentSprites map[string]*pixel.Sprite

	spritePinIn      *pixel.Sprite
	spritePinOut     *pixel.Sprite
	spriteChipPort   *pixel.Sprite
	spriteSensorPort *pixel.Sprite
)

var (
	componentList []string
)

var (
	car *elcar.Car
)

var (
	hoodScale  float64 = 3
	worldScale float64 = 4

	connectingFromState int
	connectingFromID    int
	connectingFromPort  int

	selectingComponent string
)

const (
	NotConnecting int = iota
	ConnectingFromInput
	ConnectingFromOutput
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Electronics Jam",
		Bounds: pixel.R(0, 0, 1800, 1000),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	_, err = toml.DecodeFile(filepath.Join("resources", "definitions.toml"), &elcar.Definitions)
	if err != nil {
		panic(err)
	}

	for typeName, def := range elcar.Definitions.Components {
		if def.Usable {
			componentList = append(componentList, typeName)
		}
	}

	var world *world.Objects
	_, err = toml.DecodeFile(filepath.Join("resources", "world.toml"), &world)
	if err != nil {
		panic(err)
	}
	// For now, calculate from screen size - later on we need to draw a border and add camera panning
	world.WorldBorder = pixel.R(0, 0, 1920/worldScale, 1080/worldScale)

	fontAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)

	stonePic, err := loadPicture("stone.png")
	if err != nil {
		panic(err)
	}
	stoneSprite := pixel.NewSprite(stonePic, stonePic.Bounds())

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

	componentSprites = make(map[string]*pixel.Sprite)

	componentSpriteSheet, err := loadPicture("components.png")
	if err != nil {
		panic(err)
	}
	componentEmpty = pixel.NewSprite(componentSpriteSheet,
		pixel.R(0, 96, 32, 128))
	componentSprites[elcar.CTypeMultiply] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(32, 96, 64, 128))
	componentSprites[elcar.CTypeAdd] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(64, 96, 96, 128))
	componentUnknown = pixel.NewSprite(componentSpriteSheet,
		pixel.R(96, 96, 128, 128))

	componentSprites[elcar.CTypeBuiltinSteering] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(0, 64, 32, 96))
	componentSprites[elcar.CTypeSubtract] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(32, 64, 64, 96))
	componentSprites[elcar.CTypeBuiltinAcceleration] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(64, 64, 96, 96))
	componentSprites[elcar.CTypeBuiltinBraking] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(96, 64, 128, 96))

	componentSprites[elcar.CTypeRadar] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(0, 32, 32, 64))
	componentSprites[elcar.CTypeCompareEquals] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(32, 32, 64, 64))
	componentSprites[elcar.CTypeSplitSignal] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(64, 32, 96, 64))
	componentSprites[elcar.CTypeRadarShortrange] = pixel.NewSprite(componentSpriteSheet,
		pixel.R(96, 32, 128, 64))

	//componentSprites[elcar.CType] = pixel.NewSprite(componentSpriteSheet,
	//	pixel.R(0, 0, 32, 32))
	//componentSprites[elcar.CType] = pixel.NewSprite(componentSpriteSheet,
	//	pixel.R(32, 0, 64, 32))
	//componentSprites[elcar.CType] = pixel.NewSprite(componentSpriteSheet,
	//	pixel.R(64, 0, 96, 32))
	//componentSprites[elcar.CType] = pixel.NewSprite(componentSpriteSheet,
	//	pixel.R(96, 0, 128, 32))

	pcbSpriteSheet, err := loadPicture("pcb.png")
	if err != nil {
		panic(err)
	}
	spritePinIn = pixel.NewSprite(pcbSpriteSheet, pixel.R(0, 0, 6, 6))
	spritePinOut = pixel.NewSprite(pcbSpriteSheet, pixel.R(0, 6, 6, 6+6))
	spriteSensorPort = pixel.NewSprite(pcbSpriteSheet, pixel.R(20, 0, 20+14, 18))
	spriteChipPort = pixel.NewSprite(pcbSpriteSheet, pixel.R(34, 0, 34+14, 18))

	car = &elcar.Car{
		Position: pixel.V(150, 120),
		Rotation: 0,
		Speed:    0,
	}

	for idx, port := range elcar.Definitions.Ports {
		if port.Prefill != "" {
			car.AddComponent(idx, port.Prefill)
		}
	}

	imd := imdraw.New(nil)

	hoodOpen := false

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.JustPressed(pixelgl.KeyTab) {
			hoodOpen = !hoodOpen
		}

		car.Update(dt, world)

		win.Clear(colornames.Gainsboro)

		for _, o := range world.Collidables {
			stoneSprite.Draw(win, pixel.IM.Moved(o.Bounds().Center()).Scaled(pixel.ZV, worldScale))

			imd.Clear()
			imd.Color = colornames.Gray
			imd.Push(o.Pos.Scaled(worldScale), o.Pos.Add(o.Size).Scaled(worldScale))
			imd.Rectangle(2)
			imd.Draw(win)
		}

		mat := pixel.IM.Rotated(pixel.ZV, -car.Rotation)
		mat = mat.Moved(car.Position)
		mat = mat.Scaled(pixel.ZV, worldScale)
		carSprite.Draw(win, mat)

		for _, debug := range car.DebugPoints {
			imd.Clear()

			imd.Color = colornames.Magenta
			imd.EndShape = imdraw.SharpEndShape
			imd.Push(car.Position.Scaled(worldScale), debug.Scaled(worldScale))
			imd.Line(1)
			imd.Draw(win)
		}

		for _, debug := range car.DebugLines {
			imd.Clear()

			imd.Color = colornames.Magenta
			imd.EndShape = imdraw.SharpEndShape
			imd.Push(debug.A.Scaled(worldScale), debug.B.Scaled(worldScale))
			imd.Line(1)
			imd.Draw(win)
		}

		if hoodOpen {
			drawHood(win, dt)
			drawComponentSelector(win, dt)
		} else {
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				dragRectStartPoint = win.MousePosition().Scaled(1 / worldScale)
			}
			if win.JustReleased(pixelgl.MouseButtonLeft) {
				dragRectEndPoint := win.MousePosition().Scaled(1 / worldScale)
				rect := pixel.R(dragRectStartPoint.X, dragRectStartPoint.Y, dragRectEndPoint.X, dragRectEndPoint.Y)
				fmt.Printf("[[Collidables]]\nPos = { X = %4.2f, Y = %4.2f }\nSize = { X = %4.2f, Y = %4.2f }\n",
					rect.Min.X, rect.Min.Y, rect.Size().X, rect.Size().Y)
			}
		}

		win.Update()
	}
}

var dragRectStartPoint pixel.Vec

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
				if selectingComponent != "" && !isComponentAllowedInSlot(idx, selectingComponent) {
					tint = color.RGBA{R: 200, A: 40}
				} else {
					tint = color.Alpha{A: 40}
				}
				componentEmpty.DrawColorMask(win, pixel.IM.Moved(port.HoodPosition).Scaled(pixel.ZV, hoodScale), tint)

				// Change component
				if mouseJustReleased {
					if connectingFromState != NotConnecting {
						connectingFromState = NotConnecting
					} else if selectingComponent != "" {
						if isComponentAllowedInSlot(idx, selectingComponent) {
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
	basePos := pixel.V(270, 160)

	for i, typeName := range componentList {
		singlePos := pixel.V(float64((i%2)*32), float64((i/-2)*32))

		sprite := componentSprites[typeName]
		if sprite == nil {
			sprite = componentUnknown
		}
		sprite.Draw(win, pixel.IM.Moved(basePos).Moved(singlePos).Moved(pixel.V(16, 16)).Scaled(pixel.ZV, hoodScale))

		rect := pixel.R(singlePos.X+basePos.X+10, singlePos.Y+basePos.Y+7, singlePos.X+basePos.X+24, singlePos.Y+basePos.Y+26)

		if rect.Contains(win.MousePosition().Scaled(1 / hoodScale)) {

			componentEmpty.DrawColorMask(win, pixel.IM.Moved(basePos).Moved(singlePos).Moved(pixel.V(16, 16)).Scaled(pixel.ZV, hoodScale), color.Alpha{A: 40})

			// Change selected component
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				selectingComponent = typeName
			}
		}
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

func getInPinPosition(typeName string, port int) pixel.Vec {
	def, ok := elcar.Definitions.Components[typeName]
	if !ok {
		return pixel.ZV
	}

	if port < 0 || port >= len(def.InputPins) {
		return pixel.ZV
	}

	return def.InputPins[port].Position
}

func getOutPinPosition(typeName string, port int) pixel.Vec {
	def, ok := elcar.Definitions.Components[typeName]
	if !ok {
		return pixel.ZV
	}

	if port < 0 || port >= len(def.OutputPins) {
		return pixel.ZV
	}

	return def.OutputPins[port].Position
}

func isComponentAllowedInSlot(id int, typeName string) bool {
	if id < 0 || id >= len(elcar.Definitions.Ports) {
		return false
	}
	portDef := elcar.Definitions.Ports[id]
	componentDef, ok := elcar.Definitions.Components[typeName]
	if !ok {
		return false
	}
	return portDef.PortKind == componentDef.PortKind
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

		pinOffsetOut := getOutPinPosition(comp.TypeName, outPin)

		targetPos := elcar.Definitions.Ports[conn.ID].HoodPosition

		pinOffsetIn := getInPinPosition(targetComponent.TypeName, conn.Pin)

		imd := imdraw.New(nil)
		imd.Color = colornames.Red
		imd.EndShape = imdraw.RoundEndShape
		imd.Push(pos.Add(pinOffsetOut).Scaled(hoodScale), targetPos.Add(pinOffsetIn).Scaled(hoodScale))
		imd.Line(5)
		imd.Draw(win)
	}
}
