# Traktor X1 MK1 Driver (Mac Silicon + MIDI Conversion)

This project provides a macOS driver and MIDI converter for the Native Instruments Traktor X1 MK1 controller.  
Native Instruments discontinued support for Mac Silicon, so this tool revives X1 MK1 usability for new Macs.

In addition to basic driver functionality, it also **converts X1 MK1 into a fully configurable MIDI controller**, enabling use in DAWs like Ableton Live.

Inspired by the research from the [x1-mk1-usb2midi Rust project](https://github.com/Opa-/x1-mk1-usb2midi).

---

## Features
- Full driver support for Traktor X1 MK1 on Mac Silicon.
- Converts X1 MK1 to a MIDI controller.
- Easily customizable MIDI mappings for each button, knob, and encoder.
- Supports a **Shift Mode** to access a second layer of MIDI mappings.

---

## Spec Format

The device configuration is managed through a YAML spec file (`spec.yaml`).  
Each entry in the spec defines the behavior of a specific button, knob, or encoder.

| Field        | Description |
| ------------ | ----------- |
| `name`       | Human-readable name of the control. |
| `type`       | One of: `"Shift"`, `"Toggle"`, `"Hold"`, `"Knob"`, `"Encoder"`. |
| `bufIdx`     | Low-level index of this control in the USB input stream. |
| `ledIdx`     | LED index in the USB stream (only for buttons). |
| `onMidiCC`   | MIDI CC to send when the control is activated (only for buttons). |
| `offMidiCC`  | MIDI CC to send when the control is deactivated (only for buttons). |
| `onVelocity` | Velocity when activated (only for buttons). |
| `offVelocity`| Velocity when deactivated (only for buttons). |

---

## Special Modes
- **Shift Mode**:  
  Holding the "Shift" button unlocks a secondary layer of mappings.  
  In Shift mode, all MIDI signals are sent on **channel 9** instead of **channel 8** (default).

---

## How to Modify the Spec
- Open `spec.yaml`.
- Add, edit, or remove entries to define which controls send which MIDI signals.
- After modifying the spec, restart the driver to load the new configuration.

Each change allows fine-tuned control of your X1 MK1 to match your DAW workflow or live performance setup.

/*
Package driver implements a USB and MIDI driver for the Native Instruments Traktor X1 MK1 controller on Mac Silicon.

## Overview

The driver connects to the X1 MK1 via USB and emulates a MIDI controller. It uses `usb` for raw device communication and `rtmididrv` to send virtual MIDI signals to the operating system.

## USB Communication

- Data is read from the device using the IN endpoint address `0x84`.
- LED states are sent to the device via the WRITE endpoint address `0x01`.
- An additional unlock read is performed on `0x81` after every LED update to confirm write readiness.

The input USB buffer is 24 bytes long:
- Bytes 1-6: Boolean button states, flattened into an array of true/false values.
- Other bytes contain values for knobs and encoders.

## LED Handling

Each button with an LED is assigned a `LEDIndex`.
A LED buffer of 32 bytes is built on every update:
- First byte (0x0C) is a descriptor for LED control (Consumer Page).
- Following 31 bytes represent individual LED values.

The LED buffer is written back to the device every frame to update the button lighting states based on user interaction.

## Debugging / Logging / Inspection Options

The driver supports functional options during creation:

- `WithLogging()`: Enables logging of button presses, MIDI signals, and knob changes to the console.
- `WithInspect()`: Enables low-level USB input inspection, printing raw button state changes.

To enable, pass the options when constructing the driver:

	driver, err := driver.NewDriver(driver.WithLogging(), driver.WithInspect())

This makes debugging and development easier.

*/

---

## License
This project is open-source under the MIT License.