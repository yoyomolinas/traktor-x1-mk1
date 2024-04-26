package main

import (
	"log"

	"github.com/yoyomolinas/drivers/x1"
)

func main() {
	d, err := x1.NewDriver(
	// x1.WithLogging(),
	// x1.WithDebug(),
	// x1.WithInspect(),
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
