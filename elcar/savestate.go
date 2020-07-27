package elcar

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type SavedCar struct {
	Components []SavedComponent
}

type SavedComponent struct {
	ID               int
	TypeName         string
	ConnectedOutputs []ComponentDestination
}

func (c *Car) Save(filename string) error {
	saved := SavedCar{
		Components: make([]SavedComponent, len(c.Components)),
	}
	for i, comp := range c.Components {
		saved.Components[i] = SavedComponent{
			ID:               comp.ID,
			TypeName:         comp.TypeName,
			ConnectedOutputs: comp.ConnectedOutputs,
		}
	}

	dir, _ := filepath.Split(filename)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := toml.NewEncoder(file)
	err = enc.Encode(saved)
	return err
}

func (c *Car) Load(filename string) error {
	var saved SavedCar
	_, err := toml.DecodeFile(filename, &saved)
	if err != nil {
		return err
	}

	c.Components = make([]UsedComponent, len(saved.Components))
	for i, comp := range saved.Components {
		maker, ok := ComponentMakerFuncs[comp.TypeName]
		if !ok {
			return errors.New("saved car has unknown component " + comp.TypeName)
		}

		c.Components[i] = UsedComponent{
			ID:               comp.ID,
			TypeName:         comp.TypeName,
			ConnectedOutputs: comp.ConnectedOutputs,
			State:            maker(),
		}
	}
	return nil
}
