---
title: "pisstv"
description: "The gorpitx pisstv module: send a 320-pixel-wide raw RGB picture as slow scan television in the Martin 1 mode."
nav_order: 8
parent: "Modules"
---

# pisstv: SSTV pictures

Sends a picture as slow scan television in the Martin 1 mode, which SSTV decoders such as QSSTV, MMSSTV or Robot36 receive. The Go constant is `ModuleNamePISSSTV`, with three S. Code: [`pisstv.go`](https://github.com/psyb0t/gorpitx/blob/main/pisstv.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `pictureFile` | `PictureFile` | `string` | yes | Path to an existing raw RGB file, 320 pixels wide |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |

## The picture file

The file is raw RGB with no header: three bytes per pixel (red, green, blue), 320 pixels per line, so 960 bytes per line. The program reads lines until the file ends. Martin 1 has 256 lines, which makes a 245760-byte file; other heights still transmit, but decoders expect 256.

```bash
convert input.jpg -resize 320x256! -depth 8 picture.rgb
```

Keep `-depth 8`. Without it a stock ImageMagick build writes 16 bits per sample and the file doubles to 491520 bytes. gorpitx checks only that the file exists. The repository's test fixture [`.fixtures/martin1.rgb`](https://github.com/psyb0t/gorpitx/blob/main/.fixtures/martin1.rgb) has this format.

A Martin 1 picture of 256 lines takes about two minutes.

## Example

```json
{
  "pictureFile": "/home/pi/picture.rgb",
  "frequency": 144500000
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/pisstv" /home/pi/picture.rgb 144500000
```

## What the program sends

From rpitx's `pisstv.cpp`: a VIS header that identifies Martin 1, then for each line a 1200 Hz sync pulse and the green, blue and red scans, each pixel as a tone from 1500 Hz (black) to just under 2300 Hz (white).
