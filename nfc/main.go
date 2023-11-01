package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/graugans/tinygo-examples/drivers/pn532"
)

func main() {
	const delay = 3
	for i := 0; i <= delay; i++ {
		time.Sleep(time.Second) // allow to attach the monitor
		fmt.Printf("Sleeping %d/%d \n", i, delay)
	}
	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       0,
		SCL:       1,
	})
	if err != nil {
		fmt.Printf("Error I2C set-up %s \n", err.Error())
	}

	nfc := pn532.NewI2C(machine.I2C0)
	if err := nfc.Configure(); err != nil {
		fmt.Printf("Error Configure: %s \n", err.Error())
		return
	}

	// Retrieve the Firmware version from the PN
	version, err := nfc.FirmwareVersion()
	if err != nil {
		fmt.Printf("Error: %s \n", err.Error())
		return
	}
	fmt.Println("-------------------------------------------------------------")
	fmt.Println(version.String())
	fmt.Println("Sleeping for 3 seconds .....")
	time.Sleep(3 * time.Second)
	fmt.Println("-------------------------------------------------------------")
	for {
		time.Sleep(time.Second)
		uid, err := nfc.ReadPassiveTargetID(pn532.MIFARE_ISO14443A, 0)
		if err != nil {

			fmt.Println(err)
			continue
		}
		fmt.Println("-------------------------------------------------------------")
		fmt.Println("Found an ISO14443A card")
		fmt.Printf("  UID Length: %d bytes\n", len(uid))
		printBuffer("  UID Value: ", uid)
		fmt.Println("-------------------------------------------------------------")
		if len(uid) == 4 { // We assume that this is a Mifare Classic card ...
			printMifareClasicUID(uid)
			mifare := pn532.NewMifareClasic(&nfc)
			// Now we try to go through all 16 sectors (each having 4 blocks)
			// authenticating each sector, and then dumping the blocks
			authenticated := false
			fmt.Println("------------------------ Dumping the card content -------------------------")
			for currentblock := 0; currentblock < 64; currentblock++ {
				if mifare.IsFirstBlock(uint32(currentblock)) {
					authenticated = false
				}
				if !authenticated {
					// Starting of a new sector ... try to to authenticate
					fmt.Printf("------------------------ Sector %02d -------------------------\n", uint(currentblock/4))
					var err error
					if currentblock == 0 {
						err = mifare.AuthenticateBlock(uid, uint32(currentblock), pn532.MifareClassicKeyA)
					} else {
						err = mifare.AuthenticateBlock(uid, uint32(currentblock), pn532.MifareClassicKeyB)
					}
					if err != nil {
						fmt.Println("Authentication error: ", err)
						continue
					}
					authenticated = true
				}
				if !authenticated {
					// If we're still not authenticated just skip the block
					fmt.Printf("Block %d unable to authenticate\n", currentblock)
					continue
				}
				// Authenticated ... we should be able to read the block now
				// Dump the data into the 'data' slice
				data, err := mifare.ReadDataBlock(uint8(currentblock))
				if err != nil {
					// Oops ... something happened
					fmt.Printf("Block %d unable to read this block\n", currentblock)
				}
				printBuffer(fmt.Sprintf("Block %02d", currentblock), data)

			}

			if err := mifare.AuthenticateBlock(uid, 4, pn532.MifareClassicKeyA); err != nil {
				fmt.Println(err)
				continue
			}
		}
		fmt.Println("-------------------------------------------------------------")
		fmt.Println("Sleeping for 3 seconds .....")
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
	fmt.Printf("Mifare Clasic card ID: %d (0x%X)\n", cardid, cardid)
}

func printBuffer(name string, buffer []byte) {
	fmt.Printf("%s buffer: [ ", name)
	for _, b := range buffer {
		fmt.Printf("0x%02X ", b)
	}
	fmt.Printf("]\n")
}
