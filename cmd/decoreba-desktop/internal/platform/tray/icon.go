package tray

import (
	"image"
	"image/color"
	"image/draw"
)

func generateIcon() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 24, 24))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.Point{}, draw.Src)

	fg := color.RGBA{255, 255, 255, 255}
	glyph := []struct{ x, y int }{
		{8, 6}, {9, 7}, {10, 8}, {11, 9}, {12, 10}, {13, 11}, {14, 11}, {15, 12},
		{14, 13}, {13, 13}, {12, 14}, {11, 15}, {10, 16}, {9, 17}, {8, 18},
	}
	for _, p := range glyph {
		img.Set(p.x, p.y, fg)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	buf := make([]byte, 8+w*h*4)
	putLE32(buf[0:4], uint32(w))
	putLE32(buf[4:8], uint32(h))
	off := 8
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixel := uint32(a>>8)<<24 | uint32(r>>8)<<16 | uint32(g>>8)<<8 | uint32(b>>8)
			putLE32(buf[off:off+4], pixel)
			off += 4
		}
	}
	return buf
}

func putLE32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}
