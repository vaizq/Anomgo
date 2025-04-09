package captcha

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

type Solution struct {
	Radius int
	X      int
	Y      int
}

type Captcha struct {
	Image    *image.RGBA
	Solution Solution
}

func New() Captcha {
	radius := 20
	tickness := 2
	numCircles := 6

	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{180, 120}})
	pos := createCaptcha(img, radius, tickness, numCircles)

	solution := Solution{
		Radius: radius,
		X:      pos.X,
		Y:      pos.Y,
	}

	return Captcha{
		Image:    img,
		Solution: solution,
	}
}

func Distance(a, b image.Point) float64 {
	return math.Sqrt(math.Pow(float64(a.X-b.X), 2) + math.Pow(float64(a.Y-b.Y), 2))
}

func drawCircle(img *image.RGBA, center image.Point, radius int, color color.RGBA, tickness int) {
	r := float64(radius)
	da := 1 / r / 10
	for a := 0.0; a < 6.3; a += da {
		cos := math.Cos(a)
		sin := math.Sin(a)
		rtmp := r
		for i := 0; i < tickness; i++ {
			x := int(rtmp*cos) + center.X
			y := int(rtmp*sin) + center.Y
			img.Set(x, y, color)
			rtmp--
		}
	}
}

func drawArch(img *image.RGBA, center image.Point, radius int, color color.RGBA, tickness int, startAngle float64, endAngle float64) {
	r := float64(radius)
	da := 1 / r / 10
	for a := startAngle; a < endAngle; a += da {
		cos := math.Cos(a)
		sin := math.Sin(a)
		rtmp := r
		for i := 0; i < tickness; i++ {
			x := int(rtmp*cos) + center.X
			y := int(rtmp*sin) + center.Y
			img.Set(x, y, color)
			rtmp--
		}
	}
}

func createCaptcha(img *image.RGBA, radius int, tickness int, numCircles int) image.Point {
	black := color.RGBA{0, 0, 0, 255}

	randPoint := func() image.Point {
		return image.Point{rand.Intn(img.Bounds().Max.X-2*radius) + radius, rand.Intn(img.Bounds().Max.Y-2*radius) + radius}
	}

	centers := []image.Point{}

	for i := 0; i < numCircles; i++ {
		centers = append(centers, randPoint())
	}

	for _, center := range centers {
		drawCircle(img, center, radius, black, tickness)
	}

	p := image.Point{}
	found := false
	for i := 0; i < 10 && !found; i++ {
		p = randPoint()
		found = true
		for _, center := range centers {
			if Distance(p, center) < float64(radius) {
				found = false
				break
			}
		}
	}

	drawArch(img, p, radius, black, tickness, 1.0, 6.3)
	return p
}
