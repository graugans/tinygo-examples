# Introduction

Right before Halloween people are carving their Pumpkins. This projects intention is to give the pumpkin some illumination. Two WS2812B LEDs for the eyes and one for the rest of the head.

![Schematic](/doc/images/halloween.png)

Watch the Jack O Lantern in action

[![Example Video](/doc/images/halloween-jack.png)](https://www.youtube.com/embed/Sz_yLomxdYA?si=bnogcEC23ATQtMl7 "Tinygo driven Halloween Jack O Lantern")

## Microcontroller

For this project I have used the [Waveshare RP2040-Zero](https://www.waveshare.com/rp2040-zero.htm)

![Waveshare RP2040-Zero](/doc/images/RP2040-Zero-details-7.jpg)

This board comes in the right size.


## Flashing

```sh
tinygo flash -size short -target=waveshare-rp2040-zero halloween/main.go
```