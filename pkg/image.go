package pkg

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png"

	"golang.org/x/image/draw"
)

func GenerateThumbnail(data []byte, maxWidth int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	if width > maxWidth {
		height = (height * maxWidth) / width
		width = maxWidth
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)

	var buf bytes.Buffer

	err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 40})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
