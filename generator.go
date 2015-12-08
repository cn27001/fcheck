package fcheck

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//Generator represents a file system walker that generates meta db of files it sees on the system
type Generator struct {
	numWorker int
	FileInfoWriter
	excludes []string
	sem      chan int
	verbose  bool
}

//NewGenerator returns new Generator instance backed by the DB in dbfname
func NewGenerator(dbfname string, num int, verbose bool) *Generator {
	if err := os.Remove(dbfname); err != nil {
		log.Printf("Trouble removing old db file %s: %s\n", dbfname, err)
	}
	return &Generator{
		numWorker:      num,
		FileInfoWriter: NewDBWriter(dbfname),
		verbose:        verbose}
}

//StartWalking starts the actual walking of the filesystem to generate the DB
func (g *Generator) StartWalking(path string, exclude StringSet) error {
	g.excludes = exclude.Items()
	return filepath.Walk(path, g.Walk)
}

//Walk is the implemention of filepath.WalkFunc meant to be passed to filepath.Walk
func (g *Generator) Walk(path string, info os.FileInfo, err error) error {
	for _, v := range g.excludes {
		if strings.HasPrefix(path, v) {
			//path is excluded
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
	}
	if err != nil {
		log.Printf("Trouble in Generator.Walk: %s\n", err)
		return nil
	}
	g.sem <- 1
	if g.verbose && info.IsDir() {
		fmt.Printf("Entering %s\n", path)
	}
	go func() {
		defer func() { <-g.sem }()
		fc := &FileCheckInfo{
			Path:    path,
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
		}
		g.saveFc(fc)
	}()
	return nil
}

func (g *Generator) saveFc(fc *FileCheckInfo) {
	if err := fc.CalcDigest(); err != nil {
		log.Printf("Trouble calculating digest %s: %s\n", fc.Path, err)
	}
	err := g.Put(fc)
	if err != nil {
		log.Printf("Trouble with Set %s: %s\n", fc.Path, err)
	}
}

//Start initializes generator before walking (e.g. start workers, open DB)
func (g *Generator) Start() error {
	g.sem = make(chan int, g.numWorker)
	return g.FileInfoWriter.Start()
}

//Stop cleans up after generator finished walking (e.g. wait for pending operation, close DB)
func (g *Generator) Stop() error {
	//wait for workers to finish
	for i := 0; i < g.numWorker; i++ {
		g.sem <- 1
	}
	return g.FileInfoWriter.Stop()
}
