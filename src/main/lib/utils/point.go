package utils

import (
	"image"
	"math"
)

type Point2D struct {
	X, Y float64
}

func Pt2D(X, Y float64) Point2D {
	return Point2D{X: X, Y: Y}
}

func NewPoint2D(X, Y float64) *Point2D {
	p := new(Point2D)
	p.X = X
	p.Y = Y

	return p
}

func (p Point2D) Dist(q Point2D) float64 {
	return math.Sqrt((p.X-q.X)*(p.X-q.X) + (p.Y-q.Y)*(p.Y-q.Y))
}

func (p Point2D) In(r image.Rectangle) bool {
	return float64(r.Min.X) <= p.X && p.X < float64(r.Max.X) &&
		float64(r.Min.Y) <= p.Y && p.Y < float64(r.Max.Y)
}
