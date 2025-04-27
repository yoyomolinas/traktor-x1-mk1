package driver

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed specs/spec.yaml
var f embed.FS

type spec struct {
	Name        string `yaml:"name"`        // the button description
	Type        string `yaml:"type"`        // Toggle, Encoder, Hold or Hotcue
	BufferIndex int    `yaml:"bufIdx"`      // buffer index
	LEDIndex    int    `yaml:"ledIdx"`      // led index
	OnMIDICC    int    `yaml:"onMidiCC"`    // turn on midi cc
	OffMIDICC   int    `yaml:"offMidiCC"`   // turn off midi cc
	OnVelocity  int    `yaml:"onVelocity"`  // turn on velocity button
	OffVelocity int    `yaml:"offVelocity"` // turn off velocity button
}

func readSpecs() ([]spec, error) {
	data, err := f.ReadFile("specs/spec.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg []spec
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
