package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log"
	"math"
	"os"

	"github.com/toikarin/sgf"
)

func main() {
	iPath, oPath, err := parseArgs()
	if err != nil {
		defer usage()
		log.Fatal(err)
	}

	g, err := sgfToGif(iPath)
	if err != nil {
		log.Fatal(err)
	}

	err = save(oPath, g)
	if err != nil {
		log.Fatal(err)
	}
}

func parseArgs() (string, string, error) {
	if len(os.Args) != 3 {
		return "", "", fmt.Errorf("bad number of arguments")
	}
	return os.Args[1], os.Args[2], nil
}

func save(path string, g *gif.GIF) (err error) {
	f, err := os.OpenFile(path,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	defer func() {
		errClose := f.Close()
		if err == nil {
			err = errClose
		}
	}()

	return gif.EncodeAll(f, g)
}

func usage() {
	log.Printf("usage: %s input_sgf_file output_gif_file\n", os.Args[0])
}

func sgfToGif(path string) (*gif.GIF, error) {
	c, err := sgf.ParseSgfFile(path)
	if err != nil {
		return nil, err
	}

	game, err := firstGame(c)
	if err != nil {
		return nil, err
	}

	moves, err := movesFromGame(game)
	if err != nil {
		return nil, err
	}

	frames, err := movesToFrames(moves)
	if err != nil {
		return nil, err
	}

	g, err := framesToGif(frames)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func firstGame(c *sgf.Collection) (*sgf.GameTree, error) {
	switch n := len(c.GameTrees); n {
	case 0:
		return nil, fmt.Errorf("no games in the file")
	case 1:
		break // OK
	default:
		log.Printf("found %d games: using the first, ignoring the rest", n)
	}

	return c.GameTrees[0], nil
}

type move struct {
	white bool
	x     int
	y     int
}

func notAMove(p *sgf.Property) bool {
	return p.Ident != "B" && p.Ident != "W"
}

func movesFromGame(g *sgf.GameTree) ([]*move, error) {
	ret := []*move{}
	for _, n := range g.Nodes {
		for _, p := range n.Properties {
			if notAMove(p) {
				continue
			}
			if len(p.Values) != 1 {
				return nil, fmt.Errorf("malformed move: %#v", p.Values)
			}

			x, y, err := lettersToCoords(p.Values[0])
			if err != nil {
				return nil, err
			}

			m := &move{
				white: p.Ident == "W",
				x:     x,
				y:     y,
			}
			ret = append(ret, m)
		}
	}
	return ret, nil
}

func lettersToCoords(letters string) (int, int, error) {
	if len(letters) != 2 {
		return 0, 0, fmt.Errorf("malformed move value: %s", letters)
	}
	x := int(letters[0] - 'a')
	y := int(letters[1] - 'a')
	return x, y, nil
}

func movesToFrames(ms []*move) ([]*image.Paletted, error) {
	ret := []*image.Paletted{}

	var frame *image.Paletted
	var err error

	for _, m := range ms {
		frame, err = newFrame(m, frame)
		if err != nil {
			return nil, err
		}
		ret = append(ret, frame)
	}
	return ret, nil
}

func framesToGif(frames []*image.Paletted) (*gif.GIF, error) {
	g := &gif.GIF{LoopCount: len(frames)}
	for _, f := range frames {
		g.Image = append(g.Image, f)
		g.Delay = append(g.Delay, delay)
	}
	return g, nil
}

var palette = []color.Color{
	color.RGBA{0xE6, 0xBF, 0x83, 0xFF}, // wood
	color.Black,
	color.White,
}

const (
	background = iota
	black
	white
)

const (
	delay         = 100 // delay between frames in 10ms units
	stoneDiameter = 40  // pixels
	boardSize     = 19
)

// side of the board in pixels
func side() int {
	return boardSize*stoneDiameter + 2
}

func newFrame(m *move, old *image.Paletted) (*image.Paletted, error) {
	side := side()
	rect := image.Rect(0, 0, side, side)
	img := image.NewPaletted(rect, palette)

	if old == nil {
		// fill background
		for i := 0; i < side; i++ {
			for j := 0; j < side; j++ {
				img.SetColorIndex(i, j, background)
			}
		}

		// vertical lines
		for i := 0; i < boardSize; i++ {
			for j := stoneDiameter / 2; j < side-stoneDiameter/2; j++ {
				x := i*stoneDiameter + stoneDiameter/2
				img.SetColorIndex(x, j, black)
			}
		}

		// horizontal lines
		for i := 0; i < boardSize; i++ {
			for j := stoneDiameter / 2; j < side-stoneDiameter/2; j++ {
				y := i*stoneDiameter + stoneDiameter/2
				img.SetColorIndex(j, y, black)
			}
		}
	} else {
		// copy from old
		for i := 0; i < side; i++ {
			for j := 0; j < side; j++ {
				img.SetColorIndex(i, j, old.ColorIndexAt(i, j))
			}
		}
	}

	drawMove(img, m)

	return img, nil
}

func drawMove(img *image.Paletted, m *move) {
	fmt.Printf("%#v\n", m)
	side := side()
	x := stoneDiameter/2 + m.x*stoneDiameter
	y := stoneDiameter/2 + m.y*stoneDiameter
	for i := 0; i < side; i++ {
		for j := 0; j < side; j++ {
			if dist(i, j, x, y) <= stoneDiameter/2 {
				var color uint8 = black
				if m.white {
					color = white
				}
				img.SetColorIndex(i, j, color)
			}
		}
	}
}

func dist(x1, y1, x2, y2 int) int {
	x := x2 - x1
	if x < 0 {
		x = -x
	}
	y := y2 - y1
	if y < 0 {
		y = -y
	}
	h := float64(x*x + y*y)
	sq := math.Sqrt(h)
	return int(sq)
}
