package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/akamensky/argparse"
)

func LoadFile(fn string) ([]string, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0)
	for _, txt := range strings.Split(string(data), "\n") {
		lines = append(lines, strings.Trim(txt, "\r\n"))
	}
	for len(lines) > 0 && len(lines[0]) == 0 {
		lines = lines[1:]
	}
	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-2]
	}
	return lines, nil
}

func LoadIntFile(fn string) ([][]int, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0)
	w := 0
	for _, txt := range strings.Split(string(data), "\n") {
		str := strings.Trim(txt, "\r\n")
		lines = append(lines, str)
		if len(str) > w {
			w = len(str)
		}
	}
	for len(lines) > 0 && len(lines[0]) == 0 {
		lines = lines[1:]
	}
	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-2]
	}

	grid := make([][]int, 0)
	for _, l := range lines {
		row := make([]int, len(l))
		for col, ch := range l {
			num, ok := CharToNum(ch)
			if ok {
				row[col] = num
			}
		}
		grid = append(grid, row)
	}

	return grid, nil
}

func main() {
	parser := argparse.NewParser("mutantcheckerboard", "Solver for binary determination puzzles")
	var puzzleType *string = parser.String("t", "type", &argparse.Options{
		Default: "kuromasu",
	})
	var inputFilename *string = parser.StringPositional(&argparse.Options{
		Required: true,
	})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Printf("error parsing command line arguments: %s", err)
		os.Exit(-1)
	}
	if inputFilename == nil || len(*inputFilename) == 0 {
		fmt.Printf("usage: %s -t [puzzle type] [input filename]\n", os.Args[0])
		os.Exit(-1)
	}
	switch *puzzleType {
	case "kuromasu":
		inp, err := LoadFile(*inputFilename)
		if err != nil {
			fmt.Printf("error loading file: %s\n", err)
			os.Exit(-1)
		}
		b := KuromasuBoardFromLines(inp)
		fmt.Printf("%s\n", b.String())
		b.Solve()
		fmt.Printf("%s\n", b.String())
		solved, err := b.IsSolved()
		if err != nil {
			fmt.Printf("Solved: %v (%s)\n", solved, err)
		} else {
			fmt.Printf("Solved: %v\n", solved)
		}
	case "towers":
		inp, err := LoadIntFile(*inputFilename)
		if err != nil {
			fmt.Printf("error loading file: %s\n", err)
			os.Exit(-1)
		}
		b, err := TowerBoardFromLines(inp)
		fmt.Printf("Board:\n\n%s\n\nerr: %s\n", b, err)
		b.Solve()
		fmt.Printf("Board:\n\n%s\n\nerr: %s\n", b, err)
		s, se := b.IsSolved()
		fmt.Printf("Solved: %v (%v)\n", s, se)

	default:
		fmt.Printf("unrecognized puzzle type \"%s\"\n", *puzzleType)
		os.Exit(-1)
	}
}
