package x1

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed spec.yaml
var f embed.FS

type spec struct {
	Name       string `yaml:"name"`   // the button description
	Type       string `yaml:"type"`   // Toggle, Encoder, Hold or Hotcue
	BuferIndex int    `yaml:"bufIdx"` // buffer index
	LEDIndex   int    `yaml:"ledIdx"` // led index
	MIDICC     int    `yaml:"midiCC"` // midi control change mapping
}

func readSpecs() ([]spec, error) {
	data, err := f.ReadFile("spec.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg []spec
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
