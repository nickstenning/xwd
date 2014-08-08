package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/nickstenning/xwd"
)

var f = template.FuncMap{
	"inc":       func(n int) int { return n + 1 },
	"isnumcell": func(cell *xwd.Cell) bool { return cell.Num != -1 },
}
var t = template.Must(template.New("").Funcs(f).ParseFiles("puzzle.tpl"))

var logger = log.New(os.Stderr, "xwdweb: ", log.LstdFlags)

type PuzzleServer struct {
	puzzleRoot string
	upstream   http.Handler
}

func NewPuzzleServer(puzzleRoot string) *PuzzleServer {
	p := &PuzzleServer{puzzleRoot: puzzleRoot}
	p.upstream = http.FileServer(http.Dir(puzzleRoot))
	return p
}

func (p *PuzzleServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".puz") {
		p.upstream.ServeHTTP(w, r)
	}

	puz, err := loadPuzzle(path.Join(p.puzzleRoot, r.URL.Path))
	if err != nil {
		w.Write([]byte("Puzzle failed to load: " + err.Error()))
		return
	}

	t.ExecuteTemplate(w, "puzzle.tpl", puz)
}

func loadPuzzle(path string) (*xwd.Puzzle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	puz := &xwd.Puzzle{}
	err = puz.Load(data)
	if err != nil {
		return nil, err
	}
	return puz, nil
}

func usage() {
	fmt.Fprintf(
		os.Stderr,
		"Usage: %s [options] <puzzlesdir>\n",
		path.Base(os.Args[0]),
	)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("you must supply a directory from which to serve puzzles")
	}

	puzzles := flag.Arg(0)
	server := NewPuzzleServer(puzzles)

	http.Handle("/", server)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}
	host := os.Getenv("HOST")

	logger.Fatalln(http.ListenAndServe(host+":"+port, nil))
}
