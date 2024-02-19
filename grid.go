package main

import "fmt"

type Coord struct {
	X uint8
	Y uint8
}

func (c Coord) String() string {
	return fmt.Sprintf("(%d,%d)", c.X, c.Y)
}

type Delta struct {
	X int8
	Y int8
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

func (d Delta) Times(n uint8) Delta {
	return Delta{d.X * int8(n), d.Y * int8(n)}
}

var LEFT = Delta{-1, 0}
var RIGHT = Delta{1, 0}
var UP = Delta{0, -1}
var DOWN = Delta{0, 1}

var DIRECTIONS = []Delta{LEFT, RIGHT, UP, DOWN}

func (c Coord) Minus(o Coord) Delta {
	d := Delta{
		int8(c.X) - int8(o.X),
		int8(c.Y) - int8(o.Y),
	}
	if d.X < 0 {
		d.X *= -1
	}
	if d.Y < 0 {
		d.Y *= -1
	}
	return d
}

func (c Coord) MHDist(o Coord) uint8 {
	val := uint8(0)
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
		X: uint8(int8(c.X) + d.X),
		Y: uint8(int8(c.Y) + d.Y),
	}
}

func MkCoord(x, y int) Coord {
	return Coord{
		X: uint8(x),
		Y: uint8(y),
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
		return '.'
	}
	return ' '
}

func (c Cell) String() string {
	return string(c.Ch())
}

func NumToCh(n uint8) rune {
	if n < 10 {
		return rune(uint8('0') + n)
	} else if n <= 35 {
		return rune(uint8('a') + (n - 10))
	}
	return '?'
}

func MakeGrid(w, h uint8) [][]Cell {
	g := make([][]Cell, 0, h)
	for i := 0; i < int(h); i++ {
		g = append(g, make([]Cell, w))
	}
	return g
}
