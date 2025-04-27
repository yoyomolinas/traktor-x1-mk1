package driver

type knob struct {
	spec  *spec
	value int // knob value between 0 - 127
}

func (k *knob) setValue(d int) {
	k.value = d
}

func knobFromSpec(spec spec) *knob {
	if spec.Type != "Knob" {
		return nil
	}

	return &knob{
		spec:  &spec,
		value: 0,
	}
}

func knobsFromSpec(specs []spec) []*knob {
	knobs := []*knob{}

	for i := range specs {
		if k := knobFromSpec(specs[i]); k != nil {
			knobs = append(knobs, k)
		}
	}

	return knobs
}
