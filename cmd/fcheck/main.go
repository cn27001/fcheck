package main

import (
	"flag"
	"log"
	"runtime"
	"strconv"

	"github.com/jlabath/fcheck"
)

const (
	dbfile  = "fcheck.db"
	version = "0.1 (May 2015)"
)

func main() {
	var (
		generateDB = flag.Bool("gendb", false, "generates the db")
		pathPtr    = flag.String("path", "/", "path to check/generate db for")
		showPtr    = flag.Bool("show", false, "show entries that start with provided path")
		cpuPtr     = flag.String("numcpu", "auto", "How many system threads can be executed at the time; default is number of CPUs")
		walker     fcheck.Walker
	)

	flag.Parse()

	//how many concurrent system threads to run
	numCPU := runtime.NumCPU()
	askedCPU, err := strconv.Atoi(*cpuPtr)
	if err != nil || askedCPU > numCPU || askedCPU < 1 {
		askedCPU = numCPU
	}
	runtime.GOMAXPROCS(askedCPU)

	log.Printf("fcheck %s\n", version)
	switch {
	case *showPtr:
		walker = fcheck.NewPrinter(dbfile)
	case *generateDB:
		walker = fcheck.NewGenerator(dbfile)
	default:
		walker = fcheck.NewComparator(dbfile)
	}
	if err := walker.Start(); err != nil {
		log.Fatalf("Unable to start fs walker due to %s", err.Error())
	}
	defer func() {
		log.Println("finished")
	}()
	defer walker.Stop()
	walker.StartWalking(*pathPtr)
}
