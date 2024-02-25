package main

import (
	"fmt"
	"os"
	"strings"
)

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
	// return
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
