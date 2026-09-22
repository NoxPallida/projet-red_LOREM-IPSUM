package cine

import (
	"strings"
	"testing"

	"runa/internal/tui"
)

func TestDrawFrameCentering(t *testing.T) {
	// Movie of size 10x10 with an 'O' at (5, 5)
	var frameLines []string
	var colorLines []string
	for y := 0; y < 10; y++ {
		row := ""
		for x := 0; x < 10; x++ {
			if x == 5 && y == 5 {
				row += "O"
			} else {
				row += " "
			}
		}
		frameLines = append(frameLines, row)
		colorLines = append(colorLines, "0000000000")
	}

	frameStr := ""
	for i, l := range frameLines {
		frameStr += l
		if i < len(frameLines)-1 {
			frameStr += "\n"
		}
	}
	colorStr := ""
	for i, l := range colorLines {
		colorStr += l
		if i < len(colorLines)-1 {
			colorStr += "\n"
		}
	}

	m := Movie{
		W:           10,
		H:           10,
		FPS:         10,
		Frames:      []string{frameStr},
		ColorFrames: []string{colorStr},
	}

	// Canvas larger: 20x20. Movie center is (5, 5), Canvas center is (10, 10)
	c := tui.NewCanvas(20, 20)
	DrawFrame(c, m, 0)

	lines := strings.Split(c.String(), "\n")
	if len(lines) != 20 || []rune(lines[10])[10] != 'O' {
		t.Fatalf("expected 'O' at centered coordinate (10, 10), got line %q", lines[10])
	}

	// Canvas smaller: 6x6. Offset is (6-10)/2 = -2, -2
	// Target of (5, 5) is (5-2, 5-2) = (3, 3), which is inside the 6x6 canvas
	cSmall := tui.NewCanvas(6, 6)
	DrawFrame(cSmall, m, 0)
	linesSmall := strings.Split(cSmall.String(), "\n")
	if len(linesSmall) != 6 || []rune(linesSmall[3])[3] != 'O' {
		t.Fatalf("expected 'O' at centered coordinate (3, 3) on small canvas, got line %q", linesSmall[3])
	}
}
