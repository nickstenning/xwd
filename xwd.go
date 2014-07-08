package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/nickstenning/xwd/puzzle"
)

var logger = log.New(os.Stderr, "xwd: ", log.LstdFlags)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("you must supply a .puz file")
	}

	filename := flag.Arg(0)

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	puz := &puzzle.Puzzle{}
	err = puz.Load(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v\n", puz)
}
