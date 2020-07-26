package elcar

type AddComponent struct {
	inputs []float64
	value  float64
}

func (c *AddComponent) Update() {
	var newValue float64
	for _, val := range c.inputs {
		newValue += val
	}
	c.value = newValue
}
func (c *AddComponent) GetSpriteName() string {
	return CTypeAdd
}

func (c *AddComponent) GetInputCount() int {
	return 3
}
func (c *AddComponent) GetOutputCount() int {
	return 1
}

func (c *AddComponent) SetInputs(values []float64) {
	c.inputs = values
}
func (c *AddComponent) GetOutputs() []float64 {
	return []float64{c.value}
}

type ConstantValue struct {
}

func (c *ConstantValue) GetSpriteName() string {
	return CTypeConstant
}

func (c *ConstantValue) Update() {
}

func (c *ConstantValue) GetInputCount() int {
	return 0
}
func (c *ConstantValue) GetOutputCount() int {
	return 1
}

func (c *ConstantValue) SetInputs(values []float64) {
}
func (c *ConstantValue) GetOutputs() []float64 {
	return []float64{1}
}

type ComponentRadar struct {
}

func (c *ComponentRadar) GetSpriteName() string {
	return CTypeRadar
}

func (c *ComponentRadar) Update() {
}

func (c *ComponentRadar) GetInputCount() int {
	return 0
}
func (c *ComponentRadar) GetOutputCount() int {
	return 1
}

func (c *ComponentRadar) SetInputs(values []float64) {
}
func (c *ComponentRadar) GetOutputs() []float64 {
	return []float64{1}
}
