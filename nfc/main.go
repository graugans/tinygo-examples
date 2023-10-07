package main

import (
	"fmt"
	"machine"

	"github.com/graugans/tinygo-examples/drivers/pn532"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
	})
	nfc := pn532.New(machine.I2C0)

	version, err := nfc.FirmwareVersion()
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	fmt.Printf("PN532 firmware version: %d\n", version)
}
