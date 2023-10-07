package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/at24cx"
	"tinygo.org/x/drivers/ds3231"
)

// Returns a new RTC object on success, an error is returned in case of an error
func NewRTC(i2c drivers.I2C) (*ds3231.Device, error) {
	rtc := ds3231.New(i2c)
	rtc.Configure()
	valid := rtc.IsTimeValid()
	if !valid {
		date := time.Date(2023, 10, 5, 19, 34, 12, 0, time.UTC)
		rtc.SetTime(date)
	}

	running := rtc.IsRunning()
	if !running {
		err := rtc.SetRunning(true)
		if err != nil {
			return nil, err
		}
	}
	return &rtc, nil
}

func main() {
	var rawDS3231Temperature int32
	var rawCPUTemperature int32

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
	})

	rtc, err := NewRTC(machine.I2C0)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	eeprom := at24cx.New(machine.I2C0)
	eeprom.Configure(at24cx.Config{})

	dt, err := rtc.ReadTime()
	if err != nil {
		fmt.Println("Error reading date:", err)
	}
	data := []byte(
		fmt.Sprintf("Start Date: %d/%s/%02d %02d:%02d:%02d", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second()),
	)
	written, err := eeprom.Write(data)
	if err != nil {
		fmt.Printf("Error while writing to EEPROM: %s\n", err.Error())
	}
	fmt.Printf("Num bytes %d, written to EEPROM\n", written)

	for {
		dt, err := rtc.ReadTime()
		if err != nil {
			fmt.Println("Error reading date:", err)
		} else {
			fmt.Printf("Date: %d/%s/%02d %02d:%02d:%02d \r\n", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
		}
		if rawDS3231Temperature, err = rtc.ReadTemperature(); err != nil {
			fmt.Println("Error while reading the Temperature")
			continue
		}

		if rawCPUTemperature = machine.ReadTemperature(); err != nil {
			fmt.Println("Error while reading the Temperature")
			continue
		}

		fmt.Printf("DS3231 Temperature: %.2f °C \r\n", float32(rawDS3231Temperature)/1000)
		fmt.Printf("CPU    Temperature: %.2f °C \r\n", float32(rawCPUTemperature)/1000)

		eeprom.Seek(0, 0) // seek to the beginning
		tmp := make([]byte, len(data))
		size, err := eeprom.Read(tmp)
		if err != nil {
			fmt.Printf("Error while reading from EEPROM: %s\n", err.Error())
		}
		fmt.Printf("EEPROM: %s (%d)\n", string(tmp), size)
		time.Sleep(time.Second * 1)
	}
}
