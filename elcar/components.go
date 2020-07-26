package elcar

import (
	"fmt"
	"math"
	"strconv"

	"founderio.net/eljam/world"
	"github.com/faiface/pixel"
)

type BuiltinSteering struct {
	steerLeft, steerRight float64
}

func (c *BuiltinSteering) Update(car *Car, objects *world.Objects, port PortDefinition) {
	car.Steering = c.steerRight - c.steerLeft
}
func (c *BuiltinSteering) GetDebugState() string {
	return fmt.Sprintf("l: %g r: %g", c.steerLeft, c.steerRight)
}

func (c *BuiltinSteering) SetInputs(values []float64, connected []bool) {
	c.steerLeft, c.steerRight = values[0], values[1]
}
func (c *BuiltinSteering) GetOutputs() []float64 {
	return []float64{}
}

type BuiltinAcceleration struct {
	acceleration float64
}

func (c *BuiltinAcceleration) Update(car *Car, objects *world.Objects, port PortDefinition) {
	car.Acceleration = c.acceleration
}
func (c *BuiltinAcceleration) GetDebugState() string {
	return strconv.FormatFloat(c.acceleration, 'g', 3, 64)
}

func (c *BuiltinAcceleration) SetInputs(values []float64, connected []bool) {
	c.acceleration = values[0]
}
func (c *BuiltinAcceleration) GetOutputs() []float64 {
	return []float64{}
}

type BuiltinBraking struct {
	braking float64
}

func (c *BuiltinBraking) Update(car *Car, objects *world.Objects, port PortDefinition) {
	car.Braking = c.braking
}
func (c *BuiltinBraking) GetDebugState() string {
	return strconv.FormatFloat(c.braking, 'g', 3, 64)
}

func (c *BuiltinBraking) SetInputs(values []float64, connected []bool) {
	c.braking = values[0]
}
func (c *BuiltinBraking) GetOutputs() []float64 {
	return []float64{}
}

type CompareEquals struct {
	a, b  float64
	value float64
}

func (c *CompareEquals) Update(car *Car, objects *world.Objects, port PortDefinition) {
	absDiff := math.Abs(c.a - c.b)
	if absDiff < 0.5 {
		c.value = (0.5 - absDiff) * 2
	} else {
		c.value = 0
	}
}
func (c *CompareEquals) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *CompareEquals) SetInputs(values []float64, connected []bool) {
	c.a, c.b = values[0], values[1]
}
func (c *CompareEquals) GetOutputs() []float64 {
	return []float64{c.value}
}

type Subtract struct {
	a, b  float64
	value float64
}

func (c *Subtract) Update(car *Car, objects *world.Objects, port PortDefinition) {
	c.value = c.a - c.b
}
func (c *Subtract) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *Subtract) SetInputs(values []float64, connected []bool) {
	c.a, c.b = values[0], values[1]
}
func (c *Subtract) GetOutputs() []float64 {
	return []float64{c.value}
}

type Add struct {
	inputs []float64
	value  float64
}

func (c *Add) Update(car *Car, objects *world.Objects, port PortDefinition) {
	var newValue float64
	for _, val := range c.inputs {
		newValue += val
	}
	c.value = newValue
}
func (c *Add) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *Add) SetInputs(values []float64, connected []bool) {
	c.inputs = values
}
func (c *Add) GetOutputs() []float64 {
	return []float64{c.value}
}

type Multiply struct {
	inputs    []float64
	connected []bool
	value     float64
}

func (c *Multiply) Update(car *Car, objects *world.Objects, port PortDefinition) {
	var newValue float64 = 1
	for i, val := range c.inputs {
		if c.connected[i] {
			newValue *= val
		}
	}
	c.value = newValue
}
func (c *Multiply) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *Multiply) SetInputs(values []float64, connected []bool) {
	c.inputs = values
	c.connected = connected
}
func (c *Multiply) GetOutputs() []float64 {
	return []float64{c.value}
}

type ConstantValue struct {
}

func (c *ConstantValue) GetDebugState() string {
	return strconv.FormatFloat(1, 'g', 3, 64)
}

func (c *ConstantValue) Update(car *Car, objects *world.Objects, port PortDefinition) {
}

func (c *ConstantValue) SetInputs(values []float64, connected []bool) {
}
func (c *ConstantValue) GetOutputs() []float64 {
	return []float64{1, 1, 1}
}

type SplitSignal struct {
	input float64
}

func (c *SplitSignal) GetDebugState() string {
	return strconv.FormatFloat(c.input, 'g', 3, 64)
}

func (c *SplitSignal) Update(car *Car, objects *world.Objects, port PortDefinition) {
}

func (c *SplitSignal) SetInputs(values []float64, connected []bool) {
	c.input = values[0]
}
func (c *SplitSignal) GetOutputs() []float64 {
	return []float64{c.input, c.input, c.input}
}

func castLine(objects *world.Objects, line pixel.Line, maxDistance float64) (float64, pixel.Vec) {
	closestPoint := line.B
	closestDistance := maxDistance

	intersections := objects.WorldBorder.IntersectionPoints(line)
	for _, intersectionPoint := range intersections {
		dist := absDistance(intersectionPoint, line.A)
		if dist < closestDistance {
			closestDistance = dist
			closestPoint = intersectionPoint
		}
	}

	for _, o := range objects.Collidables {
		intersections := o.Bounds().IntersectionPoints(line)
		for _, intersectionPoint := range intersections {
			dist := absDistance(intersectionPoint, line.A)
			if dist < closestDistance {
				closestDistance = dist
				closestPoint = intersectionPoint
			}
		}
	}

	return closestDistance, closestPoint
}

func castCircle(objects *world.Objects, circle pixel.Circle, direction pixel.Vec, maxDistance float64) (float64, pixel.Vec) {
	closestDistance := maxDistance
	closestPoint := circle.Center.Add(direction.Scaled(maxDistance))

	for _, line := range objects.WorldBorder.Edges() {
		intersects := circle.IntersectLine(line)
		if intersects != pixel.ZV {
			intersectionPoint := line.Closest(circle.Center)

			dist := absDistance(intersectionPoint, circle.Center)
			if dist < closestDistance {
				closestDistance = dist
				closestPoint = intersectionPoint
			}
		}
	}

	for _, o := range objects.Collidables {
		for _, line := range o.Bounds().Edges() {
			intersects := circle.IntersectLine(line)
			if intersects != pixel.ZV {
				intersectionPoint := line.Closest(circle.Center)

				dist := absDistance(intersectionPoint, circle.Center)
				if dist < closestDistance {
					closestDistance = dist
					closestPoint = intersectionPoint
				}
			}
		}
	}

	return closestDistance, closestPoint
}

type Radar struct {
	value float64
}

func (c *Radar) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *Radar) Update(car *Car, objects *world.Objects, port PortDefinition) {
	beamLength := float64(50)
	shortBeamLength := float64(10)
	beamStart := port.WorldPosition.Rotated(-car.Rotation).Add(car.Position)
	beamDirection := port.Direction.Rotated(-car.Rotation).Unit()
	maxBeamExtent := beamStart.Add(beamDirection.Scaled(beamLength))

	checkLine := pixel.L(beamStart, maxBeamExtent)
	beamCircle := pixel.Circle{
		Center: beamStart,
		Radius: beamLength,
	}

	// Long-distance check, linecast
	closestDistance, closestPoint := castLine(objects, checkLine, beamLength)

	// Check in circle directly around the sensor
	circleDistance, circlePoint := castCircle(objects, beamCircle, beamDirection, shortBeamLength)
	if circleDistance < closestDistance &&
		circleDistance < (shortBeamLength-0.001) {
		closestDistance = circleDistance
		closestPoint = circlePoint
	}

	car.DebugLines = append(car.DebugLines, pixel.L(beamStart, closestPoint))
	c.value = 1 - closestDistance/beamLength
}

func (c *Radar) SetInputs(values []float64, connected []bool) {
}
func (c *Radar) GetOutputs() []float64 {
	return []float64{c.value}
}

type RadarShortrange struct {
	value float64
}

func (c *RadarShortrange) GetDebugState() string {
	return strconv.FormatFloat(c.value, 'g', 3, 64)
}

func (c *RadarShortrange) Update(car *Car, objects *world.Objects, port PortDefinition) {
	beamLength := float64(10)
	beamStart := port.WorldPosition.Rotated(-car.Rotation).Add(car.Position)
	beamDirection := port.Direction.Rotated(-car.Rotation).Unit()
	beamCircle := pixel.Circle{
		Center: beamStart,
		Radius: beamLength,
	}

	closestDistance, closestPoint := castCircle(objects, beamCircle, beamDirection, beamLength)

	car.DebugLines = append(car.DebugLines, pixel.L(beamStart, closestPoint))
	c.value = 1 - closestDistance/beamLength
}

func (c *RadarShortrange) SetInputs(values []float64, connected []bool) {
}
func (c *RadarShortrange) GetOutputs() []float64 {
	return []float64{c.value}
}
