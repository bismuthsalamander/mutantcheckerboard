package main

import (
	"fmt"
)

type TowerBoard struct {
	RectNumBoard
	Order     int
	Observers []*Observer
	ObsSorted []*Observer
	Perms     [][]int
	RowPerms  []*[]int
	ColPerms  []*[]int
}

type Observer struct {
	Start     Coord
	Direction Delta
	Count     int
}

func (o Observer) IsRow() bool {
	return o.Direction.X != 0
}

func (o Observer) IsCol() bool {
	return o.Direction.Y != 0
}

func (o Observer) IsBackwards() bool {
	return o.Direction.X < 0 || o.Direction.Y < 0
}

func (o Observer) IsForwards() bool {
	return !o.IsBackwards()
}

func TowerBoardFromLines(input [][]int) (*TowerBoard, error) {
	if len(input) < 1 || len(input[0]) != 1 {
		return nil, fmt.Errorf("first line must contain ")
	}
	order := input[0][0]
	if order <= 0 {
		return nil, fmt.Errorf("order must be >= 1; got %d", order)
	}
	boardLines := make([][]int, 0, len(input)-3)
	for idx, line := range input {
		if idx < 2 {
			continue
		} else if idx == len(input)-1 {
			continue
		}
		boardLines = append(boardLines, line[1:len(line)-1])
	}
	rect := RectNumBoardFromNums(boardLines)
	allObservers := make([]*Observer, 0, rect.W*rect.H*2)
	obsSorted := make([]*Observer, rect.W*rect.H*2)
	b := TowerBoard{
		RectNumBoard: *rect,
		Order:        order,
		Observers:    allObservers,
		ObsSorted:    obsSorted,
		Perms:        PermuteN(order),
		RowPerms:     make([]*[]int, rect.H),
		ColPerms:     make([]*[]int, rect.W),
	}
	b.Allowed = MakeAllowedSets(rect.W, rect.H, order)
	// row obs
	for ri := 0; ri < b.H; ri++ {
		b.AddObs(Coord{X: 0, Y: ri}, Delta{X: 1, Y: 0}, input[ri+2][0])
		b.AddObs(Coord{X: b.W - 1, Y: ri}, Delta{X: -1, Y: 0}, input[ri+2][b.W+1])
	}
	// col obs
	for ci := 0; ci < b.W; ci++ {
		b.AddObs(Coord{X: ci, Y: 0}, Delta{X: 0, Y: 1}, input[1][ci+1])
		b.AddObs(Coord{X: ci, Y: b.H - 1}, Delta{X: 0, Y: -1}, input[b.H+2][ci+1])
	}
	// init perms
	b.PopulateRowColPerms()
	b.Inited = true
	b.EachCell(func(c Coord, v int) bool {
		b.PostMark(c, v)
		return false
	})
	return &b, nil
}

func (b *TowerBoard) Mark(c Coord, v int) (bool, error) {
	res, err := b.Set(c, v)
	if !res {
		return res, err
	}
	b.SetDirty()
	b.PostMark(c, v)
	return res, err
}

func (b *TowerBoard) PostMark(c Coord, v int) (bool, error) {
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
		}
	}
	return changed, nil
}

func (b *TowerBoard) PopulateRowColPerms() {
	pi := 0
	for ri := 0; ri < b.H; ri++ {
		b.RowPerms[ri] = b.PermsForObs(b.ObsSorted[pi], b.ObsSorted[pi+1])
		pi += 2
	}
	for ci := 0; ci < b.W; ci++ {
		b.ColPerms[ci] = b.PermsForObs(b.ObsSorted[pi], b.ObsSorted[pi+1])
		pi += 2
	}
}

func (b *TowerBoard) AddObs(start Coord, d Delta, ct int) *Observer {
	idx := b.ObsIndex(start, d)
	if ct == 0 {
		return nil
	}
	o := Observer{
		Start:     start,
		Direction: d,
		Count:     ct,
	}

	b.ObsSorted[idx] = &o
	return &o
}

func (b *TowerBoard) ObsIndex(start Coord, d Delta) int {
	// Order is row 0 fwd, row 0 bwd, ... row n-1 fwd, row n-1 bwd
	// Then col 0 fwd, col 0 bwd, ...
	var idx int
	if d.Y == 0 { // column
		idx = start.Y * 2
		if d.X == -1 {
			idx++
		}
	} else { // row
		idx = (start.X + b.H) * 2
		if d.Y == -1 {
			idx++
		}
	}
	return idx
}

// PermsForObs generates a slice of the permutation indexes that fit both
// observers. If both are nil, returns nil. Must be called after b.Perms has
// been initialized.
func (b *TowerBoard) PermsForObs(fwd, bwd *Observer) *[]int {
	if fwd == nil && bwd == nil {
		return nil
	}
	out := make([]int, 0)
	for i := 0; i < len(b.Perms); i++ {
		if PermFitsObs(b.Perms[i], fwd, bwd) {
			out = append(out, i)
		}
	}
	return &out
}

// PermFitsObs checks whether a given row or column is consistent with both
// observers. Nil inputs are ignored, so PermFitsObs(_, nil, nil) always
// returns true.
func PermFitsObs(p []int, fwd, bwd *Observer) bool {
	if fwd != nil {
		vis := 0
		highest := 0
		for i := 0; i < len(p); i++ {
			if p[i] > highest {
				highest = p[i]
				vis++
				if vis > fwd.Count {
					return false
				}
			}
		}
		if vis != fwd.Count {
			return false
		}
	}
	if bwd != nil {
		vis := 0
		highest := 0
		for i := len(p) - 1; i >= 0; i-- {
			if p[i] > highest {
				highest = p[i]
				vis++
				if vis > bwd.Count {
					return false
				}
			}
		}
		if vis != bwd.Count {
			return false
		}
	}
	return true
}

// ObsChar is a helper function that locates the observer specified by the
// t(ype), index and direction parameters, then returns a string to be
// displayed in the board string.
func (b *TowerBoard) ObsChar(start Coord, d Delta) string {
	idx := b.ObsIndex(start, d)
	o := b.ObsSorted[idx]
	if o == nil {
		return " "
	}
	return string(IntToCh(o.Count))
}

func (b *TowerBoard) Get(c Coord) int {
	if !b.IsValid(c) {
		return 0
	}
	return b.Grid[c.Y][c.X]
}

// CharAt generates a character for the specified cell in the board's grid.
func (b *TowerBoard) CharAt(coord Coord) string {
	if !b.IsValid(coord) {
		return " "
	}
	return string(IntToCh(b.Get(coord)))
}

func (b *TowerBoard) String() string {
	out := " "
	for ci := 0; ci < b.W; ci++ {
		out += b.ObsChar(Coord{X: ci, Y: 0}, Delta{X: 0, Y: 1})
	}
	out += "\n"
	for ri := 0; ri < b.H; ri++ {
		out += b.ObsChar(Coord{X: 0, Y: ri}, Delta{X: 1, Y: 0})
		for ci := 0; ci < b.W; ci++ {
			out += b.CharAt(Coord{ci, ri})
		}
		out += b.ObsChar(Coord{X: b.W - 1, Y: ri}, Delta{X: -1, Y: 0})
		out += "\n"
	}
	out += " "
	for ci := 0; ci < b.W; ci++ {
		out += b.ObsChar(Coord{X: ci, Y: b.H - 1}, Delta{X: 0, Y: -1})
	}
	return out
}

// Solved returns true iff all observers are satisfied and all cells are
// filled. As of now, it does not confirm that the sudoku rule is satisfied.
func (b *TowerBoard) IsSolved() (bool, error) {
	for c := b.TopLeft(); b.IsValid(c); c = b.Next(c) {
		if b.Get(c) == UNKNOWN {
			return false, fmt.Errorf("cell %s is unknown", c)
		}
	}
	for _, o := range b.Observers {
		if ok, num := b.ObserverSatisfied(o); !ok {
			return false, fmt.Errorf("observer %v unsatisfied (want %d, see %d)", o, o.Count, num)
		}
	}
	for _, r := range b.AllRegions {
		if !b.IsRegionSolved(*r) {
			return false, fmt.Errorf("region %v unsatisfied", *r)
		}
	}
	return true, nil
}

func (b *TowerBoard) IsRegionSolved(r []Coord) bool {
	ns := NewNumSet(len(r))
	for _, c := range r {
		ns.Del(b.Get(c))
	}
	return ns.Size() == 0
}

// ObserverSatisfied returns false if the grid's contents are consistent with
// the constraint for the specified observer. This function will treat empty
// cells as a zero (meaning that such cells are never visible and never
// obstruct other cells), so the return value may be misleading if called when
// the relevant row or column is incomplete.
func (b *TowerBoard) ObserverSatisfied(o *Observer) (bool, int) {
	vis := 0
	highest := 0
	for c := o.Start; b.IsValid(c); c = c.Plus(o.Direction) {
		val := b.Get(c)
		if val > highest {
			vis++
			highest = val
		}
	}
	return vis == o.Count, vis
}

// MarkMandatory searches for cells with only one entry in Allowed and marks
// the appropriate value. Returns true iff a change was made. The redo flag
// repeats the loop if marking a mandatory cell eliminated entries in Allowed
// for other cells in the same row or column. If, for example, cell (a, b) is
// the last empty cell in its row and column, then marking (a, b) will not
// eliminate any possibilities from other cells, and repeating the loop is not
// necessary.
func (b *TowerBoard) MarkMandatory() bool {
	changed := false
	redo := false
	b.EachCell(func(c Coord, v int) bool {
		allowed := b.Allowed[c.Y][c.X]
		if len(allowed.M) != 1 || b.Get(c) != UNKNOWN {
			return false
		}
		k := allowed.GetOne()
		_, err := b.Mark(c, k)
		if err != nil {
			panic(err)
		}
		return false
	})
	if redo {
		b.MarkMandatory()
	}
	return changed
}

// TrimPermsFromAllowed removes entries in RowPerns and ColPerms that are not
// possible because they would violate the Allowed maps. Returns true iff any
// changes were made.
func (b *TowerBoard) TrimPermsFromAllowed() bool {
	changed := false
	for ri, rp := range b.RowPerms {
		if rp == nil {
			continue
		}
		newPerms := make([]int, 0, len(*rp))
		for _, pi := range *rp {
			isPermOk := true
			for ci := 0; ci < b.W; ci++ {
				if !b.IsAllowed(Coord{ci, ri}, b.Perms[pi][ci]) {
					isPermOk = false
					break
				}
			}
			if isPermOk {
				newPerms = append(newPerms, pi)
			}
		}
		if len(*rp) != len(newPerms) {
			b.RowPerms[ri] = &newPerms
			changed = true
		}
	}
	for ci, cp := range b.ColPerms {
		if cp == nil {
			continue
		}
		newPerms := make([]int, 0, len(*cp))
		for _, pi := range *cp {
			isPermOk := true
			for ri := 0; ri < b.H; ri++ {
				if !b.IsAllowed(Coord{ci, ri}, b.Perms[pi][ri]) {
					isPermOk = false
					break
				}
			}
			if isPermOk {
				newPerms = append(newPerms, pi)
			}
		}
		if len(*cp) != len(newPerms) {
			b.ColPerms[ci] = &newPerms
			changed = true
		}
	}
	return changed
}

// TrimAllowedFromPerms will eliminate a permutation from RowPerms or ColPerms
// if it is inconsistent with any cell's Allowed list. Returns true iff at
// least one permutation was eliminated.
func (b *TowerBoard) TrimAllowedFromPerms() bool {
	changed := false
	for ri := 0; ri < b.H; ri++ {
		for ci := 0; ci < b.W; ci++ {
			for n, _ := range b.Allowed[ri][ci].M {
				//Is n allowed in slot ci in a perm for row ri?
				found := false
				for _, permI := range *b.RowPerms[ri] {
					if b.Perms[permI][ci] == n {
						found = true
						break
					}
				}
				if !found {
					b.Disallow(Coord{ci, ri}, n)
					b.Allowed[ri][ci].Del(n)
					changed = true
					continue
				}
				//Is n allowed in slot ri in a perm for col ci?
				found = false
				if b.ColPerms[ci] != nil {
					for _, permI := range *b.ColPerms[ci] {
						if b.Perms[permI][ri] == n {
							found = true
							break
						}
					}
					if !found {
						b.Allowed[ri][ci].Del(n)
						changed = true
					}
				}
			}
		}
	}
	return changed
}

// AutoSolve runs all implemented solving heuristics until the puzzle is solved
// or we run out of improvements. Missing heuristics include the opposite of
// naked sets (i.e., cells X and Y are the only possible locations for numbers
// N and M, so X and Y can't have any other numbers) and pairwise permutation
// consistency between rows or columns.
func (b *TowerBoard) Solve() (bool, error) {
	changed := true
	for _, err := b.IsSolved(); changed && err != nil; {
		changed = false
		if b.MarkMandatory() {
			changed = true
		}
		if b.TrimAllowedFromPerms() {
			changed = true
		}
		if b.TrimPermsFromAllowed() {
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
