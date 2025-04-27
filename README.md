# Traktor X1 MK1 Driver (Mac Silicon + MIDI Conversion)

This project provides a macOS driver and MIDI converter for the Native Instruments Traktor X1 MK1 controller.  
Native Instruments discontinued support for Mac Silicon, so this tool revives X1 MK1 usability for modern Macs.

In addition to basic driver functionality, it also **converts X1 MK1 into a fully configurable MIDI controller**, enabling use in DAWs like Ableton Live.

Inspired by research from the [x1-mk1-usb2midi Rust project](https://github.com/Opa-/x1-mk1-usb2midi).

---

## Features
- Full driver support for Traktor X1 MK1 on Mac Silicon.
- Converts X1 MK1 into a MIDI controller.
- Easily customizable MIDI mappings for each button, knob, and encoder.
- Supports a **Shift Mode** for a second layer of mappings.

---

## How It Works

### USB Communication
- Data is read from the device using the IN endpoint address `0x84`.
- LED states are sent to the device via the WRITE endpoint address `0x01`.
- After each LED update, an unlock read is performed on endpoint `0x81` to confirm write readiness.

The input USB buffer is 24 bytes long:
- Bytes 1–6 represent boolean button states.
- Other bytes contain values for knobs and encoders.

### LED Handling
Each button with an LED is assigned a `ledIdx`.  
An LED buffer of 32 bytes is built for updates:
- Byte 0 is a descriptor (0x0C) for LED control (Consumer Page).
- Bytes 1–31 correspond to individual LED brightness levels.

This buffer is sent to the device every frame to update lighting based on interaction.

### Debugging, Logging, and Inspection
The driver provides optional debug functionality during initialization:
- `WithLogging()`: Logs button presses, knob movement, and MIDI messages.
- `WithInspect()`: Prints low-level USB input inspection data.

Example:

```go
driver, err := driver.NewDriver(driver.WithLogging(), driver.WithInspect())
```

---

## Spec Format

The device configuration is managed through a YAML spec file (`spec.yaml`).  
Each entry in the spec defines behavior for a control.

| Field        | Description |
| ------------ | ----------- |
| `name`       | Human-readable name of the control. |
| `type`       | One of: `"Shift"`, `"Toggle"`, `"Hold"`, `"Knob"`, `"Encoder"`. |
| `bufIdx`     | Index of this control in the USB input stream. |
| `ledIdx`     | LED index in the USB stream (only for buttons). |
| `onMidiCC`   | MIDI CC to send when activated (only for buttons). |
| `offMidiCC`  | MIDI CC to send when deactivated (only for buttons). |
| `onVelocity` | Velocity when activated (only for buttons). |
| `offVelocity`| Velocity when deactivated (only for buttons). |

---

## Special Modes

- **Shift Mode**:  
  Holding the "Shift" button unlocks a secondary layer of mappings.  
  MIDI signals in Shift mode are sent on **channel 9** instead of **channel 8** (default).

---

## How to Modify the Spec
- Open `spec.yaml`.
- Add, edit, or remove entries to define which controls send which MIDI signals.
- Restart the driver after modifying the spec.

This allows you to fine-tune the behavior of your X1 MK1 for any DAW or live performance setup.

---

## License

This project is open-source under the MIT License.