package puzzle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type Puzzle struct {
	version   string
	cols      int
	rows      int
	solution  []string
	state     []string
	title     string
	author    string
	copyright string
	clues     []string
	notes     string
}

type puzzleChecksums struct {
	file      uint16
	cib       uint16
	masked    [8]byte
	scrambled uint16
}

type puzzleAux struct {
	numClues int
}

type puzzleReader interface {
	Read([]byte) (int, error)
	ReadString(delim byte) (string, error)
}

// Load parses the Across Lite data format and fills in the necessary fields in
// the Puzzle.
//
func (p *Puzzle) Load(data []byte) error {
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
	sum := &puzzleChecksums{}
	buf := bytes.NewBuffer(data)

	// Whole-file checksum
	err := binary.Read(buf, binary.LittleEndian, &sum.file)
	if err != nil {
		return err
	}

	// Magic bytes
	magic := "ACROSS&DOWN\x00"
	magicBytes := make([]byte, len(magic))
	err = binary.Read(buf, binary.LittleEndian, &magicBytes)
	if err != nil {
		return err
	}
	if string(magicBytes) != magic {
		return errors.New("this doesn't look like an Across Lite formatted puzzle")
	}

	// Checksums
	err = binary.Read(buf, binary.LittleEndian, &sum.cib)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &sum.masked)
	if err != nil {
		return err
	}

	// Version string
	var version [4]byte
	err = binary.Read(buf, binary.LittleEndian, &version)
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		if version[i] == 0x00 {
			p.version = string(version[:i])
			break
		}
		p.version = string(version[:])
	}

	// Junk? ("Reserved1C: In many files, this is uninitialized memory")
	buf.Next(2)

	// Scrambled puzzle checksum
	err = binary.Read(buf, binary.LittleEndian, &sum.scrambled)
	if err != nil {
		return err
	}

	// Junk? ("Reserved20: In files where Reserved1C is garbage, this is garbage
	// too.")
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
	p.cols = int(cols)
	p.rows = int(rows)

	// Number of clues
	var numClues uint16
	err = binary.Read(buf, binary.LittleEndian, &numClues)
	if err != nil {
		return err
	}

	// Junk? ("Unknown bitmask: A bitmask. Operations unknown.")
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
	p.solution, err = readMatrix(buf, p.rows, p.cols)
	if err != nil {
		return err
	}

	// Read cell state matrix
	p.state, err = readMatrix(buf, p.rows, p.cols)
	if err != nil {
		return err
	}

	p.title, err = readString(buf)
	if err != nil {
		return err
	}

	p.author, err = readString(buf)
	if err != nil {
		return err
	}

	p.copyright, err = readString(buf)
	if err != nil {
		return err
	}

	p.clues = make([]string, 0)
	for i := 0; i < int(numClues); i++ {
		str, err := readString(buf)
		if err != nil {
			return err
		}
		p.clues = append(p.clues, str)
	}

	p.notes, err = readString(buf)
	// Some puzzles just fall right off the end of the notes field, so io.EOF is
	// allowable here and shouldn't cause an error.
	if err != nil && err != io.EOF {
		return err
	}

	// Having verified the structure of the puzzle, now check the checksums.
	if !p.verify(data, sum) {
		return errors.New("checksums aren't correct")
	}

	// TODO: parse extra data
	//
	// https://code.google.com/p/puz/wiki/FileFormat#Extra_Sections

	return nil
}

func (p *Puzzle) verify(data []byte, sum *puzzleChecksums) bool {

	size := p.cols * p.rows

	dCib := data[0x2c : 0x2c+8]
	dSol := data[0x34 : 0x34+size]
	dState := data[0x34+size : 0x34+size+size]
	// 8 bytes starting at the puzzle width
	cCib := cksum(dCib, 0x0000)
	if cCib != sum.cib {
		return false
	}

	// Whole-file checksum
	cFile := cCib
	// The solution grid
	cFile = cksum(dSol, cFile)
	// The state grid
	cFile = cksum(dState, cFile)

	// Masked checksums
	cSol := cksum(dSol, 0x0000)
	cState := cksum(dState, 0x0000)
	cPart := uint16(0x0000)

	// The rest of the file
	c := 0 // cursor to track position
	d := data[0x34+size+size:]

	if len(p.title) > 0 {
		cFile = cksum(d[c:c+len(p.title)+1], cFile)
		cPart = cksum(d[c:c+len(p.title)+1], cPart)
		c += len(p.title)
	}
	c++

	if len(p.author) > 0 {
		cFile = cksum(d[c:c+len(p.author)+1], cFile)
		cPart = cksum(d[c:c+len(p.author)+1], cPart)
		c += len(p.author)
	}
	c++

	if len(p.copyright) > 0 {
		cFile = cksum(d[c:c+len(p.copyright)+1], cFile)
		cPart = cksum(d[c:c+len(p.copyright)+1], cPart)
		c += len(p.copyright)
	}
	c++

	for _, clue := range p.clues {
		cFile = cksum(d[c:c+len(clue)], cFile)
		cPart = cksum(d[c:c+len(clue)], cPart)
		c += len(clue) + 1
	}

	// The notes field is only included in the file and masked checksums from 1.3
	// onwards
	if p.version != "1.2" && p.version != "1.2c" && len(p.notes) > 0 {
		cFile = cksum(d[c:c+len(p.notes)+1], cFile)
		cPart = cksum(d[c:c+len(p.notes)+1], cPart)
	}

	if cFile != sum.file {
		return false
	}

	cMasked := [8]byte{
		0x49 ^ byte(cCib&0xFF),
		0x43 ^ byte(cSol&0xFF),
		0x48 ^ byte(cState&0xFF),
		0x45 ^ byte(cPart&0xFF),
		0x41 ^ byte((cCib&0xFF00)>>8),
		0x54 ^ byte((cSol&0xFF00)>>8),
		0x45 ^ byte((cState&0xFF00)>>8),
		0x44 ^ byte((cPart&0xFF00)>>8),
	}

	if cMasked != sum.masked {
		return false
	}

	return true
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

func readMatrix(rd puzzleReader, rows, cols int) ([]string, error) {
	var matrix = make([]string, rows, rows)
	for i := 0; i < rows; i++ {
		var row = make([]byte, cols, cols)
		err := binary.Read(rd, binary.LittleEndian, &row)
		if err != nil {
			return nil, err
		}
		matrix[i] = string(row)
	}
	return matrix, nil
}

func readString(rd puzzleReader) (string, error) {
	out, err := rd.ReadString(0x0)
	if err != nil {
		return out, err
	}
	// Chop NUL off the end
	return out[:len(out)-1], nil
}
