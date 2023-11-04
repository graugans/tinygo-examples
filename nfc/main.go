package main

import (
	"encoding/hex"
	"machine"
	"strconv"
	"time"

	"github.com/graugans/tinygo-examples/drivers/pn532"
)

func main() {
	const delay = 3
	for i := 0; i <= delay; i++ {
		time.Sleep(time.Second) // allow to attach the monitor
		println("Sleeping...")
	}
	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       0,
		SCL:       1,
	})
	if err != nil {
		println("Error I2C set-up", err.Error())
	}

	nfc := pn532.NewI2C(machine.I2C0)
	if err := nfc.Configure(); err != nil {
		println("Error Configure: ", err.Error())
		return
	}
	// Enable/Disbale the debug output
	nfc.Debug(false)
	// Retrieve the Firmware version from the PN
	version, err := nfc.FirmwareVersion()
	if err != nil {
		println("Error:", err.Error())
		return
	}
	println("-------------------------------------------------------------")
	println(version.String())
	println("Sleeping for 3 seconds .....")
	time.Sleep(3 * time.Second)
	println("-------------------------------------------------------------")
	for {
		time.Sleep(time.Second)
		uid, err := nfc.ReadPassiveTargetID(pn532.MIFARE_ISO14443A, 0)
		if err != nil {

			println(err)
			continue
		}
		println("-------------------------------------------------------------")
		println("Found an ISO14443A card")
		println("  UID Length: " + strconv.Itoa(len(uid)))
		println("  UID Value:", hex.EncodeToString(uid))
		println("-------------------------------------------------------------")
		if len(uid) == 4 { // We assume that this is a Mifare Classic card ...
			printMifareClasicUID(uid)
			mifare := pn532.NewMifareClasic(&nfc)
			// Now we try to go through all 16 sectors (each having 4 blocks)
			// authenticating each sector, and then dumping the blocks
			authenticated := false
			println("------------------------ Dumping the card content -------------------------")
			for currentblock := 0; currentblock < 64; currentblock++ {
				if mifare.IsFirstBlock(uint32(currentblock)) {
					authenticated = false
				}
				if !authenticated {
					// Starting of a new sector ... try to to authenticate
					println("------------------------ Sector " + strconv.Itoa(currentblock/4) + " -------------------------")
					var err error
					if currentblock == 0 {
						err = mifare.AuthenticateBlock(uid, uint32(currentblock), pn532.MifareClassicKeyA)
					} else {
						err = mifare.AuthenticateBlock(uid, uint32(currentblock), pn532.MifareClassicKeyB)
					}
					if err != nil {
						println("Authentication error: ", err)
						continue
					}
					authenticated = true
				}
				if !authenticated {
					// If we're still not authenticated just skip the block
					println("Unable to authenticate block: ", currentblock)
					continue
				}
				// Authenticated ... we should be able to read the block now
				// Dump the data into the 'data' slice
				data, err := mifare.ReadDataBlock(uint8(currentblock))
				if err != nil {
					// Oops ... something happened
					println("Unable to read this block:", currentblock)
				}
				print(hex.Dump(data))
			}

			if err := mifare.AuthenticateBlock(uid, 4, pn532.MifareClassicKeyA); err != nil {
				println(err)
				continue
			}
		}
		println("-------------------------------------------------------------")
		println("Sleeping for 3 seconds .....")
		time.Sleep(3 * time.Second)
	}
}

func printMifareClasicUID(uid []byte) {
	// We probably have a Mifare Classic card ...
	var cardid uint32 = 0x00
	for _, elem := range uid {
		cardid <<= 8
		cardid |= uint32(elem)
	}
	println("Mifare Clasic card ID:", cardid, "(0x"+hex.EncodeToString(uid)+")")
}
