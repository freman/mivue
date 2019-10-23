# MiVue Projects

 - [sdbin](#sdbin)
   * [Removing the header](#removing-the-header)
   * [Build](#build)
   * [Execution](#execution)

## sdbin

[sdbin](cmd/sdbin) is a tool for re-applying the header and md5sum to MiVue firmware files.

This tool is provided without warranty and for your use at your own risk. I'm not responsible if you brick your dashcam.

### Removing the header

1. Obtain a flashable firmware file you want to start with. The filename will be SD_CarDV.bin.
2. Strip the first 32 bytes from that file. You can use `dd` or a GUI hex editor such as [Frhed](http://frhed.sourceforge.net/en/).
3. Make any changes you want to the firmware file, for example with the hex editor linked above. It's all there in the open, unencrypted and uncompressed.
4. Use this utility to generate a 32-byte header for the firmware, so that the device accepts it for flashing.
5. Flash the firmware the usual way.

### Build

To build from source you need **[Git](https://git-scm.com/downloads)** and **[Go](https://golang.org/doc/install)** (1.13 or newer).

1. Run `go get github.com/freman/mivue/cmd/sdbin`

### Execution

```
sdbin source [dest]

source:         source is input file name
dest:           dest is output file name default is "SD_CarDV.bin"
```

Example:

```
sdbin firmware.bin
Input file size 16515072
f9a5284932578282d10a9caa4b9e3831
```