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
	CTypeBuiltinSteering     = "builtin_steering"
	CTypeBuiltinAcceleration = "builtin_acceleration"
	CTypeBuiltinBraking      = "builtin_braking"

	CTypeAdd             = "add"
	CTypeMultiply        = "multiply"
	CTypeSubtract        = "subtract"
	CTypeRadar           = "radar"
	CTypeRadarShortrange = "radar_shortrange"
	CTypeConstant        = "constant"
	CTypeSplitSignal     = "split_signal"
	CTypeCompareEquals   = "compare_equals"
)

var Definitions Defs

var ComponentMakerFuncs = map[string]func() Component{
	CTypeBuiltinSteering: func() Component {
		return &BuiltinSteering{}
	},
	CTypeBuiltinAcceleration: func() Component {
		return &BuiltinAcceleration{}
	},
	CTypeBuiltinBraking: func() Component {
		return &BuiltinBraking{}
	},

	CTypeConstant: func() Component {
		return &ConstantValue{}
	},
	CTypeSplitSignal: func() Component {
		return &SplitSignal{}
	},
	CTypeAdd: func() Component {
		return &Add{}
	},
	CTypeSubtract: func() Component {
		return &Subtract{}
	},
	CTypeMultiply: func() Component {
		return &Multiply{}
	},
	CTypeCompareEquals: func() Component {
		return &CompareEquals{}
	},
	CTypeRadar: func() Component {
		return &Radar{}
	},
	CTypeRadarShortrange: func() Component {
		return &RadarShortrange{}
	},
}

func absDistance(a, b pixel.Vec) float64 {
	return math.Abs(a.To(b).Len())
}

type Car struct {
	Position pixel.Vec
	Rotation float64
	Speed    float64

	Steering     float64
	Acceleration float64
	Braking      float64

	Components  []UsedComponent
	DebugPoints []pixel.Vec
	DebugLines  []pixel.Line
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

func (c *Car) AddComponent(id int, typeName string) {

	def, ok := Definitions.Components[typeName]
	if !ok {
		return
	}

	for i, component := range c.Components {
		if component.ID == id {
			component.TypeName = typeName
			component.State = ComponentMakerFuncs[typeName]()

			// Ensure the connections are initialized and NOT connected to anything (-1)
			component.ConnectedOutputs = make([]ComponentDestination, len(def.OutputPins))
			for i, o := range component.ConnectedOutputs {
				o.ID = -1
				component.ConnectedOutputs[i] = o
			}

			c.Components[i] = component
			return
		}
	}
	component := UsedComponent{
		ID:               id,
		TypeName:         typeName,
		State:            ComponentMakerFuncs[typeName](),
		ConnectedOutputs: make([]ComponentDestination, len(def.OutputPins)),
	}
	// Ensure the connections are initialized and NOT connected to anything (-1)
	for i, o := range component.ConnectedOutputs {
		o.ID = -1
		component.ConnectedOutputs[i] = o
	}
	c.Components = append(c.Components, component)
}

func (c *Car) RemoveComponent(id int) {
	for i, component := range c.Components {
		if component.ID == id {
			c.Components = append(c.Components[:i], c.Components[i+1:]...)
			return
		}
	}
}

func (c *Car) ConnectPorts(id, pin, targetID, targetPin int) {
	for i, component := range c.Components {
		if component.ID == id {

			if pin < 0 || pin >= len(component.ConnectedOutputs) {
				return
			}

			component.ConnectedOutputs[pin] = ComponentDestination{
				ID:  targetID,
				Pin: targetPin,
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
	// Update electronics
	//TODO: Electronics should update on separate tick
	{
		// Clear debug points, they will be filled with each update
		c.DebugPoints = make([]pixel.Vec, 0)
		c.DebugLines = make([]pixel.Line, 0)

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
			if component.ID < 0 || component.ID >= len(Definitions.Ports) {
				continue
			}
			port := Definitions.Ports[component.ID]
			def := Definitions.Components[component.TypeName]

			inputs, connected := calculateComponentInputs(component.ID, len(def.InputPins), outputValues)
			component.State.SetInputs(inputs, connected)
			component.State.Update(c, objects, port)
		}
	}

	// Apply current movement changes and collision
	{
		const acceleration = 3
		const maxSpeed = 15

		const steerRate = math.Pi / 4

		c.Rotation += c.Steering * steerRate * dt

		c.Speed += c.Acceleration * acceleration * dt
		if c.Speed > maxSpeed {
			c.Speed = maxSpeed
		}
		c.Speed -= c.Braking * acceleration * dt
		if c.Speed < 0 {
			c.Speed = 0
		}

		dir := c.Forward().Scaled(c.Speed * dt)
		newPosition := c.Position.Add(dir)
		if c.collidesWhenMovedTo(newPosition, objects) {
			c.Speed = 0
		} else {
			c.Position = newPosition
		}
	}

}

func (c *Car) collidesWhenMovedTo(pos pixel.Vec, objects *world.Objects) bool {

	carWidth := float64(10)

	// Stop at world border to contain the car
	if !objects.WorldBorder.Contains(pos) {
		return true
	}
	for _, edge := range objects.WorldBorder.Edges() {
		closestOnLine := edge.Closest(pos)
		if absDistance(closestOnLine, pos) < carWidth {
			return true
		}
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
		if value.DestinationComponent.Pin < 0 ||
			value.DestinationComponent.Pin >= len(inputs) {
			continue
		}

		inputs[value.DestinationComponent.Pin] = maxValue(
			inputs[value.DestinationComponent.Pin],
			value.Value,
		)
		connected[value.DestinationComponent.Pin] = true
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
	TypeName         string
	ConnectedOutputs []ComponentDestination
	State            Component
}

type ComponentDestination struct {
	ID  int
	Pin int
}

type Component interface {
	Update(car *Car, objects *world.Objects, port PortDefinition)

	GetDebugState() string

	SetInputs(values []float64, connected []bool)
	GetOutputs() []float64
}
