package x1

type knob struct {
	spec  *spec
	mode  int // 2 for knob
	value int // knob value between 0 - 127
}

func (k *knob) setValue(d int) {
	k.value = d
}

func knobFromSpec(spec spec) *knob {
	var mode int
	if spec.Type == "Knob" {
		mode = 2
	} else {
		return nil
	}

	return &knob{
		spec:  &spec,
		mode:  mode,
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
