package fcheck

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//Comparator represents a file system walker that checks previously generated db and compares it to a filesystem
type Comparator struct {
	FileInfoReader
	newFiles     []string
	changedFiles []string
	removedFiles []string
	pathWalked   string
	taskCh       chan *FileCheckInfo
	quitCh       chan bool
	doneCh       chan bool
	changedCh    chan string
	newCh        chan string
	numWorkers   int
	console      io.Writer
	excludes     []string
}

//NewComparator returns new Comparator instance backed by the DB in dbfname
func NewComparator(dbfname string) *Comparator {
	return &Comparator{
		FileInfoReader: NewDBReader(dbfname),
		numWorkers:     runtime.NumCPU(),
		console:        os.Stdout}
}

//Start initializes generator before walking (e.g. start workers, open DB)
func (rcv *Comparator) Start() error {
	rcv.taskCh = make(chan *FileCheckInfo)
	rcv.quitCh = make(chan bool)
	rcv.doneCh = make(chan bool)
	rcv.newCh = make(chan string)
	rcv.changedCh = make(chan string)
	//start the append routine
	go func() {
	FLOOP:
		for {
			select {
			case x := <-rcv.changedCh:
				rcv.changedFiles = append(rcv.changedFiles, x)
			case x := <-rcv.newCh:
				rcv.newFiles = append(rcv.newFiles, x)
			case <-rcv.doneCh:
				break FLOOP
			}
		}
		//sync with main by waiting for quit
		<-rcv.quitCh
	}()
	//start the compare workers
	for i := 0; i < rcv.numWorkers; i++ {
		go func(n int) {
			for {
				select {
				case fc := <-rcv.taskCh:
					//log.Printf("worker %d saving %s\n", n, fc.Path)
					rcv.compareFc(fc)
				case <-rcv.quitCh:
					//log.Printf("worker %d quitting\n", n)
					return
				}
			}
		}(i)
	}
	return rcv.FileInfoReader.Start()
}

//StartWalking will start the actual filesystem walking and comparison with DB
func (rcv *Comparator) StartWalking(path string, exclude StringSet) error {
	rcv.pathWalked = path
	rcv.excludes = exclude.Items()
	return filepath.Walk(path, rcv.Walk)
}

//Walk is the implemention of filepath.WalkFunc meant to be passed to filepath.Walk
func (rcv *Comparator) Walk(path string, info os.FileInfo, err error) error {
	for _, v := range rcv.excludes {
		if strings.HasPrefix(path, v) {
			//path is excluded
			return nil
		}
	}
	fc := &FileCheckInfo{
		Path: path,
	}
	if err != nil {
		log.Print(err)
		if os.IsNotExist(err) {
			return nil
		}
	} else {
		fc.Size = info.Size()
		fc.Mode = info.Mode()
		fc.ModTime = info.ModTime()
	}
	rcv.taskCh <- fc
	return nil
}

func (rcv *Comparator) compareFc(fc *FileCheckInfo) {
	old, err := rcv.Get(fc.Path)
	if err == ErrNotFound {
		old = nil
	} else if err != nil {
		log.Fatalf("Trouble with Get(\"%s\") %s\n", fc.Path, err.Error())
	}
	if old == nil {
		//ok does not exist in db stop right here
		rcv.newCh <- fc.Path
		return
	}
	//to save time only calc digest if not obviously different
	if fc.LiteMatch(old) {
		if err := fc.CalcDigest(); err != nil {
			log.Printf("Trouble calculating digest: %s\n", err.Error())
		}
	}
	if !fc.Match(old) {
		rcv.changedCh <- fc.Path
	}
}

//Stop is a wrapper around underlying DB.Stop that also prints the final report of comparison
func (rcv *Comparator) Stop() error {
	defer rcv.FileInfoReader.Stop()
	//wait for compare tasks to finish
	for i := 0; i < rcv.numWorkers; i++ {
		rcv.quitCh <- true
	}
	close(rcv.doneCh)
	//this one is for the appender routine
	rcv.quitCh <- true
	//find the deleted files
	maperror := rcv.Map(rcv.pathWalked, func(fc *FileCheckInfo) error {
		for _, v := range rcv.excludes {
			if strings.HasPrefix(fc.Path, v) {
				//path is excluded
				return nil
			}
		}
		if _, perr := os.Lstat(fc.Path); perr != nil {
			if os.IsNotExist(perr) {
				rcv.removedFiles = append(rcv.removedFiles, fc.Path)
			} else {
				log.Printf("Error in Map step: %s", perr)
			}
		}
		return nil
	})
	if maperror != nil {
		log.Printf("Error in Map: %s", maperror.Error())
		return maperror
	}
	//Print the report
	fmt.Fprintf(rcv.console, "\n\nChanged files %d\n\n", len(rcv.changedFiles))
	for _, v := range rcv.changedFiles {
		fmt.Fprintln(rcv.console, v)
	}
	fmt.Fprintf(rcv.console, "\n\nNew files %d\n\n", len(rcv.newFiles))
	for _, v := range rcv.newFiles {
		fmt.Fprintln(rcv.console, v)
	}
	fmt.Fprintf(rcv.console, "\n\nDeleted files %d\n\n", len(rcv.removedFiles))
	for _, v := range rcv.removedFiles {
		fmt.Fprintln(rcv.console, v)
	}
	return nil
}
