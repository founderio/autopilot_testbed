package elcar

import (
	"github.com/faiface/pixel"
)

type World struct {
	Scale            float64
	BackgroundSprite string
	Size             pixel.Vec
	Walls            []Wall
	Props            []Prop
}

func (w World) Bounds() pixel.Rect {
	return pixel.Rect{
		Min: pixel.ZV,
		Max: w.Size,
	}
}

type Wall struct {
	Pos  pixel.Vec
	Size pixel.Vec

	Solidity            float64
	ReflectivenessLight float64
	ReflectivenessRadar float64
}

func (c Wall) Bounds() pixel.Rect {
	return pixel.Rect{
		Min: c.Pos,
		Max: c.Pos.Add(c.Size),
	}
}

type Prop struct {
	Pos  pixel.Vec
	Name string
}

func (c Prop) Bounds(def SpriteDefinition) pixel.Rect {
	return pixel.Rect{
		Min: c.Pos,
		Max: c.Pos.Add(def.Size),
	}
}
