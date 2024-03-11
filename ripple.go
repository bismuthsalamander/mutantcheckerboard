package main

import "fmt"

type RippleBoard struct {
	RectNumBoard
}

func RippleBoardFromLines(input []string) (*RippleBoard, error) {
	if len(input)%2 != 0 || len(input) == 0 {
		return nil, fmt.Errorf("must have a region grid and a number grid; have %d lines", len(input))
	}
	h := len(input) / 2

	allRegions, regionGrid := LinesToRegionGrid(input[:h])
	numgrid, err := LinesToIntGrid(input[:h])
	if err != nil {
		return nil, err
	}
	rect := RectNumBoardFromNums(numgrid)
	b := RippleBoard{
		RectNumBoard: *rect,
	}
	b.AllRegions = allRegions
	b.RegionGrid = regionGrid
	b.Allowed = make([][]*Set[int], b.H)
	for y := 0; y < b.H; y++ {
		b.Allowed[y] = make([]*Set[int], b.W)
		for x := 0; x < b.W; x++ {
			sz := max(b.W, b.H)
			for _, r := range b.RegionGrid[y][x] {
				l := len(*r)
				sz = min(l, sz)
			}
			b.Allowed[y][x] = NewNumSet(sz)
		}
	}
	b.Inited = true
	b.EachCell(func(c Coord, v int) bool {
		b.PostMark(c, v)
		return false
	})
	return &b, nil
}

func (b *RippleBoard) PostMark(c Coord, v int) (bool, error) {
	if v == UNKNOWN {
		return false, nil
	}
	changed := false
	for _, region := range b.RegionGrid[c.Y][c.X] {
		for _, neighbor := range *region {
			if neighbor == c || !b.Allowed[neighbor.Y][neighbor.X].Has(v) {
				continue
			}
			b.Disallow(neighbor, v)
			changed = true
		}
	}
	for _, dir := range DIRECTIONS {
		for i := 1; i <= v; i++ {
			c := c.Plus(dir.Times(i))
			if b.Disallow(c, v) {
				changed = true
			}
		}
	}

	return changed, nil
}

func (b *RippleBoard) String() string {
	out := ""
	for ri := 0; ri < b.H; ri++ {
		for ci := 0; ci < b.W; ci++ {
			out += b.CharAt(Coord{ci, ri})
		}
		out += "\n"
	}
	return out
}

// Solved returns true iff all observers are satisfied and all cells are
// filled. As of now, it does not confirm that the sudoku rule is satisfied.
func (b *RippleBoard) IsSolved() (bool, error) {
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		if b.Get(c) == UNKNOWN {
			return false, fmt.Errorf("cell %s is unknown", c)
		}
	}
	for _, r := range b.AllRegions {
		if !b.IsRegionSolved(*r) {
			return false, fmt.Errorf("region %v unsatisfied", *r)
		}
	}
	return true, nil
}

// AutoSolve runs all implemented solving heuristics until the puzzle is solved
// or we run out of improvements. Missing heuristics include the opposite of
// naked sets (i.e., cells X and Y are the only possible locations for numbers
// N and M, so X and Y can't have any other numbers) and pairwise permutation
// consistency between rows or columns.
func (b *RippleBoard) Solve() (bool, error) {
	changed := true
	for _, err := b.IsSolved(); changed && err != nil; {
		changed = false
		if b.MarkMandatory() {
			changed = true
		}
		if b.TrimAllFoundGroups() {
			changed = true
		}
		if b.TrimAllNakedSets() {
			changed = true
		}
	}
	return b.IsSolved()
}
