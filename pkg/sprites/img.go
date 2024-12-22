package sprites

import (
	"crypto/sha1"
	"encoding/hex"
	"image"
	"image/png"
)

type WithSubImage interface {
	image.Image
	SubImage(r image.Rectangle) image.Image
}

func splitImage(img WithSubImage, w, h int) []image.Image {
	var imgs []image.Image
	bounds := img.Bounds().Size()
	for iy := 0; iy < bounds.Y; iy += h {
		for ix := 0; ix < bounds.X; ix += w {
			r := image.Rect(ix, iy, ix+w, iy+h)
			sub := img.SubImage(r)
			imgs = append(imgs, sub)
		}
	}
	return imgs
}

func deduplicateImages(imgs []image.Image) []image.Image {
	unique := make(map[string]struct{})
	var deduped []image.Image
	for _, img := range imgs {
		h := sha1.New()
		png.Encode(h, img)
		key := hex.EncodeToString(h.Sum(nil))
		if _, ok := unique[key]; !ok {
			deduped = append(deduped, img)
			println(key, len(unique))
			unique[key] = struct{}{}
		}
	}
	return deduped
}
