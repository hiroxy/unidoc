/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package draw

import (
	"math"

	"github.com/unidoc/unidoc/pdf/model"
)

// CubicBezierCurve describes a cubic bezier curve which is defined by:
// R(t) = P0*(1-t)^3 + P1*3*t*(1-t)^2 + P2*3*t^2*(1-t) + P3*t^3
// where P0 is the current point, P1, P2 control points and P3 the final point.
type CubicBezierCurve struct {
	P0 Point // Starting point.
	P1 Point // Control point 1.
	P2 Point // Control point 2.
	P3 Point // Final point.
}

// NewCubicBezierCurve returns a CubicBezierCurve with points (xi, yi) i=0..3
func NewCubicBezierCurve(x0, y0, x1, y1, x2, y2, x3, y3 float64) CubicBezierCurve {
	curve := CubicBezierCurve{}
	curve.P0 = NewPoint(x0, y0)
	curve.P1 = NewPoint(x1, y1)
	curve.P2 = NewPoint(x2, y2)
	curve.P3 = NewPoint(x3, y3)
	return curve
}

// AddOffsetXY returns a copy of `curve` with all points translated` by `dX`,`dY`.
func (curve CubicBezierCurve) AddOffsetXY(dX, dY float64) CubicBezierCurve {
	curve.P0.X += dX
	curve.P1.X += dX
	curve.P2.X += dX
	curve.P3.X += dX

	curve.P0.Y += dY
	curve.P1.Y += dY
	curve.P2.Y += dY
	curve.P3.Y += dY

	return curve
}

// GetBounds returns a PdfRectangle of the bounding box of curve.
func (curve CubicBezierCurve) GetBounds() model.PdfRectangle {
	minX := curve.P0.X
	maxX := curve.P0.X
	minY := curve.P0.Y
	maxY := curve.P0.Y

	// 1000 points.
	for t := 0.0; t <= 1.0; t += 0.001 {
		Rx := curve.P0.X*math.Pow(1-t, 3) +
			curve.P1.X*3*t*math.Pow(1-t, 2) +
			curve.P2.X*3*math.Pow(t, 2)*(1-t) +
			curve.P3.X*math.Pow(t, 3)
		Ry := curve.P0.Y*math.Pow(1-t, 3) +
			curve.P1.Y*3*t*math.Pow(1-t, 2) +
			curve.P2.Y*3*math.Pow(t, 2)*(1-t) +
			curve.P3.Y*math.Pow(t, 3)

		if Rx < minX {
			minX = Rx
		}
		if Rx > maxX {
			maxX = Rx
		}
		if Ry < minY {
			minY = Ry
		}
		if Ry > maxY {
			maxY = Ry
		}
	}

	bounds := model.PdfRectangle{}
	bounds.Llx = minX
	bounds.Lly = minY
	bounds.Urx = maxX
	bounds.Ury = maxY
	return bounds
}

// CubicBezierPath represents a pdf path composed of cubic Bezier curves
type CubicBezierPath struct {
	Curves []CubicBezierCurve
}

// NewCubicBezierPath returns a CubicBezierPath with no curves
func NewCubicBezierPath() CubicBezierPath {
	bpath := CubicBezierPath{}
	bpath.Curves = []CubicBezierCurve{}
	return bpath
}

// AppendCurve returns a copy of `bpath` with `curve` appended
func (bpath CubicBezierPath) AppendCurve(curve CubicBezierCurve) CubicBezierPath {
	bpath.Curves = append(bpath.Curves, curve)
	return bpath
}

// Copy returns a copy of bpath
func (bpath CubicBezierPath) Copy() CubicBezierPath {
	bpathcopy := CubicBezierPath{}
	bpathcopy.Curves = []CubicBezierCurve{}
	for _, c := range bpath.Curves {
		bpathcopy.Curves = append(bpathcopy.Curves, c)
	}
	return bpathcopy
}

// AddOffsetXY returns a copy of `bpath` with all points translated by `dX`,`dY`.
// XXX Why is this not called Translate?
func (bpath CubicBezierPath) Offset(dX, dY float64) CubicBezierPath {
	for i, c := range bpath.Curves {
		bpath.Curves[i] = c.AddOffsetXY(dX, dY)
	}
	return bpath
}

// Length returns the number of curves in `bpath`.
func (bpath CubicBezierPath) Length() int {
	return len(bpath.Curves)
}

// GetBoundingBox returns the bounding box of bpath as a Rectangle.
func (bpath CubicBezierPath) GetBoundingBox() Rectangle {
	bbox := Rectangle{}
	if len(bpath.Curves) == 0 {
		return bbox
	}

	curveBounds := bpath.Curves[0].GetBounds()
	minX := curveBounds.Llx
	maxX := curveBounds.Urx
	minY := curveBounds.Lly
	maxY := curveBounds.Ury

	for _, c := range bpath.Curves {
		curveBounds := c.GetBounds()
		if curveBounds.Llx < minX {
			minX = curveBounds.Llx
		}
		if curveBounds.Urx > maxX {
			maxX = curveBounds.Urx
		}
		if curveBounds.Lly < minY {
			minY = curveBounds.Lly
		}
		if curveBounds.Ury > maxY {
			maxY = curveBounds.Ury
		}
	}

	bbox.X = minX
	bbox.Y = minY
	bbox.Width = maxX - minX
	bbox.Height = maxY - minY
	return bbox
}
