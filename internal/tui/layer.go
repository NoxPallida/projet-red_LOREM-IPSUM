package tui

// Layer is a positioned off-screen buffer composited onto a Canvas
// Last writer wins per cell: a dialog drawn last hides the map beneath it
type Layer struct {
	X, Y    int
	Visible bool
	C       *Canvas
}

// NewLayer creates a visible w*h layer filled with spaces
func NewLayer(x, y, w, h int) Layer {
	return Layer{X: x, Y: y, Visible: true, C: NewCanvas(w, h)}
}

// DrawLayer composites one layer onto c. Nil/invisible/empty layers are skipped
// La case entiere est recopiee (lettre + couleurs) : le style survit
// Le clipping est fait a la main ici pour ne pas perdre les couleurs
// via Set (qui efface le style)
func (c *Canvas) DrawLayer(l Layer) {
	if c == nil || !l.Visible || l.C == nil {
		return
	}
	for j := 0; j < l.C.H; j++ {
		ty := l.Y + j
		if ty < 0 || ty >= c.H {
			continue
		}
		for i := 0; i < l.C.W; i++ {
			tx := l.X + i
			if tx < 0 || tx >= c.W {
				continue
			}
			c.cells[ty*c.W+tx] = l.C.cells[j*l.C.W+i]
		}
	}
}

// DrawLayers composites layers in slice order
func (c *Canvas) DrawLayers(layers []Layer) {
	for _, l := range layers {
		c.DrawLayer(l)
	}
}

// FuncLayer adapts a draw callback to the Layer model without
// allocating an off-screen buffer. Useful for full-screen backgrounds
type FuncLayer struct {
	Visible bool
	Draw    func(c *Canvas)
}

// DrawFuncLayers runs callbacks in order. Nil callbacks are skipped
func (c *Canvas) DrawFuncLayers(layers []FuncLayer) {
	if c == nil {
		return
	}
	for _, l := range layers {
		if !l.Visible || l.Draw == nil {
			continue
		}
		l.Draw(c)
	}
}
