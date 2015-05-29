package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
)

const (
	dbfile  = "fcount.db"
	version = "0.1 (May 2015)"
)

type Counter struct {
	cnt *big.Int
}

func (bn *Counter) Walk(path string, fi os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
	} else {
		bn.Inc()
	}
	return nil
}

func (bn *Counter) Inc() {
	one := big.NewInt(1)
	bn.cnt.Add(bn.cnt, one)
}

func (bn *Counter) String() string {
	return bn.cnt.String()
}

func main() {
	var pathPtr = flag.String("path", "/", "path to count fs entries for")
	flag.Parse()
	fmt.Printf("fcount %s\n", version)
	cnt := &Counter{big.NewInt(0)}
	filepath.Walk(*pathPtr, cnt.Walk)
	fmt.Printf("Entries found %s\n", cnt.String())
}
