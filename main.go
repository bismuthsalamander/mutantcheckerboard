package main

import (
	"fmt"
	"os"
	"strings"
)

func (b *RangeBoard) CrossAt(c Coord) *Cross {
	return b.Crosses[c.Y][c.X]
}

func (b *RangeBoard) IsSolved() (bool, error) {
	var err error
	b.EachCell(func(c Coord, v Cell) bool {
		if v == UNKNOWN {
			err = fmt.Errorf("cell %s is unknown", c)
			return true
		}
		return false
	})
	if err != nil {
		return false, err
	}
	for _, cross := range b.AllCrosses {
		ct := 1
		for _, dir := range DIRECTIONS {
			coord := cross.Root.Plus(dir)
			for b.IsValid(coord) && b.IsClear(coord) {
				ct++
				coord = coord.Plus(dir)
			}
		}
		if ct != cross.Size {
			return false, fmt.Errorf("cross at %s needs %d; has %d", cross.Root, cross.Size, ct)
		}
	}
	reached := NewCoordSet()
	var start Coord
	b.EachCell(func(cd Coord, v Cell) bool {
		if v == CLEAR {
			start = cd
			return true
		}
		return false
	})
	frontier := make([]Coord, 0, 1)
	frontier = append(frontier, start)
	for len(frontier) > 0 {
		touch := frontier[0]
		reached.Add(touch)
		frontier = frontier[1:]
		b.EachNeighbor(touch, func(c Coord, v Cell) bool {
			if !reached.Has(c) {
				frontier = append(frontier, c)
			}
			return false
		})
	}
	b.EachCell(func(c Coord, v Cell) bool {
		if !reached.Has(c) && v == CLEAR {
			err = fmt.Errorf("cannot reach clear cell %s from %s", c, start)
			return true
		}
		return false
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *RangeBoard) Mark(c Coord, v Cell) (bool, error) {
	res, err := b.Set(c, v)
	if res == false {
		return res, err
	}
	fmt.Printf("MARKING %s as %s\n", c, v)
	// fmt.Printf("%s\n", b)
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
	out := fmt.Sprintf("Cross at %s sz %d", c.Root, c.Size)
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

func MakeCrosses(w, h int) [][]*Cross {
	c := make([][]*Cross, 0, h)
	for i := 0; i < h; i++ {
		c = append(c, make([]*Cross, w))
	}
	return c
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

func (b *RangeBoard) MakeWings(c Coord) map[Delta]*Wing {
	mp := make(map[Delta]*Wing)
	mp[LEFT] = &Wing{LEFT, 0, c.X, false}
	mp[RIGHT] = &Wing{RIGHT, 0, b.W - (c.X + 1), false}
	mp[UP] = &Wing{UP, 0, c.Y, false}
	mp[DOWN] = &Wing{DOWN, 0, b.H - (c.Y + 1), false}
	return mp
}

func (b *RangeBoard) String() string {
	out := "+" + strings.Repeat("-", b.W) + "+\n"
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
	/**
	If painted:
		mark adj clear
		limit wing maxes along all four axes (THIS CAN STOP AFTER ENCOUNTERING ANOTHER PAINTED CELL)
	If clear:
		increase minimum size along all four axes until we meet a non-clear cell

	SHORTCUT

	If painted:
		mark adj clear
	traverse four axes until we find a painted cell
	for each cross found, update its wing sizes?
	*/
	// Clear adjacent cells to paint
	if v == PAINTED {
		b.EachNeighbor(c, func(n Coord, v Cell) bool {
			b.MarkClear(n)
			return false
		})
	}
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
	// fmt.Printf("Finishing cross at %s\n", cross.Root)
	for _, wing := range cross.Wings {
		wing.Max = wing.Min
		b.FinishWing(cross, wing)
	}
}

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
	otherWingsMax := 1
	otherWingsMin := 1
	for od, ow := range cross.Wings {
		if od == dir {
			continue
		}
		otherWingsMax += ow.Max
		otherWingsMin += ow.Min
	}
	if wing.Min+otherWingsMax < cross.Size && wing.Min < cross.Size-otherWingsMax {
		wing.Min = cross.Size - otherWingsMax
		if wing.Min > wing.Max {
			panic("wat min>max")
		}
		b.SetDirty()
	}
	if wing.Max+otherWingsMin > cross.Size && otherWingsMin < cross.Size && wing.Max > cross.Size-otherWingsMin {
		wing.Max = cross.Size - otherWingsMin
		if wing.Min > wing.Max {
			panic("wat min>max")
		}
		b.SetDirty()
	}
	coord := cross.Root.Plus(dir)
	wingsz := 1
	fmt.Printf("In UWR, Wing at %s-%s started [%d-%d]\n", cross.Root, wing.Dir, wing.Min, wing.Max)
	allClear := true
	for b.IsValid(coord) && wingsz <= wing.Max {
		if b.IsClear(coord) {
			if allClear && wingsz > wing.Min {
				wing.Min = wingsz
				b.SetDirty()
			}
		} else if b.IsPainted(coord) {
			wingsz--
			if wingsz < wing.Max {
				wing.Max = wingsz
				b.SetDirty()
			}
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
		sz := 1
		for _, wing := range cross.Wings {
			sz += wing.Min
		}
		if sz == cross.Size {
			b.FinishCross(cross)
		}
	}
}

func (b *RangeBoard) ExpandWingTo(c *Cross, w *Wing, sz int) {
	if w.Min >= sz {
		return
	}
	fmt.Printf(">>>>>> Expanding wing %s-%s to %d\n", c.Root, w.Dir, sz)
	w.Min = sz
	b.SetDirty()
	for i := 1; i <= sz; i++ {
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
						// fmt.Printf("Board:\n%s\n", b)
						// fmt.Printf("Reducing max of cross at %s to %d because of merge with cross at %s\n", cross.Root, trywinglen-1, nc.Root)
						// fmt.Printf("Expanding cross at %s in dir %s\n", cross.Root, dir)
						// fmt.Printf("At winglen %d, merges with nc at %s\n", trywinglen, nc.Root)
						// fmt.Printf("That would make neighbor wing %s at least %d long; max is %d", dir.Reverse(), dist, ncWing.Max)
						w.Max = trywinglen - 1
						b.SetDirty()
						break oneWing
					}
				}
			}
			// loop N = [Min, Max]
			// let winglen = cross.Wings[dir.Reverse()] + 1 +
			// start at Root + dir * (N+1)
			// continue until non-clear; keep a slice of all crosses encountered
			//
			// for each of these neighbors
			// see if neighbor.Wings[dir.Reverse()].Min
		}
	}
}

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
		if doms[c.Y][c.X] == nil {
			continue
		}
		if doms[c.Y][c.X].Size() >= 3 {
			fmt.Printf("COordinate %s has %d dominators:\n", c, doms[c.Y][c.X].Size())
			for k, _ := range doms[c.Y][c.X].M {
				if k != start && k != c {
					fmt.Printf("\t%s\n", k)
					b.MarkClear(k)
				}
			}
		}
	}
}

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
			if b.IsPainted(c) {
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
	// lengthen dir2
	if c.Wings[dir1].Max+c.Wings[dir2].Min < axisMin {
		newMin := axisMin - c.Wings[dir1].Max
		fmt.Printf("Changing %s's Min along %s to %d\n", c.Root, dir2, newMin)
		c.Wings[dir2].Min = newMin
		b.SetDirty()
	}
	// lengthen dir1
	if c.Wings[dir2].Max+c.Wings[dir1].Min < axisMin {
		newMin := axisMin - c.Wings[dir2].Max
		fmt.Printf("Changing %s's Min along %s to %d\n", c.Root, dir1, newMin)
		c.Wings[dir1].Min = newMin
		b.SetDirty()
	}
	// shorten dir2
	if c.Wings[dir1].Min+c.Wings[dir2].Max > axisMax {
		newMax := axisMax - c.Wings[dir1].Min
		fmt.Printf("Changing %s's Max along %s to %d\n", c.Root, dir2, newMax)
		c.Wings[dir2].Max = newMax
		b.SetDirty()
	}
	// shorten dir1
	if c.Wings[dir2].Min+c.Wings[dir1].Max > axisMax {
		newMax := axisMax - c.Wings[dir2].Min
		fmt.Printf("Changing %s's Max along %s to %d\n", c.Root, dir1, newMax)
		c.Wings[dir1].Max = newMax
		b.SetDirty()
	}
}

// TODO: change this to track axis mins and maxes
func (b *RangeBoard) ShareRanges(s *Set[Coord], cross1 Delta, cross2 Delta, shared1 Delta, shared2 Delta) {
	if s.Size() < 2 {
		return
	}
	fmt.Printf("%s\nShare ranges along %s, %s:\n", b, cross1, cross2)
	for k, _ := range s.M {
		fmt.Printf("\t%s %s\n", k, b.CrossAt(k))
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
		b.ClearDirty()
		b.UpdateWingRanges()
		b.ExpandWingMinimums()
		b.CheckCrossMerging()
		b.UpdateSharedRanges()
		b.ClearMiniDominators()
		if b.IsDirty() {
			continue
		}
		b.ClearAllDominators(Coord{0, 0})
		b.ClearAllDominators(Coord{b.W - 1, b.H - 1})
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
					Root:     c,
					Size:     val,
					Wings:    rg.MakeWings(c),
					IsCapped: false,
				}
				rg.AllCrosses = append(rg.AllCrosses, rg.Crosses[y][x])
				rg.MarkClear(c)
				fmt.Printf("coord (%d,%d)\n", x, y)
				rg.CheckAllCaps(rg.Crosses[y][x])
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
	b.ExpandWingMinimums()
	b.UpdateWingRanges()
	fmt.Printf("%s\n", b.StringVerbose())
}

func main() {
	shared()
	return
	// return
	inp, err := LoadFile("blank.txt")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	b := RangeBoardFromLines(inp)

	fmt.Printf("%s\n", b.StringVerbose())
	b.MarkPainted(MkCoord(1, 0))
	b.MarkPainted(MkCoord(0, 3))
	b.MarkPainted(MkCoord(1, 4))
	b.MarkPainted(MkCoord(1, 6))
	b.MarkPainted(MkCoord(0, 7))
	fmt.Printf("%s\nNOW CLAERING\n", b.String())
	b.ClearAllDominators(Coord{5, 8})
	fmt.Printf("%s\n", b.String())
}

func solveit() {
	inp, err := LoadFile("range1.txt")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	b := RangeBoardFromLines(inp)

	fmt.Printf("%s\n", b.StringVerbose())
	b.Solve()
	fmt.Printf("*********************************************************\n%s\n", b.StringVerbose())
	c, d := b.IsSolved()
	fmt.Printf("Solved: %v (%v)\n", c, d)
	return
}
