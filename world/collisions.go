package world

import "github.com/faiface/pixel"

type Objects struct {
	WorldBorder pixel.Rect
	Collidables []Collidable
}

type Collidable struct {
	Pos  pixel.Vec
	Size pixel.Vec

	Solidity            float64
	ReflectivenessLight float64
	ReflectivenessRadar float64
}

func (c Collidable) Bounds() pixel.Rect {
	return pixel.Rect{
		Min: c.Pos,
		Max: c.Pos.Add(c.Size),
	}
}
