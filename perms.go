package main

// permuter is a struct that manages state for the recursive permutation
// function.
type permuter struct {
	N      int
	R      int
	Lowest int
	Seq    []int
	Used   []bool
}

type permuterParams struct {
	low  int
	high int
	r    int
}

var permMemo map[permuterParams][][]int

func init() {
	permMemo = make(map[permuterParams][][]int)
}

// fact returns n! for any nonnegative input n.
func fact(n int) int {
	if n <= 1 {
		return 1
	}
	if n == 2 {
		return 2
	}
	return n * fact(n-1)
}

// Permute is the main public permutation API function. Returns all slices of
// r integers between low and high *inclusive*.
func Permute(low, high, r int) [][]int {
	param := permuterParams{low, high, r}
	if out, ok := permMemo[param]; ok {
		return out
	}
	popSize := (high - low) + 1
	out := make([][]int, 0, fact(popSize)/(fact(popSize-r)))
	p := permuter{
		N:      popSize,
		Lowest: low,
		R:      r,
		Seq:    make([]int, r),
		Used:   make([]bool, popSize),
	}
	p.permute(0, &out)
	permMemo[param] = out
	return out
}

// NPermuteR is a helper function that returns Permute(1, n, r).
func NPermuteR(n int, r int) [][]int {
	return Permute(1, n, r)
}

// PermuteN is a helper function that returns Permute(1, n, n).
func PermuteN(n int) [][]int {
	return Permute(1, n, n)
}

// Permute is the main recursive permutation function.
func (p *permuter) permute(depth int, output *[][]int) {
	if depth == p.R {
		tmp := make([]int, len(p.Seq))
		copy(tmp, p.Seq)
		*output = append(*output, tmp)
		return
	}
	for i := 0; i < p.N; i++ {
		if p.Used[i] {
			continue
		}
		p.Seq[depth] = i + p.Lowest
		p.Used[i] = true
		p.permute(depth+1, output)
		p.Seq[depth] = 0
		p.Used[i] = false
	}
}
