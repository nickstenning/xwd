package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"github.com/nickstenning/xwd"
)

var showSolution = flag.Bool("s", false, "show the solution rather than the blank puzzle")
var logger = log.New(os.Stderr, "xwd: ", log.LstdFlags)

var boxTop = []string{"┌", "─", "┬", "┐"}
var boxLin = []string{"│", "█", "│", "│"}
var boxMid = []string{"├", "─", "┼", "┤"}
var boxBot = []string{"└", "─", "┴", "┘"}

func usage() {
	fmt.Fprintf(
		os.Stderr,
		"Usage: %s [options] <puzzlefile>\n\n",
		path.Base(os.Args[0]),
	)
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("you must supply a .puz file")
	}

	filename := flag.Arg(0)
	f, err := os.Open(filename)
	if err != nil {
		logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		logger.Fatal(err)
	}

	puz := &xwd.Puzzle{}
	err = puz.Load(data)
	if err != nil {
		logger.Fatal(err)
	}

	printGrid(puz, *showSolution)

	fmt.Printf("\nAcross:\n\n")
	across := puz.CluesAcross()
	max := across[len(across)-1].Num + 1
	wrapw := int(math.Floor(math.Log10(float64(max)))) + 1
	for _, c := range across {
		fmt.Printf("%*d. %s\n", wrapw, c.Num+1, c.Clue)
	}

	fmt.Printf("\nDown:\n\n")
	down := puz.CluesDown()
	max = down[len(down)-1].Num + 1
	wrapw = int(math.Floor(math.Log10(float64(max)))) + 1
	for _, c := range down {
		fmt.Printf("%*d. %s\n", wrapw, c.Num+1, c.Clue)
	}
}

func printGrid(p *xwd.Puzzle, solution bool) {
	for i := 0; i < p.Rows; i++ {
		printRow(p, i, solution)
	}
}

func printRow(p *xwd.Puzzle, row int, solution bool) {
	if row == 0 {
		fmt.Print(boxDivider(p.Cols, boxTop))
	} else {
		fmt.Print(boxDivider(p.Cols, boxMid))
	}
	fmt.Print(boxRow(p, row, solution))
	if row+1 == p.Rows {
		fmt.Print(boxDivider(p.Cols, boxBot))
	}
}

func boxDivider(nCells int, box []string) string {
	out := make([]string, 0, nCells)
	out = append(out, strings.Join([]string{box[0], strings.Repeat(box[1], 3)}, ""))
	for j := 1; j < nCells-1; j++ {
		out = append(out, strings.Join([]string{box[2], strings.Repeat(box[1], 3)}, ""))
	}
	out = append(out, strings.Join([]string{box[2], strings.Repeat(box[1], 3), box[3], "\n"}, ""))
	return strings.Join(out, "")
}

func boxRow(p *xwd.Puzzle, row int, solution bool) string {
	out := []string{}
	i := row
	for j := 0; j < p.Cols; j++ {
		cell, err := p.Cell(i, j)
		if err != nil {
			panic(err)
		}
		cs := "   "
		if cell.Black {
			cs = strings.Repeat(boxLin[1], 3)
		} else if solution {
			cs = fmt.Sprintf(" %s ", cell.Solution)
		} else if cell.Num != -1 {
			cs = fmt.Sprintf("%3d", cell.Num+1) // cell.Num is zero-indexed
		}
		if j == 0 {
			out = append(out, strings.Join([]string{boxLin[0], cs, boxLin[2]}, ""))
		} else if j+1 == p.Cols {
			out = append(out, strings.Join([]string{cs, boxLin[3], "\n"}, ""))
		} else {
			out = append(out, strings.Join([]string{cs, boxLin[2]}, ""))
		}
	}
	return strings.Join(out, "")
}

func grey(s string) string {
	return strings.Join([]string{"\033[38;5;235m", s, "\033[39;49m"}, "")
}
