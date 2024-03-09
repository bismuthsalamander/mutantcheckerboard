package main

import (
	"fmt"
	"os"
)

type RectNumBoard struct {
	RectBoard
	Grid       [][]int
	AllRegions []*[]Coord
	RegionGrid [][][]*[]Coord
	Allowed    [][]*Set[int]
	Guess      [][]int
}

func RectNumBoardFromNums(input [][]int) *RectNumBoard {
	w := len(input[0])
	h := len(input)
	board := &RectNumBoard{
		RectBoard: RectBoard{
			W: w,
			H: h,
		},
		Grid:       MakeNumGrid(w, h),
		AllRegions: make([]*[]Coord, 0),
		RegionGrid: MakeRegionGrid(w, h),
		Guess:      MakeNumGrid(w, h),
	}
	board.AddRowColRegions()
	for r, row := range input {
		copy(board.Grid[r], row)
	}
	return board
}

func NewRegion() []Coord {
	return make([]Coord, 0)
}

func (b *RectNumBoard) AddRowColRegions() {
	for ri := 0; ri < b.W; ri++ {
		rowRegion := NewRegion()
		c := Coord{X: 0, Y: ri}
		for ; b.IsValid(c); c.X++ {
			rowRegion = append(rowRegion, c)
		}
		b.AddRegion(rowRegion)
	}
	for ci := 0; ci < b.H; ci++ {
		colRegion := NewRegion()
		c := Coord{X: ci, Y: 0}
		for ; b.IsValid(c); c.Y++ {
			colRegion = append(colRegion, c)
		}
		b.AddRegion(colRegion)
	}
}

func (b *RectNumBoard) AddRegion(r []Coord) {
	b.AllRegions = append(b.AllRegions, &r)
	for _, c := range r {
		b.RegionGrid[c.Y][c.X] = append(b.RegionGrid[c.Y][c.X], &r)
	}
}

func (b *RectNumBoard) AllowedCount(c Coord) int {
	return b.Allowed[c.Y][c.X].Size()
}

func (b *RectNumBoard) Get(c Coord) int {
	if b.Guess[c.Y][c.X] != UNKNOWN {
		return b.Guess[c.Y][c.X]
	}
	return b.Grid[c.Y][c.X]
}

func (b *RectNumBoard) ClearGuess() {
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		b.Guess[c.Y][c.X] = UNKNOWN
	}
}

func (b *RectNumBoard) IsFilled(c Coord) bool {
	return b.IsValid(c) && b.Get(c) != UNKNOWN
}
func (b *RectNumBoard) IsUnknown(c Coord) bool {
	return b.IsValid(c) && b.Get(c) == UNKNOWN
}
func (b *RectNumBoard) IsComplete() (bool, Coord) {
	for y, row := range b.Grid {
		for x, cell := range row {
			if cell == UNKNOWN {
				return false, Coord{x, y}
			}
		}
	}
	return true, Coord{}
}
func (b *RectNumBoard) IsSolved() bool {
	return false
}

func (b *RectNumBoard) Disallow(c Coord, v int) bool {
	ok := b.Allowed[c.Y][c.X].Has(v)
	b.Allowed[c.Y][c.X].Del(v)
	return ok
}

func (b *RectNumBoard) MaxRegionSize() int {
	n := 0
	for _, r := range b.AllRegions {
		n = max(len(*r), n)
	}
	return n
}

func (b *RectNumBoard) IsAllowed(c Coord, v int) bool {
	return b.Allowed[c.Y][c.X].Has(v)
}

func (b *RectNumBoard) AllowsExactly(c Coord, nums []int) bool {
	return b.Allowed[c.Y][c.X].EqualsSlice(nums)
}

func (b *RectNumBoard) Mark(c Coord, v int) (bool, error) {
	res, err := b.Set(c, v)
	if !b.Inited {
		return res, err
	}
	for _, region := range b.RegionGrid[c.Y][c.X] {
		for _, neighbor := range *region {
			if neighbor != c {
				b.Disallow(neighbor, v)
			}
		}
	}
	return res, err
}

func (b *RectNumBoard) Set(c Coord, v int) (bool, error) {
	if !b.IsValid(c) {
		return false, fmt.Errorf("coordinate (%d,%d) not valid on board of size (%d,%d)", c.X, c.Y, b.W, b.H)
	}
	if b.IsFilled(c) {
		if b.Get(c) == v {
			return false, nil
		}
		return false, fmt.Errorf("coordinate (%d,%d) is already set to %d; cannot set it to %d", c.X, c.Y, b.Get(c), v)
	}
	b.Grid[c.Y][c.X] = v
	return true, nil
}

func (b *RectNumBoard) EachCell(cb func(c Coord, v int) bool) {
	for y, row := range b.Grid {
		for x, cell := range row {
			if cb(Coord{x, y}, cell) {
				return
			}
		}
	}
}

func (b *RectNumBoard) EachNeighbor(start Coord, cb func(c Coord, v int) bool) {
	for _, dir := range DIRECTIONS {
		c := start.Plus(dir)
		if b.IsValid(c) {
			cb(c, b.Get(c))
		}
	}
}

func (b *RectNumBoard) DisallowAll(c Coord, s Set[int]) bool {
	changed := false
	for k, _ := range s.M {
		if b.Allowed[c.Y][c.X].Del(k) {
			changed = true
		}
	}
	return changed
}

// CheckRegionFoundGroup returns true iff row the region contains a found group for
// the numbers specified in numbers.
func (b *RectNumBoard) CheckRegionFoundGroup(numbers []int, r []Coord) bool {
	numberCells := make([]*Set[Coord], len(numbers))
	for i := range numbers {
		numberCells[i] = NewCoordSet()
	}
	for _, c := range r {
		for nidx, num := range numbers {
			if b.IsAllowed(c, num) {
				numberCells[nidx].Add(c)
			}
		}
	}
	if numberCells[0].Size() != len(numbers) {
		return false
	}
	for i := 1; i < len(numbers); i++ {
		if !numberCells[i].Equals(numberCells[0]) {
			return false
		}
	}
	fmt.Printf("Have a found group of numbers %s (%s)\n", numbers, numberCells[0])
	for _, c := range r {
		fmt.Printf("\tCell %s allows %s\n", c, b.Allowed[c.Y][c.X])
	}
	return true
}

// TrimFoundGroups looks at each row and column for found groups of size n and
// makes the appropriate changes to b.Allowed if any are found. Returns true
// iff at least one change was made. A found group (I haven't looked up the
// correct name for this heuristic) occurs when, e.g., the numbers 2 and 3 each
// have the same exact two possible homes in a given line. Since 2 and 3 must
// go in those two cells, all other numbers can be removed from their allowed
// lists. TODO: update with the correct term for "found groups!"
func (b *RectNumBoard) TrimFoundGroups(n int) bool {
	changed := false
	numbers := Permute(1, b.MaxRegionSize(), n)
	for _, nums := range numbers {
		for _, r := range b.AllRegions {
			if n >= len(*r) {
				continue
			}
			if !b.CheckRegionFoundGroup(nums, *r) {
				continue
			}
			for _, c := range *r {
				if b.IsAllowed(c, nums[0]) {
					if b.DisallowOthers(c, nums) {
						changed = true
					}
				}
			}
		}
	}
	return changed
}

func (b *RectNumBoard) DisallowOthers(c Coord, nums []int) bool {
	changed := false
	for n := 1; n <= b.MaxRegionSize(); n++ {
		if !SliceContains(nums, n) && b.Disallow(c, n) {
			changed = true
		}
	}
	return changed
}

// CheckRowNakedSet returns true iff row rowIndex contains a naked set at the
// indices specified in indices.
func (b *RectNumBoard) CheckRegionNakedSet(indices []int, region []Coord) bool {
	if len(indices) == 0 {
		return false
	}
	if len(indices) > len(region) {
		return false
	}
	firstCoord := region[indices[0]]
	if len(indices) != b.AllowedCount(firstCoord) {
		return false
	}
	for _, idx := range indices[1:] {
		if idx >= len(region) {
			return false
		}
		if b.Get(region[idx]) != UNKNOWN {
			return false
		}
		if !b.Allowed[region[idx].Y][region[idx].X].Equals(b.Allowed[firstCoord.Y][firstCoord.X]) {
			return false
		}
	}
	return true
}

// TrimNakedSets looks at each row and column for naked sets of size n and
// makes the appropriate changes to b.Allowed if any are found. Returns true
// iff at least one change was made. A naked set (known more commonly as a
// naked pair or naked triple) occurs when, e.g., the allowed lists for cells
// A and B are [1, 2]. It allows us to eliminate 1 and 2 from the allowed lists
// of other cells in the same line.
func (b *RectNumBoard) TrimNakedSets(n int) bool {
	result := false
	for _, region := range b.AllRegions {
		if n >= len(*region) {
			continue
		}
		regionSubsets := Permute(0, len(*region)-1, n)
		for _, idxs := range regionSubsets {
			if b.CheckRegionNakedSet(idxs, *region) {
				tmp := (*region)[idxs[0]]
				numsToDisallow := b.Allowed[tmp.Y][tmp.X]
				for idx, c := range *region {
					if SliceContains(idxs, idx) {
						continue
					}
					if b.DisallowAll(c, *numsToDisallow) {
						result = true
					}
				}
			}
		}

	}
	return result
}

func (b *RectNumBoard) TrimAllNakedSets() bool {
	changed := false
	for n := 2; n < b.MaxRegionSize(); n++ {
		if b.TrimNakedSets(n) {
			changed = true
		}
	}
	return changed
}

func (b *RectNumBoard) TrimAllFoundGroups() bool {
	changed := false
	for n := 2; n < b.MaxRegionSize(); n++ {
		if b.TrimNakedSets(n) {
			changed = true
		}
	}
	return changed
}

func TestNakedSets() {
	inp, err := LoadIntFile("tower-sets.txt")
	if err != nil {
		fmt.Printf("error loading file: %s\n", err)
		os.Exit(-1)
	}
	b, err := TowerBoardFromLines(inp)
	fmt.Printf("Board:\n\n%s\n\nerr: %s\n", b, err)
	b.Disallow(Coord{0, 0}, 1)
	b.Disallow(Coord{0, 0}, 2)
	b.Disallow(Coord{1, 0}, 1)
	b.Disallow(Coord{1, 0}, 2)
	b.Disallow(Coord{0, 1}, 1)
	b.Disallow(Coord{0, 1}, 2)
	for ci := 0; ci < b.W; ci++ {
		fmt.Printf("Allowed at (0,%d): %v\n", ci, b.Allowed[0][ci])
	}
	b.TrimNakedSets(2)
	for ci := 0; ci < b.W; ci++ {
		fmt.Printf("Allowed at (0,%d): %v\n", ci, b.Allowed[0][ci])
	}
	for ri := 0; ri < b.H; ri++ {
		fmt.Printf("Allowed at (%d,0): %v\n", ri, b.Allowed[ri][0])
	}
}

func TestFoundGroups() {
	inp, err := LoadIntFile("tower-sets-5.txt")
	if err != nil {
		fmt.Printf("error loading file: %s\n", err)
		os.Exit(-1)
	}
	b, err := TowerBoardFromLines(inp)
	fmt.Printf("Board:\n\n%s\n\nerr: %s\n", b, err)
	b.Disallow(Coord{0, 0}, 1)
	b.Disallow(Coord{0, 0}, 2)
	b.Disallow(Coord{0, 0}, 3)
	b.Disallow(Coord{1, 0}, 1)
	b.Disallow(Coord{1, 0}, 2)
	b.Disallow(Coord{1, 0}, 3)
	for ci := 0; ci < b.W; ci++ {
		fmt.Printf("Allowed at (0,%d): %v\n", ci, b.Allowed[0][ci])
	}
	// b.TrimFoundGroups(2)
	b.TrimFoundGroups(3)
	for ci := 0; ci < b.W; ci++ {
		fmt.Printf("Allowed at (0,%d): %v\n", ci, b.Allowed[0][ci])
	}
	for ri := 0; ri < b.H; ri++ {
		fmt.Printf("Allowed at (%d,0): %v\n", ri, b.Allowed[ri][0])
	}
}
