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
	W      int
	H      int
	Dirty  bool
	Inited bool
}
type RectBinBoard struct {
	RectBoard
	Grid  [][]Cell
	Guess [][]Cell
}

func RectBinBoardFromLines(input []string) *RectBinBoard {
	w := len(input[0])
	h := len(input)
	return &RectBinBoard{
		RectBoard: RectBoard{
			W: w,
			H: h,
		},
		Grid:  MakeGrid(w, h),
		Guess: MakeGrid(w, h),
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

func (b *RectBoard) IsValid(c Coord) bool {
	return c.X >= 0 && c.Y >= 0 && c.X < b.W && c.Y < b.H
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

func (b *RectBinBoard) Get(c Coord) Cell {
	if b.Guess[c.Y][c.X] != UNKNOWN {
		return b.Guess[c.Y][c.X]
	}
	return b.Grid[c.Y][c.X]
}

func (b *RectBinBoard) ClearGuess() {
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		b.Guess[c.Y][c.X] = UNKNOWN
	}
}

func (b *RectBinBoard) IsPainted(c Coord) bool {
	return b.IsValid(c) && b.Get(c) == PAINTED
}
func (b *RectBinBoard) IsClear(c Coord) bool {
	return b.IsValid(c) && b.Get(c) == CLEAR
}
func (b *RectBinBoard) IsUnknown(c Coord) bool {
	return b.IsValid(c) && b.Get(c) == UNKNOWN
}
func (b *RectBinBoard) IsComplete() (bool, Coord) {
	for y, row := range b.Grid {
		for x, cell := range row {
			if cell == UNKNOWN {
				return false, Coord{x, y}
			}
		}
	}
	return true, Coord{}
}
func (b *RectBinBoard) IsSolved() bool {
	return false
}

func (b *RectBinBoard) Set(c Coord, v Cell) (bool, error) {
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

func (b *RectBinBoard) EachCell(cb func(c Coord, v Cell) bool) {
	for y, row := range b.Grid {
		for x, cell := range row {
			if cb(Coord{x, y}, cell) {
				return
			}
		}
	}
}

func (b *RectBinBoard) EachNeighbor(start Coord, cb func(c Coord, v Cell) bool) {
	for _, dir := range DIRECTIONS {
		c := start.Plus(dir)
		if b.IsValid(c) {
			cb(c, b.Get(c))
		}
	}
}

// SliceContains returns true iff the slice haystack contains the value needle.
func SliceContains(haystack []int, needle int) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}
