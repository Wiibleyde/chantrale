package bed

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"LsmsBot/internal/database/models"

	"github.com/fogleman/gg"
)

const (
	baseImagePath = "assets/beds.png"
	fontPath      = "assets/fonts/Montserrat-Bold.ttf"
	scale         = 3
	fontSize      = 11 * scale
	textRotation  = -math.Pi / 2 // 90° counter-clockwise
	strokeOffset  = 2
	dotRadius     = 4 * scale
)

type BedConfig struct {
	Letter string
	MinX   int
	MinY   int
	MaxX   int
	MaxY   int
}

var BedConfigs = []BedConfig{
	{Letter: "A", MinX: 494, MinY: 61, MaxX: 532, MaxY: 131},
	{Letter: "B", MinX: 378, MinY: 61, MaxX: 416, MaxY: 131},
	{Letter: "C", MinX: 37, MinY: 59, MaxX: 75, MaxY: 129},
	{Letter: "D", MinX: 502, MinY: 173, MaxX: 540, MaxY: 243},
	{Letter: "E", MinX: 436, MinY: 173, MaxX: 474, MaxY: 243},
	{Letter: "F", MinX: 373, MinY: 173, MaxX: 411, MaxY: 243},
}

func assignmentMap(assignments []models.BedAssignment) map[string]models.BedAssignment {
	m := make(map[string]models.BedAssignment, len(assignments))
	for _, a := range assignments {
		m[a.BedLetter] = a
	}
	return m
}

func GenerateBedImage(assignments []models.BedAssignment) ([]byte, error) {
	f, err := os.Open(baseImagePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	baseImg, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	bounds := baseImg.Bounds()
	w := bounds.Dx() * scale
	h := bounds.Dy() * scale
	dc := gg.NewContext(w, h)

	// Draw base image scaled up
	dc.Scale(scale, scale)
	dc.DrawImage(baseImg, 0, 0)
	dc.Identity()

	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, err
	}

	aMap := assignmentMap(assignments)

	for _, bed := range BedConfigs {
		assignment, occupied := aMap[bed.Letter]
		if !occupied {
			continue
		}

		bx := float64(bed.MinX) * scale
		by := float64(bed.MinY) * scale
		bw := float64(bed.MaxX-bed.MinX) * scale
		bh := float64(bed.MaxY-bed.MinY) * scale
		cx := bx + bw/2
		cy := by + bh/2

		// Status indicator dots (before rotation, positioned in bed space)
		if assignment.UnderArrest {
			dc.SetColor(color.NRGBA{R: 220, G: 30, B: 30, A: 255})
			dc.DrawCircle(bx+dotRadius+2, by+dotRadius+2, dotRadius)
			dc.Fill()
		}
		if assignment.Death {
			dc.SetColor(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
			dc.DrawCircle(bx+dotRadius+2, by+bh-dotRadius-2, dotRadius)
			dc.Fill()
		}

		// Draw text rotated 90° counter-clockwise
		dc.Push()
		dc.RotateAbout(textRotation, cx, cy)

		// Black stroke/border (4 offsets)
		dc.SetColor(color.Black)
		for _, off := range [][2]float64{{-strokeOffset, 0}, {strokeOffset, 0}, {0, -strokeOffset}, {0, strokeOffset}} {
			dc.DrawStringWrapped(assignment.Name, cx+off[0], cy+off[1], 0.5, 0.5, bh-8, 1.2, gg.AlignCenter)
		}

		// White text on top
		dc.SetColor(color.White)
		dc.DrawStringWrapped(assignment.Name, cx, cy, 0.5, 0.5, bh-8, 1.2, gg.AlignCenter)

		dc.Pop()
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
