# Introduction

Right before Halloween people are carving their Pumpkins. This projects intention is to give the pumpkin some illumination. Two WS2812B LEDs for the eyes and one for the rest of the head.

![Schematic](doc/images/halloween.png)

## Microcontroller

For this project I have used the [Waveshare RP2040-Zero](https://www.waveshare.com/rp2040-zero.htm)

![Waveshare RP2040-Zero](doc/images/RP2040-Zero-details-7.jpg)

This board comes in the right size.


## Flashing

```sh
tinygo flash -target=waveshare-rp2040-zero halloween/main.go
```