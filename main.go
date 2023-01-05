package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
)

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var funcflag = flag.String("func", "", "Function to select")
var phaseflag = flag.String("phase", "", "Phase to select")
var infileflag = flag.String("i", "", "Input file")
var outfileflag = flag.String("o", "", "Output file")

func verb(vlevel int, s string, a ...interface{}) {
	if *verbflag >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: split-go-ast [flags] [inputs]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

type scanState int

const (
	betweenFuncs scanState = iota
	inSelectedFunc
	inNonSelectedFunc
)

func (s scanState) String() string {
	switch s {
	case betweenFuncs:
		return "between"
	case inSelectedFunc:
		return "select"
	case inNonSelectedFunc:
		return "nonselect"
	default:
		return fmt.Sprintf("scanState[%d]", s)
	}
}

func perform(inf, outf *os.File) {
	verb(1, "in perform")
	markRE := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(.+)\s*$`)
	scanner := bufio.NewScanner(inf)
	state := betweenFuncs
	for scanner.Scan() {
		line := scanner.Text()
		verb(3, "state %s line is %q", state.String(), line)
		switch state {
		case betweenFuncs:
			m := markRE.FindStringSubmatch(line)
			if len(m) != 0 && (m[1] == "before" || m[1] == "after") {
				verb(2, "marker %s", line)
				state = inSelectedFunc
				if *phaseflag != "" && m[2] != *phaseflag {
					state = inNonSelectedFunc
				}
				if *funcflag != "" && m[3] != *funcflag {
					state = inNonSelectedFunc
				}
			}
			if state == inSelectedFunc {
				fmt.Fprintf(outf, "%s\n", line)
			}
		case inNonSelectedFunc:
			if line == "" {
				state = betweenFuncs
			}
		case inSelectedFunc:
			fmt.Fprintf(outf, "%s\n", line)
			if line == "" {
				state = betweenFuncs
			}
		}
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("split-go-ast: ")
	flag.Parse()
	verb(1, "in main")
	if flag.NArg() != 0 {
		usage("unknown extra args")
	}
	var err error
	var infile *os.File = os.Stdin
	if len(*infileflag) > 0 {
		verb(1, "opening %s", *infileflag)
		infile, err = os.Open(*infileflag)
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()
	}
	var outfile *os.File = os.Stdout
	if len(*outfileflag) > 0 {
		verb(1, "opening %s", *outfileflag)
		outfile, err = os.OpenFile(*outfileflag, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := outfile.Close(); err != nil {
				log.Fatalf("closing %s: error %v", *outfileflag, err)
			}
		}()
	}
	perform(infile, outfile)
	verb(1, "leaving main")
}
