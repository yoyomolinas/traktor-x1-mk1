package driver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/yoyomolinas/usb"
	mididrivers "gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var writeEndpointAddress = 0x01  // write
var unlockEndpointAddress = 0x81 // unlocks writes
var inEndpointAdress = 0x84      // read

type Driver struct {
	device     *usb.Device
	midiDriver *rtmididrv.Driver
	midiOut    mididrivers.Out
	specs      []spec

	// Keep track of device state.
	buttons map[int][]*button
	knobs   []*knob
	hotcue  bool
	shift   bool

	timeout time.Duration

	// enable / disable various tools when running the driver.
	log     bool
	inspect bool
}

type DriverOption func(*Driver)

func WithLogging() DriverOption {
	return func(d *Driver) {
		d.log = true
	}
}

// WithInspect is a functional option to observe low level usb responses from X1.
func WithInspect() DriverOption {
	return func(d *Driver) {
		d.inspect = true
	}
}

func NewDriver(options ...DriverOption) (*Driver, error) {
	deviceConfig := usb.DeviceConfig{
		USBConfigID:             1,
		USBInterfaceID:          0,
		USBAlternativeSettingID: 0,
		VendorID:                0x17cc,
		ProductID:               0x2305,
	}

	device, err := usb.NewDevice(deviceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open new device: %v", err)
	}

	specs, err := readSpecs()
	if err != nil {
		return nil, fmt.Errorf("failed to get specs: %w", err)
	}

	midiDriver, err := rtmididrv.New()
	if err != nil {
		log.Fatal(err)
	}

	midiOut, err := midiDriver.OpenVirtualOut("Traktor X1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Midi Device: %v\n", midiOut)

	driver := &Driver{
		device:     device,
		midiDriver: midiDriver,
		midiOut:    midiOut,
		specs:      specs,
		timeout:    50 * time.Millisecond,
		// User modes for buttons.
		buttons: map[int][]*button{
			0: buttonsFromSpecs(specs),
			1: buttonsFromSpecs(specs), // SHIFT
		},
		knobs:  knobsFromSpec(specs),
		shift:  false,
		hotcue: false,
	}

	for _, option := range options {
		option(driver)
	}

	return driver, nil
}

func (d *Driver) Close() {
	d.device.Close()
	d.midiDriver.Close()
	d.midiOut.Close()
}

func (d *Driver) NextState() error {
	inBuf := make([]byte, 24)
	if err := d.device.Read(context.TODO(), inEndpointAdress, d.timeout, inBuf); err != nil {
		if errors.Is(err, usb.InsufficientBytesReadError{}) {
			return nil
		} else if errors.Is(err, usb.ReadCancelledError{}) {
			return nil
		}

		return fmt.Errorf("failed to read from device: %w", err)
	}

	if err := d.updateFromBuffer(inBuf); err != nil {
		return fmt.Errorf("failed to update state from input buffer: %w", err)
	}

	ledBuf, err := d.createLedBufferFromButtons()
	if err != nil {
		return fmt.Errorf("failed to create led buffer: %w", err)
	}

	if err := d.device.Write(context.TODO(), writeEndpointAddress, d.timeout, ledBuf); err != nil {
		return fmt.Errorf("failed to write to device: %w", err)
	}

	unlockBuf := make([]byte, 1)
	if err := d.device.Read(context.TODO(), unlockEndpointAddress, d.timeout, unlockBuf); err != nil {
		if errors.Is(err, usb.InsufficientBytesReadError{}) {
			return nil
		} else if errors.Is(err, usb.ReadCancelledError{}) {
			return nil
		}

		return fmt.Errorf("failed to unlock device: %w", err)
	}

	return nil
}

func (d *Driver) updateFromBuffer(in []byte) error {
	if len(in) != 24 {
		return fmt.Errorf("invalid number of bytes in buffer, expecte 24 got %d", len(in))
	}

	// Boolean button states live in the first 5 bytes, from 1 -> 6. These are the pressable
	// buttons on the X1.
	buttonStates := flatBytes(in[1:6])

	if d.inspect {
		// Inspect the buttons.
		for i, s := range buttonStates {
			if s {
				fmt.Printf("pressed button: %d\n", i)
			}
		}
	}

	mode := d.currentMode()

	for i, b := range d.buttons[mode] {
		if b.spec.BufferIndex > len(buttonStates) {
			return fmt.Errorf("button with index %d is overflowing the buffer of length %d", b.spec.BufferIndex, len(buttonStates))
		}

		switch b.mode {

		case 0: // TOGGLE
			pressed := buttonStates[b.spec.BufferIndex]

			// We switch states on state changes, i.e. lift finger or press, and we do nothing in between, i.e.
			// when you keep pressing the same button.
			if b.pressed != pressed {
				if pressed {
					// Toggle when the button is pressed and do nothing when it lifts.
					if b.on { // Turn off if on.
						d.buttons[mode][i].Off()
						if err := d.midiSend(d.midiOut, d.currentMidiChannel(), b.spec.OffMIDICC, b.spec.OffVelocity); err != nil {
							return fmt.Errorf("failed to send midi: %w", err)
						}

						if d.log {
							fmt.Printf("%s turned off\n", b.spec.Name)
						}
					} else { // Turn on if off.
						d.buttons[mode][i].On()
						if err := d.midiSend(d.midiOut, d.currentMidiChannel(), b.spec.OnMIDICC, b.spec.OnVelocity); err != nil {
							return fmt.Errorf("failed to send midi: %w", err)
						}

						if d.log {
							fmt.Printf("%s turned on\n", b.spec.Name)
						}
					}
				}

				d.buttons[mode][i].pressed = pressed
			}

		case 1: // HOLD
			pressed := buttonStates[b.spec.BufferIndex]
			if pressed {
				if !b.on {
					d.buttons[mode][i].On()
					if err := d.midiSend(d.midiOut, d.currentMidiChannel(), b.spec.OnMIDICC, b.spec.OnVelocity); err != nil {
						return fmt.Errorf("failed to send midi: %w", err)
					}

					if d.log {
						fmt.Printf("%s turned on\n", b.spec.Name)
					}
				}
			} else {
				if b.on {
					d.buttons[mode][i].Off()
					if err := d.midiSend(d.midiOut, d.currentMidiChannel(), b.spec.OffMIDICC, b.spec.OffVelocity); err != nil {
						return fmt.Errorf("failed to send midi: %w", err)
					}

					if d.log {
						fmt.Printf("%s turned off\n", b.spec.Name)
					}
				}
			}
		case 3: // SHIFT
			pressed := buttonStates[b.spec.BufferIndex]
			if pressed {
				if !b.on {
					d.buttons[mode][i].On()

					if d.log {
						fmt.Printf("%s turned on\n", b.spec.Name)
					}
				}
			} else {
				if b.on {
					d.buttons[mode][i].Off()

					if d.log {
						fmt.Printf("%s turned off\n", b.spec.Name)
					}
				}
			}

			d.shift = d.buttons[mode][i].on
		}
	}

	for i, k := range d.knobs {
		prev := k.value
		cur, err := d.knobValueFromBuffer(k.spec, in)
		if err != nil {
			return fmt.Errorf("failed to get knob value: %w", err)
		}

		if prev != cur {
			d.knobs[i].setValue(cur)
			if err := d.midiSend(d.midiOut, d.currentMidiChannel(), k.spec.OnMIDICC, cur); err != nil {
				return fmt.Errorf("failed to send midi: %w", err)
			}

			if d.log {
				fmt.Printf("%s value changed %d\n", d.knobs[i].spec.Name, d.knobs[i].value)
			}

		}

	}

	return nil
}

// createLedBuffer creates a byte buffer for LED's which can be written to the device.
func (d *Driver) createLedBufferFromButtons() ([]byte, error) {
	ledBuf := make([]byte, 32)

	// Collect LED's from buttons
	leds := make([]led, 31)

	mode := d.currentMode()

	for i := range d.buttons[mode] {
		// Only consider buttons with an associated led light.
		if ledIdx := d.buttons[mode][i].spec.LEDIndex; ledIdx != 0 {
			if ledIdx > len(leds) {
				return nil, fmt.Errorf("invalid led index: %d", ledIdx)
			}

			// E.g. led with index 31 is equivalent to leds[30].
			leds[ledIdx-1] = *d.buttons[mode][i].led
		}

	}

	// First byte is the descriptor and points to the Consumer Page to control the LED's in Traktor X1.
	// https://www.usbzh.com/article/detail-982.html
	ledBuf[0] = 0x0C

	for i, l := range leds {
		ledBuf[i+1] = l.val
	}

	return ledBuf, nil
}

func (d *Driver) currentMidiChannel() int {
	// Use 8 or 9, not for any particular reason, just to avoid overlapping with any other midi devices
	// that might be connected.
	return 0xB7 + d.currentMode()
}

func (d *Driver) currentMode() int {
	mode := 0
	if d.shift {
		mode = 1
	}

	return mode
}

// Disco randomly turns LEDs on and off.
func (d *Driver) Disco() {
	mode := d.currentMode()
	for i := range d.buttons[mode] {
		if rand.Float32() < 0.5 {
			d.buttons[mode][i].Off()
		} else {
			d.buttons[mode][i].On()
		}
	}
}

func (d *Driver) knobValueFromBuffer(spec *spec, buf []byte) (int, error) {
	major := float64(buf[spec.BufferIndex])
	decimal := float64(buf[spec.BufferIndex+1]) / 256
	result := ((major + decimal) / 16) * 127
	return int(math.Round(result)), nil
}

func (d *Driver) midiSend(out mididrivers.Out, channel, cc int, velocity int) error {
	if d.log {
		fmt.Printf("MIDI Signal: Channel=%d CC=%d Velocity=%d\n", channel, cc, velocity)
	}

	if err := out.Send([]byte{byte(channel), byte(cc), byte(velocity)}); err != nil {
		return fmt.Errorf("failed to send midi: %w", err)
	}

	return nil
}
