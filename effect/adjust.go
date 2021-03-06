package effect

import (
	"image"
	"image/draw"

	"github.com/kennythebard/kromatique/core"
	"github.com/kennythebard/kromatique/strategy"
	"github.com/kennythebard/kromatique/utils"
)

// Adjust returns a function that applies the given MappingRules on an image
func Adjust(rules ...strategy.MappingRule) func(image.Image) image.Image {
	return func(img image.Image) image.Image {
		ret := utils.CreateRGBA(img.Bounds())

		core.Parallelize(img.Bounds().Dy(), func(y int) {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				currentColor := img.At(x, y)
				for _, rule := range rules {
					currentColor = rule(currentColor)
				}

				ret.(draw.Image).Set(x, y, currentColor)
			}
		})
		return ret
	}
}
