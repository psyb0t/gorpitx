---
title: "spectrumpaint"
description: "The gorpitx spectrumpaint module: draw a 320-pixel-wide grayscale picture on a receiver's waterfall from a raw luminance file."
nav_order: 6
parent: "Modules"
---

# spectrumpaint: a picture on the waterfall

Paints a picture onto the waterfall display of an SDR receiver. The program reads a raw file one 320-byte row at a time and turns each byte into the strength of one point across the band. Code: [`spectrumpaint.go`](https://github.com/psyb0t/gorpitx/blob/main/spectrumpaint.go).

## Fields

| JSON | Go field | Type | Required | Rules |
| --- | --- | --- | --- | --- |
| `pictureFile` | `PictureFile` | `string` | yes | Path to an existing raw file, 320 bytes per row |
| `frequency` | `Frequency` | `float64` | yes | Hz, 5000 to 1500000000 |
| `excursion` | `Excursion` | `*float64` | no | Width of the picture in Hz, greater than 0; the program defaults to 100000 |

## Making the picture file

The width of 320 is fixed in the rpitx program. ImageMagick can write the luminance plane on its own:

```bash
convert input.jpg -resize 320x100! -flip \
  -quantize YUV -dither FloydSteinberg -colors 4 \
  -depth 8 -interlace partition picture.yuv
```

`-interlace partition` writes three files, `picture.Y`, `picture.U` and `picture.V`. Use `picture.Y`. `-depth 8` matters: a stock ImageMagick build writes 16 bits per sample without it, which makes each row 640 bytes and garbles the picture. Check the result with `ls -l`; a 320x100 picture is 32000 bytes. `-flip` turns the image upside down so it reads the right way on a waterfall that scrolls down. Leave out the `100!` height to keep the aspect ratio.

gorpitx checks only that the file exists. It does not check the width or the size.

## Example

```json
{
  "pictureFile": "/home/pi/picture.Y",
  "frequency": 434000000,
  "excursion": 100000
}
```

Runs:

```bash
stdbuf -oL "$GORPITX_PATH/spectrumpaint" /home/pi/picture.Y 434000000 100000
```
