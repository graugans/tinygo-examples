//go:build tinygo

// Package pn532 provides a driver for the NXP PN532 chip
//
// [1] Datasheet PN532: https://www.nxp.com/docs/en/nxp/data-sheets/PN532_C1.pdf
// [2] User Manual PN532: https://www.nxp.com/docs/en/user-guide/141520.pdf
// Adafruit C++ Driver: https://github.com/adafruit/Adafruit-PN532

package pn532

import (
	"bytes"
	"encoding/hex"
	"errors"
	"machine"
	"strconv"
	"time"
)

// The version information of the embedded firmware.
type FirmwareVersion struct {
	IC      uint8 // Version of the IC. For PN532, the content of this byte is 0x32,
	Ver     uint8 // Version of the firmware
	Rev     uint8 // Revision of the firmware
	Support uint8 // Indicates which are the functionalities supported by the firmware
}

func (ver *FirmwareVersion) String() string {
	res := "Found Chip PN5" + hex.EncodeToString([]byte{ver.IC}) + "\n"
	res += "Firmware version: " + strconv.Itoa(int(ver.Ver)) + "." + strconv.Itoa(int(ver.Rev)) + "\n"
	res += "Firmware Support: 0x" + hex.EncodeToString([]byte{ver.IC}) + "\n"
	return res
}

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
	COMMAND_GETFIRMWAREVERSION  = 0x02
	COMMAND_INLISTPASSIVETARGET = 0x4A // List passive targets
	COMMAND_INDATAEXCHANGE      = 0x40 // Data exchange
)

const (
	PN532_I2C_READY = 0x01
)

const (
	MIFARE_ISO14443A = 0x00
)

// Device wraps an I2C connection to a PN532 device.
type Device struct {
	bus             *machine.I2C
	address         uint16
	debug           bool
	buffer          [BUFFSIZE]byte
	txBuffer        [BUFFSIZE]byte
	rxBuffer        [BUFFSIZE]byte
	rdy             [1]byte
	ackbuff         [6]byte
	pn532ack        [6]byte // The ACK message from PN532
	firmwareVersion [6]byte
}

// NewI2C creates a new PN532 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func NewI2C(bus *machine.I2C) Device {
	return Device{
		bus:     bus,
		address: Address,
		debug:   false,
		pn532ack: [...]byte{
			0x00, 0x00, 0xFF,
			0x00, 0xFF, 0x00,
		},
		firmwareVersion: [...]byte{
			0x00, 0x00, 0xFF,
			0x06, 0xFA, 0xD5,
		},
	}
}

// enable/disable debugging
func (d *Device) Debug(b bool) {
	d.debug = b
}

func (d *Device) Configure() error {
	time.Sleep(10 * time.Millisecond)
	return d.wakeup()
}

func (d *Device) wakeup() error {
	return d.samconfig()
}

func (d *Device) samconfig() error {
	buffer := d.buffer[:4]
	buffer[0] = COMMAND_SAMCONFIGURATION
	buffer[1] = 0x01 // normal mode
	buffer[2] = 0x14 // timeout 50ms * 20 = 1 second
	buffer[3] = 0x01 // use IRQ PIN!

	if err := d.sendCommandCheckAck(buffer, 100*time.Millisecond); err != nil {
		return err
	}
	buffer = d.buffer[:9]
	if err := d.readdata(buffer); err != nil {
		return err
	}
	const offset = 6
	if buffer[offset] != 0x15 {
		return errors.New("unexpected response during sam configuration")
	}
	return nil
}

func (d *Device) sendCommandCheckAck(command []byte, timeout time.Duration) error {
	// write the command
	if err := d.writecommand(command); err != nil {
		return err
	}
	if !d.waitready(timeout) {
		return errors.New("waitready failed")
	}
	d.i2cTuning()
	if !d.isACK() {
		return errors.New("readack failed")
	}
	d.i2cTuning()
	if !d.waitready(timeout) {
		return errors.New("second waitready failed")
	}
	return nil
}

func (d *Device) writecommand(cmd []byte) error {
	packet := d.txBuffer[:8+len(cmd)]
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
	if err := d.bus.Tx(d.address, packet, nil); err != nil {
		return err
	}
	return nil
}

func (d *Device) waitready(timeout time.Duration) bool {
	const delay = 10 * time.Millisecond
	timer := 1 * time.Millisecond
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

func (d *Device) isACK() bool {
	err := d.readdata(d.ackbuff[:])
	if err != nil {
		return false
	}
	d.printBuffer("ACK", d.ackbuff[:])
	return bytes.Equal(d.ackbuff[:], d.pn532ack[:])
}

func (d *Device) printBuffer(name string, buffer []byte) {
	if d.debug {
		println(name, ":")
		print(hex.Dump(buffer))
		println()
	}
}

func (d *Device) readdata(buffer []byte) error {
	rxBuffer := d.rxBuffer[:len(buffer)+1]
	d.bus.Tx(d.address, nil, rxBuffer)
	copy(buffer, rxBuffer[1:])
	return nil
}

func (d *Device) isReady() bool {
	d.bus.Tx(d.address, nil, d.rdy[:])
	return d.rdy[0] == PN532_I2C_READY
}

func (d *Device) i2cTuning() {
	// I2C delay
	time.Sleep(1 * time.Millisecond)
}

func (d *Device) FirmwareVersion() (FirmwareVersion, error) {
	version := FirmwareVersion{}
	buffer := d.buffer[:1]
	buffer[0] = COMMAND_GETFIRMWAREVERSION
	err := d.sendCommandCheckAck(buffer, 100*time.Millisecond)
	if err != nil {
		return version, err
	}

	buffer = d.buffer[:13]
	err = d.readdata(buffer)
	if err != nil {
		return version, err
	}
	d.printBuffer("Firmware", buffer)

	if !bytes.Equal(buffer[0:len(d.firmwareVersion)], d.firmwareVersion[:]) {
		return version, errors.New("Invalid response received")
	}

	var response uint32 = 0
	for i := 7; i <= 10; i++ {
		response <<= 8
		response |= uint32(buffer[i])
	}
	version.IC = uint8((response >> 24) & 0xFF)
	version.Ver = uint8((response >> 16) & 0xFF)
	version.Rev = uint8((response >> 8) & 0xFF)
	version.Support = uint8((response) & 0xFF)

	return version, nil
}

func (d *Device) ReadPassiveTargetID(cardbaudrate uint8, timeout time.Duration) ([]byte, error) {
	buffer := d.buffer[:3]
	buffer[0] = COMMAND_INLISTPASSIVETARGET
	buffer[1] = 1 // limit this for one card at the moment
	buffer[2] = cardbaudrate

	if err := d.sendCommandCheckAck(buffer, timeout); err != nil {
		return []byte{}, errors.Join(errors.New("Failed sendCommandCheckAck"), err)
	}

	return d.ReadDetectedPassiveTargetID()
}

func (d *Device) ReadDetectedPassiveTargetID() ([]byte, error) {
	buffer := d.buffer[:20]
	if err := d.readdata(buffer); err != nil {
		return []byte{}, err
	}
	/* ISO14443A card response should be in the following format:

	   byte            Description
	   -------------   ------------------------------------------
	   b0..6           Frame header and preamble
	   b7              Tags Found
	   b8              Tag Number (only one used in this example)
	   b9..10          SENS_RES
	   b11             SEL_RES
	   b12             NFCID Length
	   b13..NFCIDLen   NFCID
	*/
	const b13 = 13
	if buffer[7] != 1 {
		return []byte{}, errors.New("invalid amount of cards detected")
	}
	var sense_res uint16 = uint16(buffer[9])

	sense_res <<= 8
	sense_res |= uint16(buffer[10])
	nfcIDLen := int(buffer[12])
	uid := buffer[b13 : b13+nfcIDLen]

	return uid, nil
}
