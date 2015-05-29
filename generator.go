package fcheck

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

//Generator represents a file system walker that generates meta db of files it sees on the system
type Generator struct {
	numWorker int
	taskCh    chan *FileCheckInfo
	quitCh    chan bool
	FileInfoWriter
}

//NewGenerator returns new Generator instance backed by the DB in dbfname
func NewGenerator(dbfname string) *Generator {
	if err := os.Remove(dbfname); err != nil {
		log.Printf("Trouble removing old db file %s: %s\n", dbfname, err)
	}
	return &Generator{runtime.NumCPU(), nil, nil, NewDBWriter(dbfname)}
}

func (rcv *Generator) StartWalking(path string) error {
	return filepath.Walk(path, rcv.Walk)
}

//Walk is the implemention of filepath.WalkFunc meant to be passed to filepath.Walk
func (g *Generator) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Printf("Trouble in Generator.Walk: %s\n", err)
		return nil
	}
	fc := &FileCheckInfo{
		Path:    path,
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
	}
	g.taskCh <- fc
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
	g.taskCh = make(chan *FileCheckInfo)
	g.quitCh = make(chan bool)
	for i := 0; i < g.numWorker; i++ {
		go func(n int) {
			for {
				select {
				case fc := <-g.taskCh:
					//log.Printf("worker %d saving %s\n", n, fc.Path)
					g.saveFc(fc)
				case <-g.quitCh:
					//log.Printf("worker %d quitting\n", n)
					return
				}
			}
		}(i)
	}
	return g.FileInfoWriter.Start()
}

//Stop cleans up after generator finished walking (e.g. wait for pending operation, close DB)
func (g *Generator) Stop() error {
	//wait for workers to finish
	for i := 0; i < g.numWorker; i++ {
		g.quitCh <- true
	}
	return g.FileInfoWriter.Stop()
}
