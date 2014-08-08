# xwd

A library for interacting with crossword puzzles, and some higher-level
utilities (for the commandline and the web) for interacting with them.

## installing

    go get github.com/nickstenning/xwd/...

## using

At the moment, the only supported puzzle format is `.puz` (AKA AcrossLite)
format:

    xwd foo.puz

Or, to serve a directory tree of puzzles on the web:

    cd $GOPATH/src/github.com/nickstenning/xwd/xwdweb
    xwdweb ~/puzzles

(Yes, I need to bundle the templates at some point.)

A couple of example puzzles can be found in the `fixtures/` directory

## caveats

The web interface doesn't do anything useful yet. It's also really ugly.
