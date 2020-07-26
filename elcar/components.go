package elcar

import (
	"strconv"

	"founderio.net/eljam/world"
	"github.com/faiface/pixel"
)

type AddComponent struct {
	inputs []float64
	value  float64
}

func (c *AddComponent) Update(car *Car, objects *world.Objects) {
	var newValue float64
	for _, val := range c.inputs {
		newValue += val
	}
	c.value = newValue
}
func (c *AddComponent) GetSpriteName() string {
	return CTypeAdd
}
func (c *AddComponent) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *AddComponent) GetInputCount() int {
	return 3
}
func (c *AddComponent) GetOutputCount() int {
	return 1
}

func (c *AddComponent) SetInputs(values []float64, connected []bool) {
	c.inputs = values
}
func (c *AddComponent) GetOutputs() []float64 {
	return []float64{c.value}
}

type MultiplyComponent struct {
	inputs    []float64
	connected []bool
	value     float64
}

func (c *MultiplyComponent) Update(car *Car, objects *world.Objects) {
	var newValue float64 = 1
	for i, val := range c.inputs {
		if c.connected[i] {
			newValue *= val
		}
	}
	c.value = newValue
}
func (c *MultiplyComponent) GetSpriteName() string {
	return CTypeMultiply
}
func (c *MultiplyComponent) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *MultiplyComponent) GetInputCount() int {
	return 3
}
func (c *MultiplyComponent) GetOutputCount() int {
	return 1
}

func (c *MultiplyComponent) SetInputs(values []float64, connected []bool) {
	c.inputs = values
	c.connected = connected
}
func (c *MultiplyComponent) GetOutputs() []float64 {
	return []float64{c.value}
}

type ConstantValue struct {
}

func (c *ConstantValue) GetSpriteName() string {
	return CTypeConstant
}
func (c *ConstantValue) GetDebugState() string {
	return strconv.FormatFloat(1, 'g', 3, 64)
}

func (c *ConstantValue) Update(car *Car, objects *world.Objects) {
}

func (c *ConstantValue) GetInputCount() int {
	return 0
}
func (c *ConstantValue) GetOutputCount() int {
	return 1
}

func (c *ConstantValue) SetInputs(values []float64, connected []bool) {
}
func (c *ConstantValue) GetOutputs() []float64 {
	return []float64{1}
}

type ComponentRadar struct {
	value float64
}

func (c *ComponentRadar) GetSpriteName() string {
	return CTypeRadar
}
func (c *ComponentRadar) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *ComponentRadar) Update(car *Car, objects *world.Objects) {
	//TODO: this beam needs to start at the actual component in world-space
	// currently, this is the center of the car
	beamLength := float64(50)
	beamStart := car.Position
	maxBeamExtent := beamStart.Add(car.Forward().Scaled(beamLength))

	checkLine := pixel.L(beamStart, maxBeamExtent)

	closestPoint := maxBeamExtent
	closestDistance := beamLength

	for _, o := range objects.Collidables {
		intersections := o.Bounds().IntersectionPoints(checkLine)
		for _, intersectionPoint := range intersections {
			dist := absDistance(intersectionPoint, beamStart)
			if dist < closestDistance {
				closestDistance = dist
				closestPoint = intersectionPoint
			}
		}
	}
	car.DebugPoints = append(car.DebugPoints, closestPoint)
	c.value = 1 - closestDistance/beamLength
}

func (c *ComponentRadar) GetInputCount() int {
	return 0
}
func (c *ComponentRadar) GetOutputCount() int {
	return 1
}

func (c *ComponentRadar) SetInputs(values []float64, connected []bool) {
}
func (c *ComponentRadar) GetOutputs() []float64 {
	return []float64{c.value}
}
