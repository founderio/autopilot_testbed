package main

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"time"

	// Enable loading of PNG files
	"image/color"
	_ "image/png"

	"founderio.net/eljam/elcar"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
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
	carHoodSprite  *pixel.Sprite
	componentEmpty *pixel.Sprite
	//componentRadar      *pixel.Sprite
	componentAdd     *pixel.Sprite
	componentUnknown *pixel.Sprite

	componentSteerLeft  *pixel.Sprite
	componentSteerRight *pixel.Sprite
	componentAccelerate *pixel.Sprite
	componentBrake      *pixel.Sprite

	componentRadar *pixel.Sprite
	// componentSteerRight *pixel.Sprite
	// componentAccelerate *pixel.Sprite
	// componentBrake      *pixel.Sprite

	componentSprites map[string]*pixel.Sprite

	componentLocations []pixel.Vec
	allowedComponents  [][]string
)

var (
	car *elcar.Car
)

var (
	hoodScale           float64 = 3
	connectingFromState int
	connectingFromID    int
	connectingFromPort  int
)

const (
	NotConnecting int = iota
	ConnectingFromInput
	ConnectingFromOutput
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Electronics Jam",
		Bounds: pixel.R(0, 0, 1024, 768),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

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

	componentSpriteSheet, err := loadPicture("components.png")
	if err != nil {
		panic(err)
	}
	componentEmpty = pixel.NewSprite(componentSpriteSheet, pixel.R(0, 96, 32, 128))
	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(32, 96, 64, 128))
	componentAdd = pixel.NewSprite(componentSpriteSheet, pixel.R(64, 96, 96, 128))
	componentUnknown = pixel.NewSprite(componentSpriteSheet, pixel.R(96, 96, 128, 128))

	componentSteerLeft = pixel.NewSprite(componentSpriteSheet, pixel.R(0, 64, 32, 96))
	componentSteerRight = pixel.NewSprite(componentSpriteSheet, pixel.R(32, 64, 64, 96))
	componentAccelerate = pixel.NewSprite(componentSpriteSheet, pixel.R(64, 64, 96, 96))
	componentBrake = pixel.NewSprite(componentSpriteSheet, pixel.R(96, 64, 128, 96))

	componentRadar = pixel.NewSprite(componentSpriteSheet, pixel.R(0, 32, 32, 64))
	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(32, 32, 64, 64))
	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(64, 32, 96, 64))
	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(96, 32, 128, 64))

	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(0, 0, 32, 32))
	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(32, 0, 64, 32))
	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(64, 0, 96, 32))
	//component = pixel.NewSprite(componentSpriteSheet, pixel.R(96, 0, 128, 32))

	componentSprites = make(map[string]*pixel.Sprite)
	componentSprites[elcar.CTypeAdd] = componentAdd
	componentSprites[elcar.CTypeRadar] = componentRadar

	car = &elcar.Car{
		Position: pixel.V(15, 10),
		Rotation: 0,
		Speed:    15,
		Components: []elcar.UsedComponent{
			{
				ID: elcar.ComponentAny,
				// ConnectedOutputs: []elcar.ComponentDestination{
				// 	{
				// 		ID:   elcar.ComponentSteerLeft,
				// 		Port: 0,
				// 	},
				// },
				State: &elcar.ConstantValue{},
			},
		},
	}

	componentLocations = []pixel.Vec{
		pixel.V(23, 256-44),  // ComponentSteerLeft
		pixel.V(88, 256-44),  // ComponentSteerRight
		pixel.V(138, 256-44), // ComponentAccelerate
		pixel.V(199, 256-44), // ComponentBrake
	}
	allowedComponents = [][]string{
		{}, // ComponentSteerLeft
		{}, // ComponentSteerRight
		{}, // ComponentAccelerate
		{}, // ComponentBrake
	}
	// ComponentAny, front-facing sensor mounts

	componentLocations = append(componentLocations, pixel.V(55, 256-25))
	componentLocations = append(componentLocations, pixel.V(169, 256-25))
	allowedComponents = append(allowedComponents, []string{elcar.CTypeRadar})
	allowedComponents = append(allowedComponents, []string{elcar.CTypeRadar})

	base := pixel.V(9, 256-85) // ComponentAny, internal chip mounts
	for x := 0; x < 7; x++ {
		for y := 0; y < 2; y++ {
			componentLocations = append(componentLocations, base.Add(pixel.V(float64(x*32), float64(y*-32))))
			allowedComponents = append(allowedComponents, []string{elcar.CTypeAdd, elcar.CTypeConstant})
		}
	}

	hoodOpen := false

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.JustPressed(pixelgl.KeyTab) {
			hoodOpen = !hoodOpen
		}

		car.Update(dt)

		win.Clear(colornames.Gainsboro)
		stoneSprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 10).Moved(win.Bounds().Center()))

		mat := pixel.IM.Scaled(pixel.ZV, 10)
		mat = mat.Rotated(pixel.ZV, -car.Rotation)
		mat = mat.Moved(win.Bounds().Center())
		mat = mat.Moved(car.Position)
		carSprite.Draw(win, mat)

		if hoodOpen {
			drawHood(win, dt)
			drawComponentSelector(win, dt)
		}

		win.Update()
	}
}

func drawHood(win *pixelgl.Window, dt float64) {
	carHoodSprite.Draw(win, pixel.IM.Moved(carHoodSprite.Frame().Center()).Scaled(pixel.ZV, hoodScale))

	for idx, location := range componentLocations {
		var sprite *pixel.Sprite
		switch idx {
		case elcar.ComponentSteerLeft:
			sprite = componentSteerLeft
		case elcar.ComponentSteerRight:
			sprite = componentSteerRight
		case elcar.ComponentAccelerate:
			sprite = componentAccelerate
		case elcar.ComponentBrake:
			sprite = componentBrake

		default:
			compo := car.GetComponent(idx)
			if compo.State != nil {
				var ok bool
				sprite, ok = componentSprites[compo.State.GetSpriteName()]
				if !ok {
					sprite = componentUnknown
				}
			}
		}

		if sprite != nil {
			sprite.Draw(win, pixel.IM.Moved(location).Moved(pixel.V(16, 16)).Scaled(pixel.ZV, hoodScale))
		}
		drawComponentConnections(win, idx)
	}

	if win.JustReleased(pixelgl.MouseButtonRight) {
		connectingFromState = NotConnecting
	}

	mouseJustReleased := win.JustReleased(pixelgl.MouseButtonLeft)

	pos := win.MousePosition()
	// Adjust to hood GUI scale
	pos = pos.Scaled(1 / hoodScale)

	for idx, location := range componentLocations {
		if idx >= elcar.ComponentAny {

			rect := pixel.R(location.X+10, location.Y+7, location.X+24, location.Y+26)
			if rect.Contains(pos) {

				componentEmpty.DrawColorMask(win, pixel.IM.Moved(location).Moved(pixel.V(16, 16)).Scaled(pixel.ZV, hoodScale), color.Alpha{A: 40})

				// Change component
				if mouseJustReleased {
					if connectingFromState != NotConnecting {
						connectingFromState = NotConnecting
					} else if car.GetComponent(idx).State == nil && selectingComponent != "" {
						car.AddComponent(idx, componentMakerFuncs[selectingComponent]())
						selectingComponent = ""
					} else {
						car.RemoveComponent(idx)
					}
				}
			}

		}

		for i := 0; i < 3; i++ {
			portPos := getInPortPosition(i).Add(location).Scaled(hoodScale)
			if math.Abs(win.MousePosition().To(portPos).Len()) < 10 {

				imd := imdraw.New(nil)
				imd.Color = colornames.Red
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(portPos)
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

				imd := imdraw.New(nil)
				imd.Color = colornames.Blueviolet
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(portPos, win.MousePosition())
				imd.Line(5)
				imd.Draw(win)
			}
		}

		for i := 0; i < 3; i++ {
			portPos := getOutPortPosition(i).Add(location).Scaled(hoodScale)
			if math.Abs(win.MousePosition().To(portPos).Len()) < 10 {

				imd := imdraw.New(nil)
				imd.Color = colornames.Red
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(portPos)
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

				imd := imdraw.New(nil)
				imd.Color = colornames.Blueviolet
				imd.EndShape = imdraw.RoundEndShape
				imd.Push(portPos, win.MousePosition())
				imd.Line(5)
				imd.Draw(win)
			}
		}
	}
}

var componentList = []string{
	elcar.CTypeConstant,
	elcar.CTypeAdd,
	elcar.CTypeRadar,
}
var selectingComponent string

var componentMakerFuncs = map[string]func() elcar.Component{
	elcar.CTypeConstant: func() elcar.Component {
		return &elcar.ConstantValue{}
	},
	elcar.CTypeAdd: func() elcar.Component {
		return &elcar.AddComponent{}
	},
	elcar.CTypeRadar: func() elcar.Component {
		return &elcar.ComponentRadar{}
	},
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

var (
	PinOffsetIn1  = pixel.V(4, 26)
	PinOffsetIn2  = pixel.V(4, 17)
	PinOffsetIn3  = pixel.V(4, 8)
	PinOffsetOut1 = pixel.V(30, 26)
	PinOffsetOut2 = pixel.V(30, 17)
	PinOffsetOut3 = pixel.V(30, 8)
)

func getInPortPosition(port int) pixel.Vec {
	switch port {
	case 0:
		return PinOffsetIn1
	case 1:
		return PinOffsetIn2
	case 2:
		return PinOffsetIn3
	default:
		return pixel.ZV
	}
}

func getOutPortPosition(port int) pixel.Vec {
	switch port {
	case 0:
		return PinOffsetOut1
	case 1:
		return PinOffsetOut2
	case 2:
		return PinOffsetOut3
	default:
		return pixel.ZV
	}
}

func drawComponentConnections(win *pixelgl.Window, id int) {
	comp := car.GetComponent(id)
	if len(comp.ConnectedOutputs) == 0 {
		return
	}
	if id < 0 || id >= len(componentLocations) {
		return
	}

	pos := componentLocations[id]

	for outPort, conn := range comp.ConnectedOutputs {

		if conn.ID < 0 || conn.ID >= len(componentLocations) {
			continue
		}

		pinOffsetOut := getOutPortPosition(outPort)

		targetPos := componentLocations[conn.ID]

		pinOffsetIn := getInPortPosition(conn.Port)

		imd := imdraw.New(nil)
		imd.Color = colornames.Red
		imd.EndShape = imdraw.RoundEndShape
		imd.Push(pos.Add(pinOffsetOut).Scaled(hoodScale), targetPos.Add(pinOffsetIn).Scaled(hoodScale))
		imd.Line(5)
		imd.Draw(win)
	}
}
