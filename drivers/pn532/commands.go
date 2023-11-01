package pn532

// Mifare Commands
const (
	MIFARE_CMD_AUTH_A           = 0x60 ///< Auth A
	MIFARE_CMD_AUTH_B           = 0x61 ///< Auth B
	MIFARE_CMD_READ             = 0x30 ///< Read
	MIFARE_CMD_WRITE            = 0xA0 ///< Write
	MIFARE_CMD_TRANSFER         = 0xB0 ///< Transfer
	MIFARE_CMD_DECREMENT        = 0xC0 ///< Decrement
	MIFARE_CMD_INCREMENT        = 0xC1 ///< Increment
	MIFARE_CMD_STORE            = 0xC2 ///< Store
	MIFARE_ULTRALIGHT_CMD_WRITE = 0xA2 ///< Write (MiFare Ultralight)
)
