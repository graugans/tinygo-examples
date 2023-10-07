package pn532

import (
	"bytes"
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// The I2C address which this device listens to.
const Address = 0x24

const (
	BUFFSIZE                 = 64
	COMMAND_SAMCONFIGURATION = 0x14
)

const (
	PN532_PREAMBLE   = 0x00
	PN532_STARTCODE1 = 0x00
	PN532_STARTCODE2 = 0xFF
	PN532_POSTAMBLE  = 0x00
)

const (
	PN532_HOSTTOPN532 = 0xD4
	PN532_PN532TOHOST = 0xD5
)

// PN532 Commands
const (
	COMMAND_GETFIRMWAREVERSION = 0x02
)

const (
	PN532_I2C_READY = 0x01
)

// Device wraps an I2C connection to a PN532 device.
type Device struct {
	bus          drivers.I2C
	Address      uint16
	rst          machine.Pin
	irq          machine.Pin
	packetbuffer []byte
}

// New creates a new PN532 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(irq machine.Pin, reset machine.Pin, bus drivers.I2C) Device {
	return Device{
		bus:          bus,
		Address:      Address,
		irq:          irq,
		rst:          reset,
		packetbuffer: make([]byte, BUFFSIZE),
	}
}

func (d *Device) Configure() {
	d.rst.Configure(machine.PinConfig{
		Mode: machine.PinOutput,
	})
}

func (d *Device) reset() {
	d.rst.Low()
	time.Sleep(1 * time.Millisecond) // min 20ns
	d.rst.High()
	time.Sleep(2 * time.Millisecond) // max 2ms
}

func (d *Device) wakeup() {
	d.samconfig()
}

func (d *Device) samconfig() error {
	buffer := make([]byte, 4)
	buffer[0] = COMMAND_SAMCONFIGURATION
	buffer[1] = 0x01 // normal mode
	buffer[2] = 0x14 // timeout 50ms * 20 = 1 second
	buffer[3] = 0x01 // use IRQ PIN!

	if err := d.sendCommandCheckAck(buffer, 100*time.Millisecond); err != nil {
		fmt.Printf("Error while sending command %s\n", err)
	}
	buffer = make([]byte, 9)
	if err := d.readdata(buffer); err != nil {
	}
	const offset = 6
	if buffer[offset] != 0x15 {
		return fmt.Errorf("Invalid response %d", buffer[offset])
	}
	return nil
}

func (d *Device) sendCommandCheckAck(command []byte, timeout time.Duration) error {
	// write the command
	d.writecommand(command)
	if !d.waitready(timeout) {
		return fmt.Errorf("waitready failed")
	}
	d.i2cTuning()
	if !d.readack() {
		return fmt.Errorf("readack failed")
	}
	d.i2cTuning()
	if !d.waitready(timeout) {
		return fmt.Errorf("second waitready failed")
	}
	return nil
}

func (d *Device) writecommand(cmd []byte) {
	packet := make([]byte, 8+len(cmd))
	LEN := byte(len(cmd) + 1)
	packet[0] = PN532_PREAMBLE
	packet[1] = PN532_STARTCODE1
	packet[2] = PN532_STARTCODE2
	packet[3] = LEN
	packet[4] = ^LEN + 1
	packet[5] = PN532_HOSTTOPN532
	var sum byte = 0
	for i := 0; i < len(cmd); i++ {
		packet[6+i] = cmd[i]
		sum += cmd[i]
	}
	packet[6+len(cmd)] = ^(PN532_HOSTTOPN532 + sum) + 1
	packet[7+len(cmd)] = PN532_POSTAMBLE
	if err := d.bus.Tx(d.Address, packet, nil); err != nil {
		fmt.Printf("Error while sending command %s\n", err)
	}
}

func (d *Device) waitready(timeout time.Duration) bool {
	const delay = 10 * time.Millisecond
	timer := 0 * time.Millisecond
	for !d.isReady() {
		if timeout != 0 {
			timer += 10 * time.Millisecond
			if timer > timeout {
				return false
			}
		}
		time.Sleep(delay)
	}
	return true
}

func (d *Device) readack() bool {
	pn532ack := []byte{
		0x00, 0x00, 0xFF,
		0x00, 0xFF, 0x00,
	} ///< ACK message from PN532

	ackbuff := make([]byte, 6)
	err := d.readdata(ackbuff)
	if err != nil {
		fmt.Printf("Error received: %s\n", err)
		return false
	}
	d.printBuffer("ACK", ackbuff)
	return bytes.Equal(ackbuff, pn532ack)
}

func (d *Device) printBuffer(name string, buffer []byte) {
	fmt.Printf("%s buffer: [ ", name)
	for _, b := range buffer {
		fmt.Printf("0x%02X ", b)
	}
	fmt.Printf("]\n")
}

func (d *Device) readdata(buffer []byte) error {
	rxBuffer := make([]byte, len(buffer)+1)
	d.bus.Tx(d.Address, nil, rxBuffer)
	for i := 0; i < len(buffer); i++ {
		buffer[i] = rxBuffer[i+1]
	}
	return nil
}

func (d *Device) isReady() bool {
	rdy := make([]byte, 1)
	d.bus.Tx(d.Address, nil, rdy)
	return rdy[0] == PN532_I2C_READY
}

func (d *Device) i2cTuning() {
	// I2C delay
	time.Sleep(1 * time.Millisecond)
}

func (d *Device) FirmwareVersion() (uint32, error) {
	buffer := make([]byte, 1)
	buffer[0] = COMMAND_GETFIRMWAREVERSION
	err := d.sendCommandCheckAck(buffer, 100*time.Millisecond)
	if err != nil {
		return 0, err
	}

	buffer = make([]byte, 13)
	err = d.readdata(buffer)
	if err != nil {
		return 0, err
	}
	d.printBuffer("Firmware", buffer)

	expFirmwareVersion := []byte{
		0x00, 0x00, 0xFF,
		0x06, 0xFA, 0xD5,
	}

	if !bytes.Equal(buffer[0:len(expFirmwareVersion)], expFirmwareVersion) {
		return 0, fmt.Errorf("Invalid response received")
	}

	var response uint32
	response = uint32(buffer[7])
	response <<= 8
	response |= uint32(buffer[8])
	response <<= 8
	response |= uint32(buffer[9])
	response <<= 8
	response |= uint32(buffer[10])
	return response, nil
}
