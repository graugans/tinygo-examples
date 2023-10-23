package main

import (
	"fmt"
	"image/color"
	"machine"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

var leds [3]color.RGBA

func main() {
	for i := 0; i <= 3; i++ {
		fmt.Println("Bootdelay: ", i, " seconds")
		time.Sleep(1 * time.Second)
	}
	// NEO Pixel
	var neo machine.Pin = machine.GPIO29
	neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ws := ws2812.New(neo)

	eveningOn := time.Duration(5) * time.Hour
	morningOn := time.Duration(3) * time.Hour
	for {
		cutOffTime := time.Now().Add(eveningOn)
		for time.Now().Before(cutOffTime) {
			jackOlantern(&ws)
		}
		switchOff(&ws)
		time.Sleep(time.Duration(6) * time.Hour)
		cutOffTime = time.Now().Add(morningOn)
		for time.Now().Before(cutOffTime) {
			jackOlantern(&ws)
		}
		switchOff(&ws)
		time.Sleep(time.Duration(10) * time.Hour)
	}
}

func jackOlantern(ws *ws2812.Device) {
	randOrange := color.RGBA{
		R: uint8(rand.Intn(255)),
		G: uint8(rand.Intn(255 * 60 / 100)),
		B: 0,
	}
	randRed := color.RGBA{
		R: uint8(rand.Intn(255)),
		G: 0,
		B: 0,
	}

	leds[0] = randOrange
	leds[1] = randRed
	leds[2] = randRed

	ws.WriteColors(leds[:])
	time.Sleep(100 * time.Millisecond)
}

func switchOff(ws *ws2812.Device) {
	for i := range leds {
		leds[i] = color.RGBA{
			R: 0,
			G: 0,
			B: 0,
		}
	}
	ws.WriteColors(leds[:])
}
