package pn532

import (
	"machine"

	"tinygo.org/x/drivers"
)

// The I2C address which this device listens to.
const Address = 0x24

// Device wraps an I2C connection to a PN532 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	reset   machine.Pin
	irq     machine.Pin
}

// New creates a new PN532 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(irq machine.Pin, reset machine.Pin, bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		irq:     irq,
		reset:   reset,
	}
}

func (d *Device) FirmwareVersion() (int, error) {
	return 0, nil
}
