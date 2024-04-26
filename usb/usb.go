package usb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/gousb"
)

// Device is a wrapper around gousb.Device
type Device struct {
	ctx             *gousb.Context
	usbDevice       *gousb.Device
	usbConfig       *gousb.Config
	usbInterface    *gousb.Interface
	serialNumber    string
	deviceConfig    DeviceConfig
	inputEndpoints  map[int]*gousb.InEndpoint
	outputEndpoints map[int]*gousb.OutEndpoint
}

type DeviceConfig struct {
	USBConfigID             int
	USBInterfaceID          int
	USBAlternativeSettingID int
	ProductID               uint16
	VendorID                uint16
}

func NewDevice(cfg DeviceConfig) (device *Device, err error) {
	ctx := gousb.NewContext()
	defer func() {
		if err != nil {
			ctx.Close()
		}
	}()

	// Open Traktor X1.
	usbDevice, err := ctx.OpenDeviceWithVIDPID(gousb.ID(cfg.VendorID), gousb.ID(cfg.ProductID))
	if err != nil {
		return nil, fmt.Errorf("failed to open a device: %w", err)
	}

	if usbDevice == nil {
		return nil, fmt.Errorf("device not found")
	}

	defer func() {
		if err != nil {
			usbDevice.Close()
		}
	}()

	serialNumber, err := usbDevice.SerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to get serial number: %w", err)
	}

	usbConfig, err := usbDevice.Config(cfg.USBConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to claim configuration: %w", err)
	}

	// Claim the given interface.
	usbIface, err := usbConfig.Interface(cfg.USBInterfaceID, cfg.USBAlternativeSettingID)
	if err != nil {
		return nil, fmt.Errorf("failed to claim interface: %w", err)
	}

	defer func() {
		if err != nil {
			usbIface.Close()
		}
	}()

	device = &Device{
		ctx:             ctx,
		usbDevice:       usbDevice,
		usbConfig:       usbConfig,
		usbInterface:    usbIface,
		serialNumber:    serialNumber,
		deviceConfig:    cfg,
		inputEndpoints:  map[int]*gousb.InEndpoint{},
		outputEndpoints: map[int]*gousb.OutEndpoint{},
	}

	return
}

func (d *Device) Close() {
	d.ctx.Close()
	d.usbDevice.Close()
	d.usbConfig.Close()
	d.usbInterface.Close()
}

func (d *Device) Read(
	ctx context.Context,
	endpointAddress int,
	timeout time.Duration,
	v []byte,
) error {
	// TODO: make this thread safe.

	var in *gousb.InEndpoint
	var err error

	in, ok := d.inputEndpoints[endpointAddress]
	if !ok {
		// Open a new input endpoint to read from.
		in, err = d.usbInterface.InEndpoint(endpointAddress)
		if err != nil {
			return fmt.Errorf("failed to open input endpoint with address %d: %v", endpointAddress, err)
		}

		d.inputEndpoints[endpointAddress] = in
	}

	// Create the context with timeout.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // releases resources if this operation completes before timeout.

	n, err := in.ReadContext(ctx, v)
	if err != nil {
		if errors.Is(err, gousb.TransferCancelled) {
			return fmt.Errorf("context cancelled: %w", ReadCancelledError{})
		}

		return fmt.Errorf("failed to read from device: %w", err)
	}

	if n != len(v) {
		// Partial data read is a thing and it disrupts our expectations.
		return fmt.Errorf("failed to read from context, invalid number of bytes read: expected %d got %d: %w", len(v), n, InsufficientBytesReadError{})
	}

	return nil
}

func (d *Device) Write(
	ctx context.Context,
	endpointAddress int,
	timeout time.Duration,
	v []byte,
) error {
	// TODO: make this thread safe.

	var out *gousb.OutEndpoint
	var err error

	out, ok := d.outputEndpoints[endpointAddress]
	if !ok {
		// Open a new output endpoint to read from.
		out, err = d.usbInterface.OutEndpoint(endpointAddress)
		if err != nil {
			return fmt.Errorf("failed to open output endpoint with address %d: %v", endpointAddress, err)
		}

		d.outputEndpoints[endpointAddress] = out
	}

	// Create the context with timeout.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // releases resources if this operation completes before timeout.

	if _, err := out.WriteContext(ctx, v); err != nil {
		return fmt.Errorf("failed to write to device: %w", err)
	}

	return nil
}

type InsufficientBytesReadError struct{}

func (e InsufficientBytesReadError) Error() string {
	return "insufficient bytes read"
}

var _ = error(InsufficientBytesReadError{})

type ReadCancelledError struct{}

func (e ReadCancelledError) Error() string {
	return "context cancelled during read"
}

var _ = error(ReadCancelledError{})
