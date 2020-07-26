package elcar

import (
	"math"

	"founderio.net/eljam/world"
	"github.com/faiface/pixel"
)

const (
	// Default components present in the car
	ComponentSteerLeft int = iota
	ComponentSteerRight
	ComponentAccelerate
	ComponentBrake

	// First "custom" ID
	ComponentAny
)

const (
	CTypeAdd      = "add"
	CTypeMultiply = "multiply"
	CTypeRadar    = "radar"
	CTypeConstant = "constant"
)

func absDistance(a, b pixel.Vec) float64 {
	return math.Abs(a.To(b).Len())
}

type Car struct {
	Position pixel.Vec
	Rotation float64
	Speed    float64

	Components  []UsedComponent
	DebugPoints []pixel.Vec
}

func (c *Car) GetComponent(id int) UsedComponent {
	for _, component := range c.Components {
		if component.ID == id {
			return component
		}
	}
	return UsedComponent{
		ID: id,
	}
}

func (c *Car) AddComponent(id int, state Component) {
	for i, component := range c.Components {
		if component.ID == id {
			component.State = state

			// Ensure the connections are initialized and NOT connected to anything (-1)
			//TODO: does not work, yet
			component.ConnectedOutputs = make([]ComponentDestination, state.GetOutputCount())
			for i, o := range component.ConnectedOutputs {
				o.ID = -1
				component.ConnectedOutputs[i] = o
			}

			c.Components[i] = component
			return
		}
	}
	c.Components = append(c.Components, UsedComponent{
		ID:    id,
		State: state,
	})
}

func (c *Car) RemoveComponent(id int) {
	for i, component := range c.Components {
		if component.ID == id {
			c.Components = append(c.Components[:i], c.Components[i+1:]...)
			return
		}
	}
}

func (c *Car) ConnectPorts(id, port, targetID, targetPort int) {
	for i, component := range c.Components {
		if component.ID == id {

			if len(component.ConnectedOutputs) == 0 {
				component.ConnectedOutputs = make([]ComponentDestination, 3)
				//TODO: do we allow more/less than 3 outputs? Code handles this differently at the moment...
			}
			if port < 0 || port >= len(component.ConnectedOutputs) {
				return
			}

			component.ConnectedOutputs[port] = ComponentDestination{
				ID:   targetID,
				Port: targetPort,
			}
			c.Components[i] = component

			return
		}
	}
}

func (c *Car) Forward() pixel.Vec {
	return pixel.Unit(-c.Rotation)
}

func (c *Car) Update(dt float64, objects *world.Objects) {
	// Clear debug points, they will be filled with each update
	//TODO: this needs to happen in the electronics update
	c.DebugPoints = make([]pixel.Vec, 0)

	// Update electronics
	outputValues := make([]OutputValue, 0, len(c.Components)*3)
	for _, component := range c.Components {
		destinations := component.ConnectedOutputs
		values := component.State.GetOutputs()
		count := len(destinations)
		if len(values) < count {
			count = len(values)
		}

		for i := 0; i < count; i++ {
			outputValues = append(outputValues, OutputValue{
				DestinationComponent: destinations[i],
				Value:                values[i],
			})
		}
	}

	for _, component := range c.Components {
		inputs, connected := calculateComponentInputs(component.ID, component.State.GetInputCount(), outputValues)
		component.State.SetInputs(inputs, connected)
		component.State.Update(c, objects)
	}
	steerLeft, _ := calculateComponentInputs(ComponentSteerLeft, 1, outputValues)
	steerRight, _ := calculateComponentInputs(ComponentSteerRight, 1, outputValues)
	accelerate, _ := calculateComponentInputs(ComponentAccelerate, 1, outputValues)
	brake, _ := calculateComponentInputs(ComponentBrake, 1, outputValues)

	const steerRate = math.Pi / 4
	const acceleration = 3

	//TODO: this needs to be separate from electronics logic!
	c.Rotation += steerLeft[0] * steerRate * dt
	c.Rotation -= steerRight[0] * steerRate * dt
	c.Speed += accelerate[0] * acceleration * dt
	c.Speed -= brake[0] * acceleration * dt

	dir := c.Forward().Scaled(c.Speed * dt)
	newPosition := c.Position.Add(dir)
	if !c.collidesWhenMovedTo(newPosition, objects) {
		c.Position = newPosition
	}

}

func (c *Car) collidesWhenMovedTo(pos pixel.Vec, objects *world.Objects) bool {

	carWidth := float64(5)

	// Stop at world border to contain the car
	if !objects.WorldBorder.Contains(pos) {
		return true
	}

	for _, o := range objects.Collidables {
		edges := o.Bounds().Edges()
		for _, edge := range edges {
			closestOnLine := edge.Closest(pos)
			if absDistance(closestOnLine, pos) < carWidth {
				return true
			}
		}
	}
	return false
}

func calculateComponentInputs(id int, inputCount int, outputValues []OutputValue) ([]float64, []bool) {
	inputs := make([]float64, inputCount)
	connected := make([]bool, inputCount)
	for _, value := range outputValues {
		if value.DestinationComponent.ID != id {
			continue
		}
		if value.DestinationComponent.Port < 0 ||
			value.DestinationComponent.Port >= len(inputs) {
			continue
		}

		inputs[value.DestinationComponent.Port] = maxValue(
			inputs[value.DestinationComponent.Port],
			value.Value,
		)
		connected[value.DestinationComponent.Port] = true
	}
	return inputs, connected
}

func maxValue(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

type OutputValue struct {
	DestinationComponent ComponentDestination
	Value                float64
}

type InputValue struct {
	Port  int
	Value float64
}

type UsedComponent struct {
	ID               int
	ConnectedOutputs []ComponentDestination
	State            Component
}

type ComponentDestination struct {
	ID   int
	Port int
}

type Component interface {
	Update(car *Car, objects *world.Objects)
	GetSpriteName() string

	GetDebugState() string

	GetInputCount() int
	GetOutputCount() int

	SetInputs(values []float64, connected []bool)
	GetOutputs() []float64
}
