package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"sort"
	"strings"
	"time"
)

const (
	SCREEN_WIDTH  = 640
	SCREEN_HEIGHT = 480
	NAME          = "Pong"
	PADDING       = 10
	VERSION       = "0.1"
)

type PongError struct {
	SDLError error
	Msg      string
}

func NewPongError(msg string) *PongError {
	e := &PongError{
		SDLError: sdl.GetError(),
		Msg:      msg,
	}
	return e
}

func (e *PongError) Error() string {
	// we don't need do much if we don't have sdl errors present
	if e.SDLError == nil {
		return e.Msg
	}
	errMsg := strings.Join(
		[]string{
			e.Msg,
			e.SDLError.Error(),
		},
		": ",
	)
	return errMsg
}

type Object struct {
	W, H, X, Y float64
	DX, DY     float64
}

type Direction bool

const (
	UP   Direction = true
	DOWN Direction = false
)

// UpdatePaddle updates a paddle
func UpdatePaddle(universeBus chan map[string]Object, errChan chan error, paddle string, d Direction) {
	v := 0.0
	switch d {
	case UP:
		v = v - 0.01
	case DOWN:
		v = v + 0.01
	}
	u := <-universeBus
	defer func() {
		universeBus <- u
	}()
	tmp, ok := u[paddle]
	if !ok {
		errChan <- fmt.Errorf(`Key "%s" does not exist in our universe!`, paddle)
		return
	}
	if 0 < tmp.Y+v && tmp.Y+v < 1 {
		tmp.Y = tmp.Y + v
	}
	// assign updated paddle back as we don't have a pointer (yet)
	u[paddle] = tmp
}

type WallIntersection struct {
	IntersectAt float64
	Wall        *Line
}

// WallIntersections returns wall intersections
func WallIntersections(walls []*Line, line *Line) []WallIntersection {
	intersections := make([]WallIntersection, 4)
	for _, wall := range walls {
		log.Print("wall: ", wall)
		var intersection WallIntersection
		intersection.IntersectAt = line.Intersect(wall)
		intersection.Wall = wall
	}
	return intersections
}

func UpdateBall(universeBus chan map[string]Object, errChan chan error, d time.Duration) {
	walls := []*Line{
		&Line{&Vector2d{0, 0}, &Vector2d{0, 1}},
		&Line{&Vector2d{0, 1}, &Vector2d{1, 1}},
		&Line{&Vector2d{1, 1}, &Vector2d{1, 0}},
		&Line{&Vector2d{1, 0}, &Vector2d{0, 0}},
	}

	u := <-universeBus
	defer func() {
		universeBus <- u
	}()
	ball := u["Ball"]
	defer func() {
		u["Ball"] = ball
	}()

	initialDir := &Vector2d{ball.DX * d.Seconds(), ball.DY * d.Seconds()}
	initialPos := &Vector2d{ball.X, ball.Y}

	pos := initialPos.Copy()
	dir := initialDir.Copy()
	newPos := pos.Add(dir)
	line := &Line{pos, newPos}
	for line.Vector2d().Len() > 0 {

		wallmap := make(map[float64]*Line, 4)

		h := make([]float64, 4)

		for i, wall := range walls {
			h[i] = line.Intersect(wall)
			wallmap[h[i]] = wall
		}

		sort.Float64s(h)

		c := 0
		for _, f := range h {
			if 0 < f && f < 1 {
				// we do intersect
				hitPos := pos.Add(dir.Scale(f))
				remainder := line.Vector2d().Len() - f
				dir = wallmap[f].Vector2d().Reflect(dir.Scale(remainder))
				newPos = hitPos.Add(dir)
				pos = hitPos
				line = &Line{pos, newPos}
				c = c + 1
			}
		}
		if c == 0 {
			newPos = pos.Add(dir)
			line = &Line{pos, newPos}
			break
		}
	}
	ball.X = newPos[0]
	ball.Y = newPos[1]
}

func main() {
	_, err := NewEngine("Pong",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		SCREEN_WIDTH,
		SCREEN_HEIGHT)
	if err != nil {
		log.Fatal("Initialization of sdl failed: ", err)
	}
}
