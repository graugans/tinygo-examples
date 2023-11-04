package pn532

import (
	"encoding/hex"
	"errors"
	"time"
)

type (
	MifareClassicKeyType uint8
	MifareClassicKey     []byte
	MifareClassic        struct {
		dev  *Device
		keys [2]MifareClassicKey
	}
)

const (
	MifareClassicKeyA MifareClassicKeyType = 0
	MifareClassicKeyB MifareClassicKeyType = 1
)

const MifareClassicBlockSize = 16

func NewMifareClasic(device *Device) MifareClassic {
	return MifareClassic{
		dev: device,
		keys: [2]MifareClassicKey{
			{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
	}
}

func (m *MifareClassic) selectKeyCommand(number MifareClassicKeyType) byte {
	if number == MifareClassicKeyA {
		return MIFARE_CMD_AUTH_A
	}
	return MIFARE_CMD_AUTH_B
}

func (m *MifareClassic) AuthenticateBlock(uid []byte,
	blockNumber uint32,
	keyNumber MifareClassicKeyType,
) error {
	const commandLen = 10
	buffer := make([]byte, commandLen+len(uid))
	buffer[0] = COMMAND_INDATAEXCHANGE /* Data Exchange Header */
	buffer[1] = 1                      /* Max card numbers */
	buffer[2] = m.selectKeyCommand(keyNumber)
	buffer[3] = byte(blockNumber)
	copy(buffer[4:], m.keys[keyNumber])
	copy(buffer[10:], uid)
	m.dev.printBuffer("Auth Buffer: ", buffer)
	if err := m.dev.sendCommandCheckAck(buffer, 100*time.Millisecond); err != nil {
		return err
	}
	buffer = make([]byte, 12)
	if err := m.dev.readdata(buffer); err != nil {
		return err
	}
	// check if the response is valid and we are authenticated???
	// for an auth success it should be bytes 5-7: 0xD5 0x41 0x00
	// Mifare auth error is technically byte 7: 0x14 but anything other and 0x00
	// is not good
	m.dev.printBuffer("Auth response", buffer)
	if buffer[7] != 0x00 {
		message := "Mifare auth error, expected 0x00 got: 0x" + hex.EncodeToString(buffer[7:7])
		return errors.New(message)
	}
	return nil
}

func (m *MifareClassic) ReadDataBlock(blockNumber uint8) ([]byte, error) {
	buffer := m.dev.buffer[:4]
	buffer[0] = COMMAND_INDATAEXCHANGE /* Data Exchange Header */
	buffer[1] = 1                      /* Card number */
	buffer[2] = MIFARE_CMD_READ        /* Card number */
	buffer[3] = blockNumber
	if err := m.dev.sendCommandCheckAck(buffer, 100*time.Millisecond); err != nil {
		return []byte{}, err
	}
	buffer = make([]byte, 26)
	if err := m.dev.readdata(buffer); err != nil {
		return []byte{}, err
	}
	m.dev.printBuffer("read response", buffer)
	if buffer[7] != 0x00 {
		message := "Mifare read response, expected 0x00 got: 0x" + hex.EncodeToString(buffer[8:8])
		return []byte{}, errors.New(message)
	}
	data := make([]byte, 16)
	copy(data, buffer[8:8+len(data)])
	m.dev.printBuffer("data buffer", data)
	return data, nil
}

func (m *MifareClassic) WriteDataBlock(blockNumber uint8, data []byte) error {
	if len(data) > MifareClassicBlockSize {
		return errors.New("The given data exceeds the block size")
	}
	buffer := m.dev.buffer[:20]
	buffer[0] = COMMAND_INDATAEXCHANGE
	buffer[1] = 1 // card number
	buffer[2] = MIFARE_CMD_WRITE
	buffer[3] = blockNumber // Block Number (0..63 for 1K, 0..255 for 4K)
	copy(buffer[4:], data)
	m.dev.printBuffer("data buffer", data)

	if err := m.dev.sendCommandCheckAck(buffer, 100*time.Millisecond); err != nil {
		return err
	}
	// Give the PN532 some time to perfrom the write
	time.Sleep(10 * time.Millisecond)
	buffer = make([]byte, 26)
	if err := m.dev.readdata(buffer); err != nil {
		return err
	}
	m.dev.printBuffer("write response", buffer)
	return nil
}

func (m *MifareClassic) SetKeyA(key MifareClassicKey) {
	m.keys[MifareClassicKeyA] = key
}

func (m *MifareClassic) SetKeyB(key MifareClassicKey) {
	m.keys[MifareClassicKeyB] = key
}

func (m *MifareClassic) IsFirstBlock(block uint32) bool {
	if block < 128 {
		return (block%4 == 0)
	}
	return (block%16 == 0)
}

func (m *MifareClassic) IsTrailerBlock(block uint32) bool {
	if block < 128 {
		return ((block+1)%4 == 0)
	}
	return ((block+1)%16 == 0)
}
