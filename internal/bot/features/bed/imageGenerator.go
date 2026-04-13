package bed

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"

	"LsmsBot/internal/database/models"

	"github.com/fogleman/gg"
)

const (
	baseImagePath = "assets/beds.png"
	fontPath      = "assets/fonts/Montserrat-Bold.ttf"
	scale         = 1
	fontSize      = 45
	textRotation  = -math.Pi / 2 // 90° counter-clockwise
	strokeOffset  = 2
	dotRadius     = 12
)

type BedConfig struct {
	Letter string
	MinX   int
	MinY   int
	MaxX   int
	MaxY   int
}

var BedConfigs = []BedConfig{
	{Letter: "A", MinX: 104, MinY: 240, MaxX: 267, MaxY: 444},
	{Letter: "B", MinX: 1248, MinY: 240, MaxX: 1411, MaxY: 444},
	{Letter: "C", MinX: 1634, MinY: 240, MaxX: 1797, MaxY: 444},
	{Letter: "D", MinX: 1230, MinY: 614, MaxX: 1393, MaxY: 818},
	{Letter: "E", MinX: 1450, MinY: 614, MaxX: 1613, MaxY: 818},
	{Letter: "F", MinX: 1670, MinY: 614, MaxX: 1833, MaxY: 818},
}

func truncateLines(dc *gg.Context, text string, maxWidth float64, maxLines int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		candidate := current + " " + w
		cw, _ := dc.MeasureString(candidate)
		if cw > maxWidth {
			lines = append(lines, current)
			current = w
		} else {
			current = candidate
		}
	}
	lines = append(lines, current)

	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}

	ellipsis := "..."
	last := lines[maxLines-1]
	for last != "" {
		candidate := last + ellipsis
		cw, _ := dc.MeasureString(candidate)
		if cw <= maxWidth {
			lines[maxLines-1] = candidate
			break
		}
		runes := []rune(last)
		last = strings.TrimRight(string(runes[:len(runes)-1]), " ")
	}
	return strings.Join(lines[:maxLines], "\n")
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

		if assignment.UnderArrest {
			dc.SetColor(color.NRGBA{R: 13, G: 0, B: 172, A: 255})
			dc.DrawCircle(bx+dotRadius+2, by+dotRadius+2, dotRadius)
			dc.Fill()
		}
		if assignment.Death {
			dc.SetColor(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
			dc.DrawCircle(bx+dotRadius+2, by+bh-dotRadius-2, dotRadius)
			dc.Fill()
		}

		dc.Push()
		dc.RotateAbout(textRotation, cx, cy)

		displayName := truncateLines(dc, assignment.Name, bh-8, 3)

		dc.SetColor(color.Black)
		for _, off := range [][2]float64{{-strokeOffset, 0}, {strokeOffset, 0}, {0, -strokeOffset}, {0, strokeOffset}} {
			dc.DrawStringWrapped(displayName, cx+off[0], cy+off[1], 0.5, 0.5, bh-8, 1.2, gg.AlignCenter)
		}

		dc.SetColor(color.White)
		dc.DrawStringWrapped(displayName, cx, cy, 0.5, 0.5, bh-8, 1.2, gg.AlignCenter)

		dc.Pop()
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
