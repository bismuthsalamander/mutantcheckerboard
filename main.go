package main

import (
	"fmt"
	"os"
	"strings"
)

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
	return d.Times(-1)
}

func (d Delta) Times(n int8) Delta {
	return Delta{d.X * n, d.Y * n}
}

var LEFT = Delta{-1, 0}
var RIGHT = Delta{1, 0}
var UP = Delta{0, -1}
var DOWN = Delta{0, 1}

var DIRECTIONS = []Delta{LEFT, RIGHT, UP, DOWN}

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
	PostMark(Coord, Cell) bool
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

type RectBoard struct {
	W    uint8
	H    uint8
	Grid [][]Cell
}

func MakeGrid(w, h uint8) [][]Cell {
	g := make([][]Cell, 0, h)
	for i := 0; i < int(h); i++ {
		g = append(g, make([]Cell, w))
	}
	return g
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

func (b *RectBoard) Get(c Coord) Cell {
	return b.Grid[c.Y][c.X]
}

func (b *RectBoard) IsPainted(c Coord) bool {
	return b.Get(c) == PAINTED
}
func (b *RectBoard) IsClear(c Coord) bool {
	return b.Get(c) == CLEAR
}
func (b *RectBoard) IsUnknown(c Coord) bool {
	return b.Get(c) == UNKNOWN
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

// func (b *RectBoard) PostMark(_ Coord, _ Cell) {}
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

func (b *RangeBoard) Mark(c Coord, v Cell) (bool, error) {
	res, err := b.Set(c, v)
	if res == false {
		return res, err
	}
	b.PostMark(c, v)
	return res, err
}

func (b *RangeBoard) MarkPainted(c Coord) (bool, error) {
	return b.Mark(c, PAINTED)
}

func (b *RangeBoard) MarkClear(c Coord) (bool, error) {
	return b.Mark(c, CLEAR)
}

type Wing struct {
	Dir      Delta
	Min      uint8
	Max      uint8
	IsCapped bool
}

type Cross struct {
	Root     Coord
	Size     uint8
	Wings    map[Delta]*Wing
	IsCapped bool
}

func (c *Cross) String() string {
	return string(NumToCh(c.Size))
}

type RangeBoard struct {
	RectBoard
	Crosses    [][]*Cross
	AllCrosses []*Cross
}

func MakeCrosses(w, h uint8) [][]*Cross {
	c := make([][]*Cross, 0, h)
	for i := 0; i < int(h); i++ {
		c = append(c, make([]*Cross, w))
	}
	return c
}

func CharToNum(ch rune) (uint8, bool) {
	if ch >= '0' && ch <= '9' {
		return uint8(ch - '0'), true
	}
	if ch >= 'a' && ch <= 'z' {
		return uint8(ch-'a') + 10, true
	}
	return 0, false
}

func (b *RangeBoard) MakeWings(c Coord) map[Delta]*Wing {
	mp := make(map[Delta]*Wing)
	mp[LEFT] = &Wing{LEFT, 0, c.X, false}
	mp[RIGHT] = &Wing{RIGHT, 0, b.W - (c.X + 1), false}
	mp[UP] = &Wing{UP, 0, c.Y, false}
	mp[DOWN] = &Wing{DOWN, 0, b.H - (c.Y + 1), false}
	return mp
}

func (b *RangeBoard) String() string {
	out := "+" + strings.Repeat("-", int(b.W)) + "+\n"
	for y, row := range b.Grid {
		out += "|"
		for x, cell := range row {
			if b.Crosses[y][x] != nil {
				out += b.Crosses[y][x].String()
			} else {
				out += cell.String()
			}
		}
		out += "|"
		if y != int(b.H)-1 {
			out += "\n"
		}
	}
	out += "\n+" + strings.Repeat("-", int(b.W)) + "+"
	return out
}

func (b *RangeBoard) StringVerbose() string {
	out := b.String()
	out += "\n\nCrosses:\n"
	for _, row := range b.Crosses {
		for _, cross := range row {
			if cross == nil {
				continue
			}
			out += fmt.Sprintf("Cross at %s sz %d\n", cross.Root, cross.Size)
			for dir, wing := range cross.Wings {
				out += fmt.Sprintf("\tWing %s [%d-%d] %v\n", dir, wing.Min, wing.Max, wing.IsCapped)
			}
		}
	}
	return out
}

func (b *RangeBoard) PostMark(c Coord, v Cell) {
	/**
	If painted:
		mark adj clear
		limit wing maxes along all four axes (THIS CAN STOP AFTER ACCOUNTING ANOTHER PAINTED CELL)
	If clear:
		increase minimum size along all four axes until we meet a non-clear cell

	SHORTCUT

	If painted:
		mark adj clear
	traverse four axes until we find a painted cell
	for each cross found, update its wing sizes?
	*/
	/**
	 */
	// Clear adjacent cells to paint
	if v == PAINTED {
		for _, dir := range DIRECTIONS {
			new := c.Plus(dir)
			if b.IsValid(new) {
				//TODO: are we gonna cause a problem? calling Mark multiple times? cascade or something?
				//TODO: alternatively, cascade a "dirty" mark to affected crosses, then update the dirty ones?
				b.MarkClear(new)
			}
		}
	}
	// Skipping the cross updates for now; a normal check will find these (albeit less efficiently)
}

func (b *RangeBoard) CheckAllCaps(c *Cross) {
	for _, w := range c.Wings {
		if w.Min == w.Max && !w.IsCapped {
			b.FinishWing(c, w)
		}
	}
}

func (c *Cross) MarkWingCapped(w *Wing) {
	w.IsCapped = true
	for _, wg := range c.Wings {
		if !wg.IsCapped {
			return
		}
	}
	c.IsCapped = true
}

func (b *RangeBoard) FinishCross(cross *Cross) {
	fmt.Printf("Finishing cross at %s\n", cross.Root)
	for _, wing := range cross.Wings {
		wing.Max = wing.Min
		b.FinishWing(cross, wing)
	}
}

func (b *RangeBoard) FinishWing(cross *Cross, w *Wing) {
	fmt.Printf("Finishing wing at %s %s (MM is %d)\n", cross.Root, w.Dir, w.Min)
	coord := cross.Root
	for i := uint8(1); i <= w.Min; i++ {
		coord = coord.Plus(w.Dir)
		if !b.IsValid(coord) {
			//TODO: report an error cleanly
			fmt.Printf("Error expanding wing %v at cell %d\n", *w, i)
			return
		}
		b.MarkClear(coord)
	}
	cap := cross.Root.Plus(w.Dir.Times(int8(w.Min + 1)))
	if b.IsValid(cap) {
		b.MarkPainted(cap)
	}
	cross.MarkWingCapped(w)
}

func (b *RangeBoard) UpdateWingRanges() {
	for _, cross := range b.AllCrosses {
		if cross.IsCapped {
			continue
		}
		for dir, wing := range cross.Wings {
			coord := cross.Root.Plus(dir)
			wingsz := uint8(0)
			fmt.Printf("Wing at %s-%s started [%d-%d]\n", cross.Root, wing.Dir, wing.Min, wing.Max)
			for b.IsValid(coord) {
				if b.IsClear(coord) {
					wingsz++
					if wingsz > wing.Min {
						fmt.Printf("Raising min to %d\n", wingsz)
						wing.Min = wingsz
					}
				} else if b.IsPainted(coord) {
					if wingsz < wing.Max {
						wing.Max = wingsz
					}
					break
				} else if b.IsUnknown(coord) {
					wingsz++
					if wingsz <= wing.Min {
						fmt.Printf("Clearing %s from cross at %s\n", coord, cross.Root)
						b.MarkClear(coord)
					} else {
						break
					}
				}
				coord = coord.Plus(dir)
			}
			if wing.Min == wing.Max && !wing.IsCapped {
				b.FinishWing(cross, wing)
			}
		}
		sz := uint8(1)
		for _, wing := range cross.Wings {
			sz += wing.Min
		}
		if sz == cross.Size {
			b.FinishCross(cross)
		}
	}
}

func (b *RangeBoard) ExpandWingTo(c *Cross, w *Wing, sz uint8) {
	if w.Min >= sz {
		return
	}
	w.Min = sz
	for i := int8(1); i <= int8(sz); i++ {
		b.MarkClear(c.Root.Plus(w.Dir.Times(i)))
	}
}

func (b *RangeBoard) ExpandWingMinimums() {
	for _, cross := range b.AllCrosses {
		if cross.IsCapped {
			continue
		}
		for _, dir := range DIRECTIONS {
			othersMax := cross.Wings[dir.TurnCW()].Max + cross.Wings[dir.TurnCCW()].Max + cross.Wings[dir.Reverse()].Max
			if othersMax+1 < cross.Size {
				fmt.Printf("Wing %s->%s must be at least %d\n", cross.Root, dir, cross.Size-(othersMax+1))
				b.ExpandWingTo(cross, cross.Wings[dir], cross.Size-(othersMax+1))
			}
		}
	}
}

func RangeBoardFromLines(input []string) *RangeBoard {
	rect := RectBoardFromLines(input)
	rg := RangeBoard{
		RectBoard:  *rect,
		Crosses:    MakeCrosses(rect.W, rect.H),
		AllCrosses: make([]*Cross, 0),
	}
	for y, row := range input {
		for x, ch := range row {
			if val, ok := CharToNum(ch); ok {
				c := MkCoord(x, y)
				rg.Crosses[y][x] = &Cross{
					Root:     Coord{X: uint8(x), Y: uint8(y)},
					Size:     val,
					Wings:    rg.MakeWings(c),
					IsCapped: false,
				}
				rg.AllCrosses = append(rg.AllCrosses, rg.Crosses[y][x])
				rg.MarkClear(c)
				rg.CheckAllCaps(rg.Crosses[y][x])
			}
		}
	}
	return &rg
}

func LoadFile(fn string) ([]string, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0)
	for _, txt := range strings.Split(string(data), "\n") {
		lines = append(lines, strings.Trim(txt, "\r\n"))
		fmt.Printf(">%s<\n", lines[len(lines)-1])
	}
	for len(lines) > 0 && len(lines[0]) == 0 {
		lines = lines[1:]
	}
	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-2]
	}
	return lines, nil
}

func main() {
	inp, err := LoadFile("range1.txt")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	b := RangeBoardFromLines(inp)

	fmt.Printf("%s\n", b.StringVerbose())
	b.UpdateWingRanges()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.ExpandWingMinimums()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.UpdateWingRanges()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.ExpandWingMinimums()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.UpdateWingRanges()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.ExpandWingMinimums()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.UpdateWingRanges()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.ExpandWingMinimums()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.UpdateWingRanges()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.ExpandWingMinimums()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.UpdateWingRanges()
	fmt.Printf("*********************************************************\n%s\n", b.String())
	b.ExpandWingMinimums()
	fmt.Printf("*********************************************************\n%s\n", b.StringVerbose())
}
