package xwd

import (
	"errors"
	"fmt"
)

// Puzzle holds the data needed to represent a crossword puzzle
type Puzzle struct {
	Cols        int
	Rows        int
	Title       string
	Author      string
	Copyright   string
	Notes       string
	solution    []string
	cellCoords  map[[2]int]int
	cluesAcross []Clue
	cluesDown   []Clue
}

// Cell represents an individual cell in a crossword puzzle
type Cell struct {
	Black    bool   // Is the cell a "black" or unfillable cell
	Num      int    // If this is a numbered cell, the cell number, else -1
	Coords   [2]int // The coordinates of the cell, [2]int{<row>, <col>}
	Solution string // The provided solution for this cell
}

// Clue is a specific down or across clue for the puzzle
type Clue struct {
	Num  int    // The clue number
	Clue string // The text of the clue
}

// The character representing an unfillable cell in the crossword grid
const P_BLACK = "."

var NoProviderFound = errors.New("no provider found that knows how to load this puzzle")
var OutOfBounds = errors.New("the provided coordinates are out of bounds for this puzzle")

// Load uses the data provided to create the Puzzle. The data will be inspected
// to determine what type of puzzle it is. Currently only AcrossLite format is
// supported.
func (p *Puzzle) Load(data []byte) error {
	provider := &AcrossLite{}
	if provider.Sniff(data) {
		err := provider.Parse(data)
		if err != nil {
			return err
		}

		provider.Load(p)
		if err != nil {
			return err
		}
		return nil
	}
	return NoProviderFound
}

// SetSolution provides a way of directly setting the puzzle solution (and by
// implication, the puzzle grid). It accepts a slice of strings, of length
// puzzle.Rows. Each string must be of length puzzle.Cols. The ASCII period
// character 0x2e (".") is treated as the marker for a black cell. All other
// string data is treated as the puzzle solution.
//
// Setting the solution grid will also prefill the clue storage structures, so
// that CluesAcross and CluesDown will return slices of Clues, to which the
// free-text clue data can be attached directly.
func (p *Puzzle) SetSolution(grid []string) error {
	if len(grid) != p.Rows {
		return errors.New("grid should contain as many rows as the puzzle")
	}
	for i, r := range grid {
		if len(r) != p.Cols {
			return fmt.Errorf("grid rows should contain as many columns as the puzzle (failed at row %v)", i)
		}
	}

	p.solution = grid
	p.cellCoords = make(map[[2]int]int)
	p.cluesAcross = make([]Clue, 0)
	p.cluesDown = make([]Clue, 0)

	c := 0
	for i := 0; i < p.Rows; i++ {
		for j := 0; j < p.Cols; j++ {
			cellNumbered := false
			if p.isAcrossCell(i, j) {
				p.cellCoords[[2]int{i, j}] = c
				cellNumbered = true
				p.cluesAcross = append(p.cluesAcross, Clue{Num: c})
			}
			if p.isDownCell(i, j) {
				if !cellNumbered {
					p.cellCoords[[2]int{i, j}] = c
					cellNumbered = true
				}
				p.cluesDown = append(p.cluesDown, Clue{Num: c})
			}
			if cellNumbered {
				c++
			}
		}
	}

	return nil
}

// Solution returns a slice of rows (themselves slices of Cells) that can be
// used to range over the contents of this puzzle.
func (p *Puzzle) Solution() [][]Cell {
	rows := make([][]Cell, p.Rows)
	for i := 0; i < p.Rows; i++ {
		rows[i] = make([]Cell, p.Cols)
		for j := 0; j < p.Cols; j++ {
			c, err := p.Cell(i, j)
			if err != nil {
				panic(err)
			}
			rows[i][j] = *c
		}
	}
	return rows
}

func (p *Puzzle) isBlackCell(row, col int) bool {
	return p.solution[row][col] == P_BLACK[0]
}

func (p *Puzzle) isAcrossCell(row, col int) bool {
	if p.isBlackCell(row, col) {
		return false
	}
	if col == 0 || p.isBlackCell(row, col-1) {
		if col+1 < p.Cols && !p.isBlackCell(row, col+1) {
			return true
		}
	}
	return false
}

func (p *Puzzle) isDownCell(row, col int) bool {
	if p.isBlackCell(row, col) {
		return false
	}
	if row == 0 || p.isBlackCell(row-1, col) {
		if row+1 < p.Rows && !p.isBlackCell(row+1, col) {
			return true
		}
	}
	return false
}

// CluesAcross returns a slice of Clue structs, representing the "across" clues
// for the puzzle. The number of clues is implicit in the structure of the
// puzzle, and so this will always return the right number of clues for the
// current solution grid.
func (p *Puzzle) CluesAcross() []Clue {
	return p.cluesAcross
}

// CluesDown returns a slice of Clue structs, representing the "down" clues for
// the puzzle. The number of clues is implicit in the structure of the
// puzzle, and so this will always return the right number of clues for the
// current solution grid.
func (p *Puzzle) CluesDown() []Clue {
	return p.cluesDown
}

// Cell returns a Cell struct for the cell at row i, column j in the current
// puzzle. The coordinates are bounds-checked and the function will return
// puzzle.OutOfBounds if incorrect coordinates are given.
func (p *Puzzle) Cell(i, j int) (*Cell, error) {
	if i < 0 || i >= p.Rows || j < 0 || j >= p.Cols {
		return nil, OutOfBounds
	}
	coords := [2]int{i, j}
	num, ok := p.cellCoords[coords]
	if !ok {
		num = -1
	}
	c := &Cell{
		Black:  p.isBlackCell(i, j),
		Num:    num,
		Coords: coords,
	}
	if !c.Black {
		c.Solution = string(p.solution[i][j])
	}
	return c, nil
}
