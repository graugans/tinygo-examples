# MIFARE NFC Card dump

This uses a Elechouse PN532 NFC Module v3 attached via I2C to a Raspberry PI Pico. This example blocks until a MIFARE Classik 1k/4k card is in range of the reader and dumps its content. The default encryption keys are expected.

The driver and the example is based on the [PN532 Adafruit C++ driver](https://github.com/adafruit/Adafruit-PN532).

## Flashing

```sh
tinygo flash -size short  -monitor -target=pico ./nfc
```


## License

This project is licensed under the BSD 3-clause license.