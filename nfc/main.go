package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/graugans/tinygo-examples/drivers/pn532"
)

func main() {
	const delay = 5
	for i := 0; i < delay; i++ {
		time.Sleep(time.Second) // allow to attach the monitor
		fmt.Printf("Sleeping %d/%d \n", i, delay)
	}

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
	})
	nfc := pn532.New(machine.GP17, machine.GP16, machine.I2C0)
	nfc.Configure()
	version, err := nfc.FirmwareVersion()
	if err != nil {
		fmt.Printf("Error: %s \n", err.Error())
	}

	fmt.Printf("Found Chip PN5%02X\n", version.IC)
	fmt.Printf("Firmware version: %d.%d\n", version.Ver, version.Rev)
	fmt.Printf("Firmware Support: 0x%02X\n", version.Support)
}
