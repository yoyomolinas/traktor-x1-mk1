package driver

type button struct {
	spec    *spec
	led     *led
	on      bool
	mode    int  // 0 for toggle, 1 for hold, 2 for knob, 3 for shift
	pressed bool // is button currently pressed?
}

func (t *button) On() {
	t.on = true
	t.led.On()
}

func (t *button) Off() {
	t.on = false
	t.led.Off()
}

func buttonFromSpec(spec spec) *button {
	var mode int
	if spec.Type == "Toggle" {
		mode = 0
	} else if spec.Type == "Hold" {
		mode = 1
	} else if spec.Type == "Shift" {
		mode = 3
	} else {
		return nil
	}

	return &button{
		spec: &spec,
		led:  newLed(),
		mode: mode,
		on:   false,
	}
}

func buttonsFromSpecs(specs []spec) []*button {
	buttons := []*button{}

	// for specs with "Toggle" or "Hold" type there is a button
	for i := range specs {
		if b := buttonFromSpec(specs[i]); b != nil {
			buttons = append(buttons, b)
		}
	}

	return buttons
}
