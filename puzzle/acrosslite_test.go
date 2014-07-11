package puzzle

import "testing"

type CksumExample struct {
	in    []byte
	start uint16
	out   uint16
}

var cksumExamples = []CksumExample{
	{[]byte{}, 0x0000, 0x0000},
	{[]byte{1, 2, 3, 4, 5}, 0x0000, 0x1008},
	{[]byte{7, 7, 7, 7, 7, 7}, 0x002a, 0x700e},
	{[]byte{100, 100, 100}, 0xffff, 0xc0ae},
}

func TestCksum(t *testing.T) {
	for _, ex := range cksumExamples {
		testCksum(t, ex)
	}
}

func testCksum(t *testing.T, ex CksumExample) {
	if res := cksum(ex.in, ex.start); res != ex.out {
		t.Errorf("checksum was wrong (expected 0x%04x, got 0x%04x)", ex.out, res)
	}
}
