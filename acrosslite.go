package xwd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// AcrossLite holds the parsed data from an AcrossLite (".puz") puzzle file
type AcrossLite struct {
	Data      []byte
	Cols      int
	Rows      int
	Title     string
	Author    string
	Copyright string
	Notes     string
	Grid      []string
	Solution  []string
	Clues     []string
	CksumFil  uint16
	CksumCib  uint16
	CksumMsk  [8]byte
	CksumScr  uint16
}

// Sniff looks at the provided data slice and returns a boolean indicating
// whether it looks like an AcrossLite puzzle.
func (a *AcrossLite) Sniff(data []byte) bool {
	buf := bytes.NewBuffer(data)

	buf.Next(2)

	// Magic bytes
	magic := make([]byte, 12)
	_, err := buf.Read(magic)
	if err == nil && string(magic) == "ACROSS&DOWN\x00" {
		return true
	}

	return false
}

// Parse parses the AcrossLite data to fills in the data structure.
func (a *AcrossLite) Parse(data []byte) error {
	// This code is based on the following structure table, taken from
	//
	//     https://code.google.com/p/puz/wiki/FileFormat#Extra_Sections
	//
	// Component           Offset Len  Type       Description
	// ------------------- ------ ---  ---------  -----------
	// Checksum            0x00   0x2  uint16     overall file checksum
	// File Magic          0x02   0xC  string     NUL-terminated constant string:
	//                                            "ACROSS&DOWN\0"
	// CIB Checksum        0x0E   0x2  uint16
	// Masked Checksum     0x10   0x8  [8]byte    a set of checksums, XOR-masked
	//                                            against a magic string
	// Version String      0x18   0x4  string     version, e.g. "1.2\0"
	// Reserved1C(?)       0x1C   0x2  ?          in many files, this is
	//                                            uninitialized memory
	// Scrambled Checksum  0x1E   0x2  short      in scrambled puzzles, a checksum
	//                                            of the real solution, otherwise
	//                                            0x0000
	// Reserved20(?)       0x20   0xC  ?          in files where Reserved1C is
	//                                            garbage, this is garbage too
	// Width               0x2C   0x1  byte       width of the board
	// Height              0x2D   0x1  byte       height of the board
	// # of Clues          0x2E   0x2  uint16     number of clues for this board
	// Unknown Bitmask     0x30   0x2  uint16     a bitmask, use unknown
	// Scrambled Tag       0x32   0x2  uint16     0 for unscrambled puzzles,
	//                                            nonzero (often 4) for scrambled
	//                                            puzzles
	a.Data = make([]byte, len(data))
	copy(a.Data, data)

	buf := bytes.NewBuffer(data)

	// Whole-file checksum
	err := binary.Read(buf, binary.LittleEndian, &a.CksumFil)
	if err != nil {
		return err
	}

	// Magic bytes
	buf.Next(12)

	// Checksums
	err = binary.Read(buf, binary.LittleEndian, &a.CksumCib)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &a.CksumMsk)
	if err != nil {
		return err
	}

	// Version string. We skip this because we just don't care, and while it has
	// an impact on how the checksum is calculated that's easily dealt with in
	// other ways (see below).
	buf.Next(4)

	// Junk (Reserved1C)
	buf.Next(2)

	// Scrambled puzzle checksum
	err = binary.Read(buf, binary.LittleEndian, &a.CksumScr)
	if err != nil {
		return err
	}

	// Junk (Reserved20)
	buf.Next(12)

	// Dimensions
	var cols, rows byte
	err = binary.Read(buf, binary.LittleEndian, &cols)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &rows)
	if err != nil {
		return err
	}
	a.Cols = int(cols)
	a.Rows = int(rows)

	// Number of clues
	var numClues uint16
	err = binary.Read(buf, binary.LittleEndian, &numClues)
	if err != nil {
		return err
	}

	// Junk (Unknown bitmask)
	buf.Next(2)

	// Scrambled tag
	var scrambled uint16
	err = binary.Read(buf, binary.LittleEndian, &scrambled)
	if err != nil {
		return err
	}
	if scrambled != 0x0000 {
		return errors.New("no support yet for scrambled puzzles")
	}

	// Read solution matrix
	a.Solution, err = readMatrix(buf, a.Rows, a.Cols)
	if err != nil {
		return err
	}

	// Read cell grid matrix. We don't need this as it's implied by the solution
	// matrix.
	buf.Next(a.Rows * a.Cols)

	a.Title, err = readString(buf)
	if err != nil {
		return err
	}

	a.Author, err = readString(buf)
	if err != nil {
		return err
	}

	a.Copyright, err = readString(buf)
	if err != nil {
		return err
	}

	a.Clues = make([]string, 0)
	for i := 0; i < int(numClues); i++ {
		str, err := readString(buf)
		if err != nil {
			return err
		}
		a.Clues = append(a.Clues, str)
	}

	a.Notes, err = readString(buf)
	// Some puzzles just fall right off the end of the notes field, so io.EOF is
	// allowable here and shouldn't cause an error.
	if err != nil && err != io.EOF {
		return err
	}

	// Having verified the structure of the puzzle, now check the checksums.
	if !a.Verify() {
		return errors.New("checksums aren't correct")
	}

	// TODO: parse extra data
	//
	// https://code.google.com/p/puz/wiki/FileFormat#Extra_Sections

	return nil
}

// Verify checks that the puzzle's checksums are valid.
func (a *AcrossLite) Verify() bool {
	size := a.Cols * a.Rows

	dCib := a.Data[0x2c : 0x2c+8]
	dSol := a.Data[0x34 : 0x34+size]
	dGrid := a.Data[0x34+size : 0x34+size+size]
	// 8 bytes starting at the puzzle width
	cCib := cksum(dCib, 0x0000)
	if cCib != a.CksumCib {
		return false
	}

	// Whole-file checksum
	cFile := cCib
	// The solution grid
	cFile = cksum(dSol, cFile)
	// The grid grid
	cFile = cksum(dGrid, cFile)

	// Masked checksums
	cSol := cksum(dSol, 0x0000)
	cGrid := cksum(dGrid, 0x0000)
	cPart := uint16(0x0000)

	// The rest of the file
	c := 0 // cursor to track position
	d := a.Data[0x34+size+size:]

	if len(a.Title) > 0 {
		cFile = cksum(d[c:c+len(a.Title)+1], cFile)
		cPart = cksum(d[c:c+len(a.Title)+1], cPart)
		c += len(a.Title)
	}
	c++

	if len(a.Author) > 0 {
		cFile = cksum(d[c:c+len(a.Author)+1], cFile)
		cPart = cksum(d[c:c+len(a.Author)+1], cPart)
		c += len(a.Author)
	}
	c++

	if len(a.Copyright) > 0 {
		cFile = cksum(d[c:c+len(a.Copyright)+1], cFile)
		cPart = cksum(d[c:c+len(a.Copyright)+1], cPart)
		c += len(a.Copyright)
	}
	c++

	for _, clue := range a.Clues {
		cFile = cksum(d[c:c+len(clue)], cFile)
		cPart = cksum(d[c:c+len(clue)], cPart)
		c += len(clue) + 1
	}

	// The notes field is only included in the file and masked checksums from
	// 1.3 onwards. In older versions (1.2 and 1.2c are the only ones I've seen)
	// the file checksum should be correct at this point, so we won't enter this
	// conditional in those cases.
	if cFile != a.CksumFil && len(a.Notes) > 0 {
		cFile = cksum(d[c:c+len(a.Notes)+1], cFile)
		cPart = cksum(d[c:c+len(a.Notes)+1], cPart)
	}

	if cFile != a.CksumFil {
		return false
	}

	// ICHEATED. Oh ho ho. Very funny.
	cMasked := [8]byte{
		0x49 ^ byte(cCib&0xFF),
		0x43 ^ byte(cSol&0xFF),
		0x48 ^ byte(cGrid&0xFF),
		0x45 ^ byte(cPart&0xFF),
		0x41 ^ byte((cCib&0xFF00)>>8),
		0x54 ^ byte((cSol&0xFF00)>>8),
		0x45 ^ byte((cGrid&0xFF00)>>8),
		0x44 ^ byte((cPart&0xFF00)>>8),
	}

	if cMasked != a.CksumMsk {
		return false
	}

	return true
}

// Load loads data from the parsed AcrossLite puzzle into the provided Puzzle
// object.
func (a *AcrossLite) Load(p *Puzzle) {
	p.Rows = a.Rows
	p.Cols = a.Cols
	p.Title = a.Title
	p.Author = a.Author
	p.Copyright = a.Copyright
	p.Notes = a.Notes
	p.SetSolution(a.Solution)
	a.loadClues(p)
}

func (a *AcrossLite) loadClues(p *Puzzle) {
	aIdx := 0
	dIdx := 0
	aMax := len(p.cluesAcross)
	dMax := len(p.cluesDown)

	for i := 0; i < len(a.Clues); i++ {
		if dIdx >= dMax {
			// We're out of down squares, so this must be an across clue
			p.cluesAcross[aIdx].Clue = a.Clues[i]
			aIdx++
			continue
		}
		if aIdx >= aMax {
			// We're out of across squares, so this must be a down clue
			p.cluesDown[dIdx].Clue = a.Clues[i]
			dIdx++
			continue
		}
		// Now we pick the next lowest numbered square. If a square has both an
		// across and a down clue, the across clue comes first.
		if p.cluesDown[dIdx].Num < p.cluesAcross[aIdx].Num {
			p.cluesDown[dIdx].Clue = a.Clues[i]
			dIdx++
		} else {
			p.cluesAcross[aIdx].Clue = a.Clues[i]
			aIdx++
		}
	}
}

func cksum(data []byte, sum uint16) uint16 {
	for _, b := range data {
		if (sum & 0x0001) != 0 {
			sum = (sum >> 1) + 0x8000
		} else {
			sum = sum >> 1
		}
		sum += uint16(b)
	}
	return sum
}

func readMatrix(buf *bytes.Buffer, rows, cols int) ([]string, error) {
	matrix := make([]string, rows)
	for i := 0; i < rows; i++ {
		bytes := make([]byte, cols)
		err := binary.Read(buf, binary.LittleEndian, &bytes)
		if err != nil {
			return nil, err
		}
		matrix[i] = string(bytes)
	}
	return matrix, nil
}

func readString(buf *bytes.Buffer) (string, error) {
	out, err := buf.ReadString(0x0)
	if err != nil {
		return out, err
	}
	// Chop NUL off the end
	return out[:len(out)-1], nil
}
