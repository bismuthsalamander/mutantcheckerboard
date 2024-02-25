package main

import (
	"fmt"
	"math/big"
	"os"
	"strings"
)

func (b *RangeBoard) CrossAt(c Coord) *Cross {
	return b.Crosses[c.Y][c.X]
}

func (b *RangeBoard) IsSolved() (bool, error) {
	// Are there any remaining unknown cells?
	res, coord := b.IsComplete()
	if !res {
		return false, fmt.Errorf("cell %s is unknown", coord)
	}

	// Do all the crosses have the right size?
	for _, cross := range b.AllCrosses {
		ct := 1
		for _, dir := range DIRECTIONS {
			coord := cross.Root.Plus(dir)
			for b.IsClear(coord) {
				ct++
				coord = coord.Plus(dir)
			}
		}
		if ct != cross.Size {
			return false, fmt.Errorf("cross at %s needs %d, but has %d", cross.Root, cross.Size, ct)
		}
	}

	// Are all clear cells contiguous?
	reached := NewCoordSet()
	var start Coord
	b.EachCell(func(cd Coord, v Cell) bool {
		if v == CLEAR {
			start = cd
			return true
		}
		return false
	})
	frontier := make([]Coord, 0, b.W*b.H)
	frontier = append(frontier, start)
	for len(frontier) > 0 {
		touch := frontier[0]
		reached.Add(touch)
		frontier = frontier[1:]
		b.EachNeighbor(touch, func(c Coord, v Cell) bool {
			if !reached.Has(c) {
				frontier = append(frontier, c)
				reached.Add(c)
			}
			return false
		})
	}
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		if !reached.Has(c) && b.Get(c) == CLEAR {
			return false, fmt.Errorf("cannot reach clear cell %s from %s", c, start)
		}
	}
	return true, nil
}

func (b *RangeBoard) Mark(c Coord, v Cell) (bool, error) {
	res, err := b.Set(c, v)
	if !res {
		return res, err
	}
	b.SetDirty()
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
	Min      int
	Max      int
	IsCapped bool
}

type Cross struct {
	Root     Coord
	Size     int
	Wings    map[Delta]*Wing
	IsCapped bool
}

func (c *Cross) String() string {
	return fmt.Sprintf("%c", NumToCh(c.Size))
}

func (c *Cross) StringVerbose() string {
	out := fmt.Sprintf("Cross at %s sz %d\n", c.Root, c.Size)
	for dir, wing := range c.Wings {
		out += fmt.Sprintf("\tWing %s [%d-%d] capped %v\n", dir, wing.Min, wing.Max, wing.IsCapped)
	}
	return out[:len(out)-1]
}

type RangeBoard struct {
	RectBoard
	Crosses    [][]*Cross
	AllCrosses []*Cross
}

// Generates an empty 2d slice of pointers to Cross structs.
func MakeCrosses(w, h int) [][]*Cross {
	c := make([][]*Cross, 0, h)
	for i := 0; i < h; i++ {
		c = append(c, make([]*Cross, w))
	}
	return c
}

// Generates default wings for an island rooted at c with a size crossSize. Initial values are
// chosen so that wings cannot run off the board OR exceed the size of the entire cross.
func (b *RangeBoard) MakeWings(c Coord, crossSize int) map[Delta]*Wing {
	maxsz := crossSize - 1
	mp := make(map[Delta]*Wing)
	mp[LEFT] = &Wing{LEFT, 0, min(maxsz, c.X), false}
	mp[RIGHT] = &Wing{RIGHT, 0, min(maxsz, b.W-(c.X+1)), false}
	mp[UP] = &Wing{UP, 0, min(maxsz, c.Y), false}
	mp[DOWN] = &Wing{DOWN, 0, min(maxsz, b.H-(c.Y+1)), false}
	return mp
}

func (b *RangeBoard) String() string {
	out := "+" + strings.Repeat("-", b.W) + "+\n"
	for y, row := range b.Grid {
		out += "|"
		for x := range row {
			if b.Crosses[y][x] != nil {
				out += b.Crosses[y][x].String()
			} else {
				out += b.Get(Coord{x, y}).String()
			}
		}
		out += "|"
		if y != b.H-1 {
			out += "\n"
		}
	}
	out += "\n+" + strings.Repeat("-", b.W) + "+"
	return out
}

func (b *RangeBoard) StringVerbose() string {
	out := b.String()
	out += "\n\nCrosses:\n"
	for _, cross := range b.AllCrosses {
		if cross == nil {
			continue
		}
		out += fmt.Sprintf("%s\n", cross.StringVerbose())
	}
	return out
}

func (b *RangeBoard) PostMark(c Coord, v Cell) {
	// Clear adjacent cells to paint
	if v == PAINTED {
		b.EachNeighbor(c, func(n Coord, v Cell) bool {
			b.MarkClear(n)
			return false
		})
	}
	// We skip updates of adjacent crosses if we haven't finished initializing the board; else,
	// the wing range updates would be inaccurate
	if !b.InitDone() {
		return
	}
	//TODO: alternatively, cascade a "dirty" mark to affected crosses, then update the dirty ones?
	//helps efficiency of batch updates?
	for _, dir := range DIRECTIONS {
		coord := c.Plus(dir)
		for b.IsValid(coord) && !b.IsPainted(coord) {
			b.UpdateWingRange(b.CrossAt(coord), dir.Reverse())
			coord = coord.Plus(dir)
		}
	}
}

func (b *RangeBoard) CheckAllWingCaps(c *Cross) {
	for _, w := range c.Wings {
		if w.Min == w.Max && !w.IsCapped {
			b.FinishWing(c, w)
		}
	}
}

// Mark this wing as capped, and mark the cross as capped if each of its wings is capped.
func (c *Cross) MarkWingCapped(w *Wing) {
	w.IsCapped = true
	for _, wg := range c.Wings {
		if !wg.IsCapped {
			return
		}
	}
	c.IsCapped = true
}

// If the wing's min and max are wider than the arguments, tighten the wing's range.
func (b *RangeBoard) LimitWing(w *Wing, min, max int) {
	if w.Min < min {
		w.Min = min
		b.SetDirty()
	}
	if w.Max > max {
		w.Max = max
		b.SetDirty()
	}
}

// This function completes each wing of the cross, using each wing's current Min as its Max size.
func (b *RangeBoard) FinishCross(cross *Cross) {
	for _, wing := range cross.Wings {
		wing.Max = wing.Min
		b.FinishWing(cross, wing)
	}
}

// Run this function when we know the wing must have size exactly equal to its Min. FinishWing will
// fill in the clear cells and the painted "cap."
func (b *RangeBoard) FinishWing(cross *Cross, w *Wing) {
	coord := cross.Root
	for i := 1; i <= w.Min; i++ {
		coord = coord.Plus(w.Dir)
		if !b.IsValid(coord) {
			//TODO: report an error cleanly
			fmt.Printf("Error expanding wing %v at cell %d\n", *w, i)
			return
		}
		b.MarkClear(coord)
	}
	cap := cross.Root.Plus(w.Dir.Times(w.Min + 1))
	if b.IsValid(cap) {
		b.MarkPainted(cap)
	}
	cross.MarkWingCapped(w)
}

func (b *RangeBoard) UpdateWingRange(cross *Cross, dir Delta) {
	if cross == nil {
		return
	}
	wing := cross.Wings[dir]
	if wing.IsCapped {
		return
	}
	// Calculate range of possible sizes of this wing based on other wings' ranges
	myWingMax := cross.Size - 1
	myWingMin := cross.Size - 1
	for od, ow := range cross.Wings {
		if od == dir {
			continue
		}
		myWingMax -= ow.Min
		myWingMin -= ow.Max
	}
	if myWingMax < 0 || myWingMin > myWingMax {
		panic(fmt.Sprintf("cross %s (%d) wants [%d,%d] in dir %s?", cross.Root, cross.Size, myWingMin, myWingMax, dir))
	}
	b.LimitWing(wing, myWingMin, myWingMax)
	if wing.Min < 0 || wing.Max < 0 {
		panic("something's neg")
	}

	// Update min and max to match reality (i.e., if there's a painted cell along dir, decrease Max
	// so that the wing can't extend past that painted cell; if there is a contiguous block of
	// clear cells larger than Min, increate Min accordingly).
	coord := cross.Root.Plus(dir)
	wingsz := 1

	// Stop increasing Min once we've seen the first unknown cell. Track with the allClear flag.
	allClear := true
	myWingMin = wing.Min
	myWingMax = wing.Max
	for b.IsValid(coord) && wingsz <= wing.Max {
		if b.IsClear(coord) {
			if allClear {
				myWingMin = wingsz
			}
		} else if b.IsPainted(coord) {
			myWingMax = wingsz - 1
			break
		} else if b.IsUnknown(coord) {
			if wingsz <= wing.Min && allClear {
				b.MarkClear(coord)
			}
			allClear = false
		}
		wingsz++
		coord = coord.Plus(dir)
	}
	b.LimitWing(wing, myWingMin, myWingMax)
	if wing.Min < 0 || wing.Max < 0 {
		panic("after uwr something's neg")
	}
	if wing.Min == wing.Max && !wing.IsCapped {
		b.FinishWing(cross, wing)
	}
}

func (b *RangeBoard) UpdateWingRanges() {
	for _, cross := range b.AllCrosses {
		if cross.IsCapped {
			continue
		}
		for _, dir := range DIRECTIONS {
			b.UpdateWingRange(cross, dir)
		}

		// If our wing minimums are enough to fill the cross, the corss is done.
		sz := 1
		for _, wing := range cross.Wings {
			sz += wing.Min
		}
		if sz == cross.Size {
			b.FinishCross(cross)
		}
	}
}

func (b *RangeBoard) RestrictWingsForExtending() {
	for _, c := range b.AllCrosses {
		for dir, w := range c.Wings {
			if w.IsCapped {
				continue
			}
			b.RestrictWingForExtending(c, dir)
		}
	}
}

func (b *RangeBoard) RestrictWingForExtending(c *Cross, dir Delta) {
	// reduce max because max would extend
	for {
		nextCell := c.Root.Plus(dir.Times(c.Wings[dir].Max + 1))
		if !b.IsClear(nextCell) {
			break
		}
		c.Wings[dir].Max--
		b.SetDirty()
	}
	// increase min because min would extend
	for {
		nextCell := c.Root.Plus(dir.Times(c.Wings[dir].Min + 1))
		if !b.IsClear(nextCell) {
			break
		}
		c.Wings[dir].Min++
		b.SetDirty()
	}
}

// TODO: I think we can unify some of these range checks?
// If extending cross C's wing would cause it to merge with cross D, and cross D can't extend that
// far, we need to reduce C's wing's Max so that it can't merge with D anymore.
func (b *RangeBoard) CheckCrossMerging() {
	for _, cross := range b.AllCrosses {
		if cross.IsCapped {
			continue
		}
	oneWing:
		for dir, w := range cross.Wings {
			if w.IsCapped {
				continue
			}
			oppWing := cross.Wings[dir.Reverse()]
			oppWingEnd := cross.Root.Plus(dir.Reverse().Times(oppWing.Min))
			for trywinglen := w.Min; trywinglen <= w.Max; trywinglen++ {
				neighbors := make([]*Cross, 0)
				tryneighbor := cross.Root.Plus(dir.Times(trywinglen + 1))
				for b.IsValid(tryneighbor) && b.IsClear(tryneighbor) {
					if b.CrossAt(tryneighbor) != nil {
						neighbors = append(neighbors, b.CrossAt(tryneighbor))
					}
					tryneighbor = tryneighbor.Plus(dir)
				}
				for _, nc := range neighbors {
					ncWing := nc.Wings[dir.Reverse()]
					dist := nc.Root.MHDist(oppWingEnd)
					if ncWing.Max < dist {
						w.Max = trywinglen - 1
						b.SetDirty()
						break oneWing
					}
				}
			}
		}
	}
}

// Looks for clear cells with one liberty and marks the liberty as clear. Limited case of
// ClearAllDominators below.
func (b *RangeBoard) ClearMiniDominators() {
	b.EachCell(func(c Coord, v Cell) bool {
		if v != CLEAR {
			return false
		}
		var lib Coord
		liberties := 0
		b.EachNeighbor(c, func(n Coord, nv Cell) bool {
			if nv != PAINTED {
				liberties++
				lib = n
				if liberties > 1 {
					return true
				}
			}
			return false
		})
		if liberties == 1 {
			b.MarkClear(lib)
		}
		return false
	})
}

// In graph theory, a DAG node D is a dominator of sink S w/r/t an origin O if every path from O
// to S must include D. Consider a graph whose nodes are non-painted (i.e., painted and unknown)
// cells and whose edges are orthogonal adjacency relationships. If D dominates ANY other node
// for ANY origin, then D is a "choke point" that must be clear in order for the board to satisfy
// the constraint that the clear cells must form a contiguous group. This function implements the
// first algorithm shown here: https://en.wikipedia.org/wiki/Dominator_(graph_theory)#Algorithms
// Note that we have to call this function twice. By definition, a node dominates itself, so by
// calling the function with two different source nodes, we are sure to find every dominator.
func (b *RangeBoard) ClearAllDominators(start Coord) {
	doms := make([][]*Set[Coord], 0)
	for y := 0; y < b.H; y++ {
		doms = append(doms, make([]*Set[Coord], b.H))
		for x := 0; x < b.W; x++ {
			if !b.IsPainted(Coord{x, y}) {
				doms[y][x] = NewCoordSet()
			}
		}
	}
	doms[start.Y][start.X].Add(start)
	def := NewCoordSet()
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		if !b.IsPainted(c) {
			def.Add(c)
		}
	}
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		if !b.IsPainted(c) && c != start {
			doms[c.Y][c.X].AddAll(def)
		}
	}
	changed := true
	for changed {
		changed = false
		for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
			if b.IsPainted(c) || c == start {
				continue
			}
			var newDoms *Set[Coord]
			b.EachNeighbor(c, func(n Coord, nv Cell) bool {
				if nv != PAINTED {
					if newDoms == nil {
						newDoms = doms[n.Y][n.X].Copy()
					} else {
						newDoms.IntersectWith(doms[n.Y][n.X])
					}
				}
				return false
			})
			newDoms.Add(c)
			if newDoms.Size() != doms[c.Y][c.X].Size() {
				doms[c.Y][c.X] = newDoms
				changed = true
			}
		}
	}
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		if doms[c.Y][c.X] != nil && doms[c.Y][c.X].Size() >= 3 {
			for k := range doms[c.Y][c.X].M {
				if k != start && k != c {
					b.MarkClear(k)
				}
			}
		}
	}
}

// Detects crosses that are connected along one axis and cross-enforces limitations
func (b *RangeBoard) UpdateSharedRanges() {
	coords := NewCoordSet()
	for x := 0; x < b.W; x++ {
		coords.Clear()
		for y := 0; y < b.H; y++ {
			c := Coord{x, y}
			if !b.IsClear(c) {
				b.ShareRangesVertical(coords)
				coords.Clear()
			} else if b.CrossAt(c) != nil {
				coords.Add(c)
			}
		}
		b.ShareRangesVertical(coords)
	}
	coords.Clear()
	for y := 0; y < b.H; y++ {
		coords.Clear()
		for x := 0; x < b.W; x++ {
			c := Coord{x, y}
			if !b.IsClear(c) {
				b.ShareRangesHorizontal(coords)
				coords.Clear()
			} else if b.CrossAt(c) != nil {
				coords.Add(c)
			}
		}
		b.ShareRangesHorizontal(coords)
	}
}

func (b *RangeBoard) ShareRangesVertical(s *Set[Coord]) {
	b.ShareRanges(s, LEFT, RIGHT, UP, DOWN)
}

func (b *RangeBoard) ShareRangesHorizontal(s *Set[Coord]) {
	b.ShareRanges(s, UP, DOWN, LEFT, RIGHT)
}

/*
two crosses share an axis
no cross can:
  - expand past the lowest axis max in the group
  - have a minimum below the highest min in the group

find the lowest shared-axis max
find the highest shared-axis min

each cross must:
  - have its own cross-axis required between its two cross-wings
*/

func (b *RangeBoard) ApplyAxisRange(c *Cross, axisMin, axisMax int, dir1, dir2 Delta) {
	min2 := axisMin - c.Wings[dir1].Max
	min1 := axisMin - c.Wings[dir2].Max
	max2 := axisMax - c.Wings[dir1].Min
	max1 := axisMax - c.Wings[dir2].Min
	b.LimitWing(c.Wings[dir1], min1, max1)
	b.LimitWing(c.Wings[dir2], min2, max2)
}

// TODO: change this to track axis mins and maxes on cross struct
func (b *RangeBoard) ShareRanges(s *Set[Coord], cross1 Delta, cross2 Delta, shared1 Delta, shared2 Delta) {
	if s.Size() < 2 {
		return
	}
	first := true
	axisSharedMin := 0
	axisSharedMax := 0
	for k := range s.M {
		c := b.CrossAt(k)
		sharedMin := (c.Size - 1) - (c.Wings[cross1].Max + c.Wings[cross2].Max)
		sharedMax := (c.Size - 1) - (c.Wings[cross1].Min + c.Wings[cross2].Min)
		sharedBoardSize := b.H
		if shared1 == LEFT || shared2 == LEFT {
			sharedBoardSize = b.W
		}
		if sharedMax > sharedBoardSize {
			sharedMax = sharedBoardSize
		}
		// fmt.Printf("Cross %s min along shared is %d; max along shared is %d\n", c, sharedMin, sharedMax)
		if first || sharedMin > axisSharedMin {
			axisSharedMin = sharedMin
		}
		if first || sharedMax < axisSharedMax {
			axisSharedMax = sharedMax
		}
		first = false
	}
	// fmt.Printf("High shared min is %d; low shared max is %d\n", axisSharedMin, axisSharedMax)
	for k := range s.M {
		c := b.CrossAt(k)
		b.ApplyAxisRange(c, axisSharedMin, axisSharedMax, shared1, shared2)

		crossAxisMin := c.Size - (1 + c.Wings[shared1].Max + c.Wings[shared2].Max)
		crossAxisMax := c.Size - (1 + c.Wings[shared1].Min + c.Wings[shared2].Min)
		b.ApplyAxisRange(c, crossAxisMin, crossAxisMax, cross1, cross2)
	}
}

func (b *RangeBoard) Solve() {
	b.SetDirty()
	for b.IsDirty() {
		fmt.Printf("Possibilities: %d\n", b.NumPossibilities())
		b.ClearDirty()
		b.UpdateWingRanges()
		b.RestrictWingsForExtending()
		b.CheckCrossMerging()
		b.UpdateSharedRanges()
		b.ClearMiniDominators()
		if b.IsDirty() {
			continue
		}
		done := 0
		for c := b.TopLeft(); b.IsValid(c) && done < 2; c = b.Next(c) {
			if b.IsClear(c) {
				b.ClearAllDominators(c)
				done++
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
				c := Coord{x, y}
				rg.Crosses[y][x] = &Cross{
					Root:     c,
					Size:     val,
					Wings:    rg.MakeWings(c, val),
					IsCapped: false,
				}
				rg.AllCrosses = append(rg.AllCrosses, rg.Crosses[y][x])
				rg.MarkClear(c)
				rg.CheckAllWingCaps(rg.Crosses[y][x])
			}
		}
	}
	rg.Inited = true
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

func shared() {
	inp, err := LoadFile("adj.txt")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	b := RangeBoardFromLines(inp)

	fmt.Printf("%s\n", b.StringVerbose())
	b.UpdateWingRanges()
	b.UpdateSharedRanges()
	fmt.Printf("%s\n", b.StringVerbose())
	b.UpdateWingRanges()
	fmt.Printf("%s\n", b.StringVerbose())
}

func main() {
	// shared()
	// return
	// return
	solveit()
	return
	inp, err := LoadFile("blank.txt")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	b := RangeBoardFromLines(inp)

	fmt.Printf("%s\n", b.StringVerbose())
	b.MarkPainted(Coord{1, 0})
	b.MarkPainted(Coord{0, 3})
	b.MarkPainted(Coord{1, 4})
	b.MarkPainted(Coord{1, 6})
	b.MarkPainted(Coord{0, 7})
	fmt.Printf("%s\nNOW CLAERING\n", b.String())
	b.ClearAllDominators(Coord{5, 8})
	fmt.Printf("%s\n", b.String())
}

func (c *Cross) NumPossibilities() uint64 {
	ct := uint64(0)
	for left := c.Wings[LEFT].Min; left <= c.Wings[LEFT].Max; left++ {
		for up := c.Wings[UP].Min; up <= c.Wings[UP].Max; up++ {
			for right := c.Wings[RIGHT].Min; right <= c.Wings[RIGHT].Max; right++ {
				down := c.Size - (left + up + right + 1)
				if down >= c.Wings[DOWN].Min && down <= c.Wings[DOWN].Max {
					ct++
				}
			}
		}
	}
	return ct
}

func (b *RangeBoard) NumPossibilities() *big.Int {
	tot := big.NewInt(1)
	for _, c := range b.AllCrosses {
		if !c.IsCapped {
			tot.Mul(tot, big.NewInt(int64(c.NumPossibilities())))
		}
	}
	return tot
}

func solveit() {
	inp, err := LoadFile("range5.txt")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	b := RangeBoardFromLines(inp)
	// tot := 1
	// for _, c := range b.AllCrosses {
	// 	// fmt.Printf("%s (%d) %d\n", c.Root, c.Size, c.NumPossibilities())
	// 	tot *= c.NumPossibilities()
	// }
	// fmt.Printf("TOTAL: %d\n", tot)
	// return
	fmt.Printf("%s\n", b.StringVerbose())
	b.Solve()
	fmt.Printf("*********************************************************\n%s\n", b.StringVerbose())
	c, d := b.IsSolved()
	fmt.Printf("Solved: %v (%v)\n", c, d)
	return
}
