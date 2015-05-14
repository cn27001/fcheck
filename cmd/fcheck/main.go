package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/jlabath/fcheck"
)

const (
	dbfile  = "fcheck.db"
	version = "0.1 (May 2015)"
)

func main() {
	var generateDB = flag.Bool("gendb", false, "generates the db")
	var pathPtr = flag.String("path", "/", "path to check/generate db for")
	var walker fcheck.Walker
	flag.Parse()
	if *generateDB {
		walker = fcheck.NewGenerator(dbfile)
	} else {
		walker = fcheck.NewComparator(dbfile)
	}
	if err := walker.Start(); err != nil {
		log.Fatalf("Unable to start fs walker due to %s", err.Error())
	}
	defer walker.Stop()
	fmt.Printf("fcheck %s\n", version)
	filepath.Walk(*pathPtr, walker.Walk)
}
