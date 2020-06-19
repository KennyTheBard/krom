package morphing

import (
	"image"
	"image/draw"

	"../utils"
)

func Morph(src, dst image.Image, srcPoints, dstPoints []Vertex, numSteps int) []image.Image {
	if len(srcPoints) != len(dstPoints) {
		panic("Unequal number of points for images")
	}

	src2dst, dst2src := make(map[Vertex]Vertex), make(map[Vertex]Vertex)
	for idx := range srcPoints {
		src2dst[srcPoints[idx]] = dstPoints[idx]
		dst2src[dstPoints[idx]] = srcPoints[idx]
	}

	srcMesh, dstMesh := NewMeshForImage(src), NewMeshForImage(dst)
	BowyerWatson(srcMesh, srcPoints)
	BowyerWatson(dstMesh, dstPoints)

	blend := func(fg, bg image.Image, alpha float64) image.Image {
		for y := fg.Bounds().Min.Y; y < fg.Bounds().Max.Y; y++ {
			for x := fg.Bounds().Min.X; x < fg.Bounds().Max.X; x++ {
				bg.(draw.Image).Set(x, y, utils.PixelLERP(fg.At(x, y), bg.At(x, y), alpha))
			}
		}

		return bg
	}

	numSteps += 1
	images := make([]image.Image, 0)
	stepSize := 1.0 / float64(numSteps)
	for t, i := 0.0, 0; t <= 1.0; t, i = t+stepSize, i+1 {
		fg := NewMeshDeformation(srcMesh, src2dst, t).Deform(src)
		bg := NewMeshDeformation(dstMesh, dst2src, 1-t).Deform(dst)

		images = append(images, blend(fg, bg, t))
	}

	return images
}