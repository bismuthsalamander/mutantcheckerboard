package main

import "fmt"

type Coord struct {
	X int
	Y int
}

func (c Coord) String() string {
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

type Delta struct {
	X int
	Y int
}

func (d Delta) String() string {
	if d == LEFT {
		return "LEFT"
	} else if d == RIGHT {
		return "RIGHT"
	} else if d == DOWN {
		return "DOWN"
	} else if d == UP {
		return "UP"
	}
	return fmt.Sprintf("(%d,%d)", d.X, d.Y)
}

func (d Delta) TurnCW() Delta {
	return Delta{
		X: d.Y * -1,
		Y: d.X,
	}
}

func (d Delta) TurnCCW() Delta {
	return Delta{
		X: d.Y,
		Y: d.X * -1,
	}
}

func (d Delta) Reverse() Delta {
	return Delta{
		d.X * -1,
		d.Y * -1,
	}
}

func (d Delta) Times(n int) Delta {
	return Delta{d.X * n, d.Y * n}
}

var LEFT = Delta{-1, 0}
var RIGHT = Delta{1, 0}
var UP = Delta{0, -1}
var DOWN = Delta{0, 1}

var DIRECTIONS = []Delta{LEFT, RIGHT, UP, DOWN}

func (c Coord) Minus(o Coord) Delta {
	d := Delta{
		c.X - o.X,
		c.Y - o.Y,
	}
	if d.X < 0 {
		d.X *= -1
	}
	if d.Y < 0 {
		d.Y *= -1
	}
	return d
}

func (c Coord) MHDist(o Coord) int {
	val := 0
	if c.X > o.X {
		val += c.X - o.X
	} else {
		val += o.X - c.X
	}
	if c.Y > o.Y {
		val += c.Y - o.Y
	} else {
		val += o.Y - c.Y
	}
	return val
}

func (c Coord) Plus(d Delta) Coord {
	return Coord{
		X: c.X + d.X,
		Y: c.Y + d.Y,
	}
}

type Cell int8

const UNKNOWN = 0
const PAINTED = 1
const CLEAR = 2

func (c Cell) Ch() rune {
	if c == PAINTED {
		return 'X'
	} else if c == CLEAR {
		return '·'
	}
	return ' '
}

func (c Cell) String() string {
	return string(c.Ch())
}

func IntToCh(n int) rune {
	if n < 10 {
		return rune(int('0') + n)
	} else if n <= 35 {
		return rune(int('a') + (n - 10))
	}
	return '?'
}

func CharToNum(ch rune) (int, bool) {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0'), true
	}
	if ch >= 'a' && ch <= 'z' {
		return int(ch-'a') + 10, true
	}
	return 0, false
}

func MakeGrid(w, h int) [][]Cell {
	g := make([][]Cell, 0, h)
	for i := 0; i < int(h); i++ {
		g = append(g, make([]Cell, w))
	}
	return g
}

func MakeNumGrid(w, h int) [][]int {
	g := make([][]int, 0, h)
	for i := 0; i < int(h); i++ {
		g = append(g, make([]int, w))
	}
	return g
}

func MakeRegionGrid(w, h int) [][][]*[]Coord {
	g := make([][][]*[]Coord, 0, h)
	for i := 0; i < int(h); i++ {
		g = append(g, make([][]*[]Coord, 0, w))
		for j := 0; j < w; j++ {
			g[i] = append(g[i], make([]*[]Coord, 0))
		}
	}
	return g
}

func MakeAllowedSets(w, h int, order int) [][]*Set[int] {
	g := make([][]*Set[int], 0, h)
	for i := 0; i < h; i++ {
		g = append(g, make([]*Set[int], 0, w))
		for j := 0; j < w; j++ {
			g[i] = append(g[i], NewNumSet(order))
		}
	}
	return g
}
