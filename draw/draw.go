package draw

import (
	basic "github.com/strickyak/basic_basic"
	"github.com/strickyak/canvas"

	"bufio"
	"io"
	"os"
)

const WIDTH = 800
const HEIGHT = 800
const MARGIN = 5.0
const SPOTSIZE = 0.5

// Decimal2RGB converts decimal number RGB to byte colors.
func Decimal2RGB(dec float64) (byte, byte, byte) {
	d := int(dec)
	b := (byte)((d % 10) * 255.0 / 9.0)
	g := (byte)(((d / 10) % 10) * 255.0 / 9.0)
	r := (byte)((d / 100) * 255.0 / 9.0)
	return r, g, b
}

type drawing struct {
	canvas *canvas.Canvas
  px, py float64
}

func (o *drawing) init() {
	if o.canvas == nil {
		o.canvas = canvas.NewCanvas(WIDTH, HEIGHT)
    o.px = 0.0 + MARGIN
    o.py = 100.0 - MARGIN
	}
}

func (o *drawing) Putchar(ch byte) {
  o.init()
  start := int(ch)-32
  if start<0 || start>96 {
    start=3
  }
  start *= 5
	for i := 0; i < 5; i++ {
		for j := 0; j < 7; j++ {
			x, y := o.px+float64(i)*SPOTSIZE, o.py-float64(j)*SPOTSIZE
			var c byte = 0 // 50
			if ((basic.Font[start+i] >> uint(j)) & 1) != 0 {
				c = 255 // 250
			}
			o.canvas.PaintTriangle(int(x/100.0*WIDTH), int(y/100.0*HEIGHT),
				int((x+SPOTSIZE)/100.0*WIDTH), int(y/100.0*HEIGHT),
				int(x/100.0*WIDTH), int((y+SPOTSIZE)/100.0*HEIGHT),
				canvas.RGB(c, c, c))
		}
	}
  o.px += SPOTSIZE * 7
  if o.px > 100 - MARGIN {
    o.px = MARGIN
    o.py -= SPOTSIZE*10
    if o.py < MARGIN {
      o.py = 100 - MARGIN
    }
  }
}

func (o *drawing) clear(t *basic.Terp, args []float64) float64 {
	o.init()
	var r, g, b byte // Default is black.
	if len(args) > 0 {
		r, g, b = Decimal2RGB(args[0]) // Use given color.
	}
	o.canvas.Fill(0, 0, WIDTH, HEIGHT, canvas.RGB(r, g, b))
	return 0
}

func (o *drawing) point(t *basic.Terp, args []float64) float64 {
	o.init()
	var r, g, b byte // Default is black.
	if len(args) > 2 {
		r, g, b = Decimal2RGB(args[2]) // Use given color.
	}
	o.canvas.Set(int(args[0]/100*WIDTH), int(args[1]/100*HEIGHT), canvas.RGB(r, g, b))
	return 0
}

func (o *drawing) triangle(t *basic.Terp, args []float64) float64 {
	o.init()
	var r, g, b byte // Default is black.
	if len(args) > 6 {
		r, g, b = Decimal2RGB(args[6]) // Use given color.
	}
	o.canvas.PaintTriangle(int(args[0]/100.0*WIDTH), int(args[1]/100.0*HEIGHT),
		int(args[2]/100.0*WIDTH), int(args[3]/100.0*HEIGHT),
		int(args[4]/100.0*WIDTH), int(args[5]/100.0*HEIGHT),
		canvas.RGB(r, g, b))
	return 0
}

func (o *drawing) fonttest(t *basic.Terp, _ []float64) float64 {
	o.init()
	x0, y0 := 5.0, 98.0
	for i := 0; i < len(basic.Font); i++ {
		for j := 0; j < 8; j++ {
			x, y := x0+float64(i), y0-float64(j)

			var c byte = 50
			if ((basic.Font[i] >> uint(j)) & 1) == 1 {
				c = 250
			}
			o.canvas.PaintTriangle(int(x/100.0*0.5*WIDTH), int(y/100.0*0.5*HEIGHT),
				int((x+1)/100.0*0.5*WIDTH), int(y/100.0*0.5*HEIGHT),
				int(x/100.0*0.5*WIDTH), int((y+1)/100.0*0.5*HEIGHT),
				canvas.RGB(c/5, c, c/5))
		}
		if i%5 == 4 {
			x0 += 1
		}
		if i%60 == 59 {
			x0, y0 = 5 - float64(i), y0-10
		}
	}
	return 0.0
}

func Register(t *basic.Terp) *drawing {
	o := &drawing{}
	t.AddExtension("clear", basic.Callable(o.clear))
	t.AddExtension("point", basic.Callable(o.point))
	t.AddExtension("triangle", basic.Callable(o.triangle))
	t.AddExtension("fonttest", basic.Callable(o.fonttest))
	return o
}
func (o *drawing) HasImage() bool {
	return o.canvas != nil
}
func (o *drawing) WritePng(w io.Writer) {
	o.init()
	o.canvas.WritePng(w)
}
func (o *drawing) SavePng(filename string) {
	o.init()
	w, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	bw := bufio.NewWriter(w)
	o.canvas.WritePng(bw)
	bw.Flush()
	w.Close()
}
