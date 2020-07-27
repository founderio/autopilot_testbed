package elcar

import "github.com/faiface/pixel"

type Defs struct {
	Components map[string]ComponentDefinition
	Ports      []PortDefinition
}

type PortKind string

const (
	PortKindSensor  PortKind = "sensor"
	PortKindChip    PortKind = "chip"
	PortKindBuiltin PortKind = "builtin"
)

type PortDefinition struct {
	HoodPosition  pixel.Vec
	WorldPosition pixel.Vec
	Direction     pixel.Vec

	PortKind PortKind
	Prefill  string
}

type ComponentDefinition struct {
	Usable     bool
	PortKind   PortKind
	InputPins  []PinDefinition
	OutputPins []PinDefinition
}

type PinDefinition struct {
	Position pixel.Vec
}

func GetOutPinPosition(typeName string, port int) pixel.Vec {
	def, ok := Definitions.Components[typeName]
	if !ok {
		return pixel.ZV
	}

	if port < 0 || port >= len(def.OutputPins) {
		return pixel.ZV
	}

	return def.OutputPins[port].Position
}

func IsComponentAllowedInSlot(id int, typeName string) bool {
	if id < 0 || id >= len(Definitions.Ports) {
		return false
	}
	portDef := Definitions.Ports[id]
	componentDef, ok := Definitions.Components[typeName]
	if !ok {
		return false
	}
	return portDef.PortKind == componentDef.PortKind
}
