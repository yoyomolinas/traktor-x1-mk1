package main

import (
	"log"

	"github.com/yoyomolinas/driver"
)

func main() {
	d, err := driver.NewDriver(
		driver.WithLogging(),

		// Inspect low level usb responses from the device.
		// driver.WithInspect(),
	)
	if err != nil {
		log.Fatalf("failed to create driver: %v", err)
	}
	defer d.Close()

	for {
		if err := d.NextState(); err != nil {
			log.Printf("next state failed: %v", err)

		}
	}
}
