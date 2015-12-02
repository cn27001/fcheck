package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/jlabath/fcheck"
)

const (
	dbfile  = "fcheck.db"
	version = "0.2 (Nov 2015)"
)

func main() {
	var (
		generateDB = flag.Bool("gendb", false, "generates the db")
		pathPtr    = flag.String("path", "/", "path to check/generate db for")
		showPtr    = flag.Bool("show", false, "show entries that start with provided path")
		cpuPtr     = flag.String("num", "runtime.NumCPU()", "How many goroutines to run when computing checksums")
		excludePtr = flag.String("exclude_from", "excludes.txt", "File which contains path prefixes to ignore")
		verbosePtr = flag.Bool("v", false, "verbose mode")
		walker     fcheck.Walker
	)

	flag.Parse()

	askedCPU, err := strconv.Atoi(*cpuPtr)
	if err != nil || askedCPU < 1 {
		askedCPU = runtime.NumCPU()
	}

	log.Printf("fcheck %s\n", version)
	switch {
	case *showPtr:
		walker = fcheck.NewPrinter(dbfile)
	case *generateDB:
		walker = fcheck.NewGenerator(dbfile, askedCPU, *verbosePtr)
	default:
		walker = fcheck.NewComparator(dbfile, askedCPU, *verbosePtr)
	}
	if err := walker.Start(); err != nil {
		log.Fatalf("Unable to start fs walker due to %s", err.Error())
	}
	defer func() {
		log.Println("finished")
	}()
	defer walker.Stop()
	walker.StartWalking(*pathPtr, makeExcludeList(*excludePtr))
}

func makeExcludeList(path string) (e fcheck.StringSet) {
	e = make(fcheck.StringSet)
	if path == "" {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		e.Add(strings.TrimSpace(scanner.Text()))
	}
	return
}
