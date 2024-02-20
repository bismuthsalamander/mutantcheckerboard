package main

import "fmt"

type BinBoard interface {
	MarkPainted(Coord) (bool, error)
	MarkClear(Coord) (bool, error)
	IsPainted(Coord) bool
	IsClear(Coord) bool
	IsUnknown(Coord) bool
	IsValid(Coord) bool
	IsComplete() bool
	IsSolved() bool
	Init(string)
	InitDone() bool
	PostMark(Coord, Cell) bool
	SetDirty()
	ClearDirty()
	IsDirty() bool
}

type RectBoard struct {
	W      uint8
	H      uint8
	Grid   [][]Cell
	Dirty  bool
	Inited bool
}

func RectBoardFromLines(input []string) *RectBoard {
	w := uint8(len(input[0]))
	h := uint8(len(input))
	return &RectBoard{
		W:    w,
		H:    h,
		Grid: MakeGrid(w, h),
	}
}

func (b *RectBoard) TopLeft() Coord {
	return Coord{0, 0}
}

func (b *RectBoard) Next(c Coord) Coord {
	c.X++
	if c.X == b.W {
		c.X = 0
		c.Y++
	}
	return c
}

func (b *RectBoard) InitDone() bool {
	return b.Inited
}

func (b *RectBoard) SetDirty() {
	b.Dirty = true
}

func (b *RectBoard) ClearDirty() {
	b.Dirty = false
}

func (b *RectBoard) IsDirty() bool {
	return b.Dirty
}

func (b *RectBoard) Get(c Coord) Cell {
	return b.Grid[c.Y][c.X]
}

func (b *RectBoard) IsPainted(c Coord) bool {
	return b.IsValid(c) && b.Get(c) == PAINTED
}
func (b *RectBoard) IsClear(c Coord) bool {
	return b.IsValid(c) && b.Get(c) == CLEAR
}
func (b *RectBoard) IsUnknown(c Coord) bool {
	return b.IsValid(c) && b.Get(c) == UNKNOWN
}
func (b *RectBoard) IsValid(c Coord) bool {
	return c.X < b.W && c.Y < b.H
}
func (b *RectBoard) IsComplete() bool {
	for _, row := range b.Grid {
		for _, cell := range row {
			if cell == UNKNOWN {
				return false
			}
		}
	}
	return true
}
func (b *RectBoard) IsSolved() bool {
	return false
}

func (b *RectBoard) Set(c Coord, v Cell) (bool, error) {
	if !b.IsValid(c) {
		return false, fmt.Errorf("coordinate (%d,%d) not valid on board of size (%d,%d)", c.X, c.Y, b.W, b.H)
	}
	if b.IsPainted(c) {
		return false, nil
	}
	if b.IsClear(c) {
		return false, fmt.Errorf("coordinate (%d,%d) is already clear; cannot paint it", c.X, c.Y)
	}
	b.Grid[c.Y][c.X] = v
	return true, nil
}

func (b *RectBoard) EachCell(cb func(c Coord, v Cell) bool) {
	for y, row := range b.Grid {
		for x, cell := range row {
			if cb(MkCoord(x, y), cell) {
				return
			}
		}
	}
}

func (b *RectBoard) EachNeighbor(start Coord, cb func(c Coord, v Cell) bool) {
	for _, dir := range DIRECTIONS {
		c := start.Plus(dir)
		if b.IsValid(c) {
			cb(c, b.Get(c))
		}
	}
}
