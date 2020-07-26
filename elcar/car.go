package elcar

import (
	"math"

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
	CTypeRadar    = "radar"
	CTypeConstant = "constant"
)

type Car struct {
	Position pixel.Vec
	Rotation float64
	Speed    float64

	Components []UsedComponent
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

func (c *Car) Update(dt float64) {
	dir := pixel.Unit(c.Rotation).Scaled(c.Speed * dt)
	c.Position = c.Position.Add(dir)

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
		inputs := calculateComponentInputs(component.ID, component.State.GetInputCount(), outputValues)
		component.State.SetInputs(inputs)
		component.State.Update()
	}
	steerLeft := calculateComponentInputs(ComponentSteerLeft, 1, outputValues)[0]
	steerRight := calculateComponentInputs(ComponentSteerRight, 1, outputValues)[0]
	accelerate := calculateComponentInputs(ComponentAccelerate, 1, outputValues)[0]
	brake := calculateComponentInputs(ComponentBrake, 1, outputValues)[0]

	const steerRate = math.Pi / 4
	const acceleration = 3

	//TODO: this needs to be separate from electronics logic!
	c.Rotation += steerLeft * steerRate * dt
	c.Rotation -= steerRight * steerRate * dt
	c.Speed += accelerate * acceleration * dt
	c.Speed -= brake * accelerate * dt
}

func calculateComponentInputs(id int, inputCount int, outputValues []OutputValue) []float64 {
	inputs := make([]float64, inputCount)
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
	}
	return inputs
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
	Update()
	GetSpriteName() string

	GetInputCount() int
	GetOutputCount() int

	SetInputs(values []float64)
	GetOutputs() []float64
}
