package puzzle

import "testing"

func TestSetSolution(t *testing.T) {
	p := &Puzzle{}
	err := p.SetSolution([]string{})
	if err != nil {
		t.Errorf("setting an empty grid on an empty puzzle failed")
	}

	p = &Puzzle{Rows: 3, Cols: 3}
	err = p.SetSolution([]string{
		"CAT",
		"..A",
		"..N",
	})
	if err != nil {
		t.Errorf("setting a 3x3 grid on a 3x3 puzzle failed")
	}

	p = &Puzzle{Rows: 3, Cols: 3}
	err = p.SetSolution([]string{
		"CAT",
		"..A",
	})
	if err == nil {
		t.Errorf("setting a 2x3 grid on a 3x3 puzzle incorrectly succeeded")
	}

	p = &Puzzle{Rows: 3, Cols: 3}
	err = p.SetSolution([]string{
		"CAT",
		"..",
		"..N",
	})
	if err == nil {
		t.Errorf("setting a ragged grid on a 3x3 puzzle incorrectly succeeded")
	}
}

type CellExample struct {
	puzzle []string
	tests  []CellTest
}

type CellTest struct {
	i, j  int
	black bool
	num   int
	sol   string
	err   error
}

var cellExamples = []CellExample{
	{
		puzzle: []string{
			"CATCH",
			".BOA.",
			"MARNE",
			".T.O.",
			"FERNY"},
		tests: []CellTest{
			{i: 0, j: 0, black: false, num: 0, sol: "C"},
			{i: 0, j: 1, black: false, num: 1, sol: "A"},
			{i: 0, j: 0, black: false, num: 0, sol: "C"},
			{i: 2, j: 0, black: false, num: 5, sol: "M"},
			{i: 1, j: 0, black: true, num: -1, sol: ""},
			{i: 1, j: 2, black: false, num: -1, sol: "O"},
			{i: 3, j: 4, black: true, num: -1, sol: ""},
			{i: 4, j: 4, black: false, num: -1, sol: "Y"},
			{i: -1, j: 4, err: OutOfBounds},
			{i: 4, j: -1, err: OutOfBounds},
			{i: 5, j: 4, err: OutOfBounds},
			{i: 4, j: 5, err: OutOfBounds},
		},
	},
}

func TestCell(t *testing.T) {
	for _, ex := range cellExamples {
		testCell(ex, t)
	}
}

func testCell(ex CellExample, t *testing.T) {
	p := &Puzzle{Rows: len(ex.puzzle), Cols: len(ex.puzzle[0])}
	err := p.SetSolution(ex.puzzle)
	if err != nil {
		t.Fatal(err)
	}
	for _, cx := range ex.tests {
		c, err := p.Cell(cx.i, cx.j)
		if err != cx.err {
			t.Errorf("expectation failure (expected: %v, got: %v)", cx.err, err)
		}

		if err != nil {
			continue
		}

		if c.Black != cx.black {
			t.Errorf("cell.Black expectation failure (expected: %v, got: %v)", cx.black, c.Black)
		}

		if c.Num != cx.num {
			t.Errorf("cell.Num expectation failure (expected: %v, got: %v)", cx.num, c.Num)
		}

		if c.Solution != cx.sol {
			t.Errorf("cell.Solution expectation failure (expected: %v, got: %v)", cx.sol, c.Solution)
		}
	}
}
