# PN532 Driver

This is still a work in progress driver. At the moment it only supports I2C. For an example check the [nfc](nfc/) example.

## Datasheet and user manual

- [PN532 User Manual](https://www.nxp.com/docs/en/user-guide/141520.pdf)
- [PN532/C1 data sheet](https://www.nxp.com/docs/en/nxp/data-sheets/PN532_C1.pdf)

For implementing the communication the [PN532 User Manual](https://www.nxp.com/docs/en/user-guide/141520.pdf) is of much more use than the Datasheet.


## Reference Driver Code

- [Arduino Adafruit-PN532](https://github.com/adafruit/Adafruit-PN532/tree/master)

## Reference Hardware

I have tested this driver on a cheap NCF Module V3 board.
![Elechouse NFC Module V3](doc/images/nfc-module-v3.png)

This board does not provide the `RSTPD_N`, this is why I did not have used the reset logic.

## License

Since this is basically a rewrite of the Adafruit Code in Go the license is BSD license as of the original code.