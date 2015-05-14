package fcheck

import (
	"fmt"
	"log"
	"os"
)

//Comparator represents a file system walker that checks previously generated db and compares it to a filesystem
type Comparator struct {
	newFiles     []string
	changedFiles []string
	removedFiles []string
	*DB
}

//NewComparator returns new Comparator instance backed by the DB in dbfname
func NewComparator(dbfname string) *Comparator {
	return &Comparator{nil, nil, nil, NewDB(dbfname)}
}

//Walk is the implemention of filepath.WalkFunc meant to be passed to filepath.Walk
func (rcv *Comparator) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Printf("Received error in Walk: %s", err.Error())
		return err
	}
	fc := &FileCheckInfo{
		Path:    path,
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
	}
	return rcv.compareFc(fc)
}

func (rcv *Comparator) compareFc(fc *FileCheckInfo) error {
	key := fc.Key()
	val, err := rcv.Get(key)
	if err != nil {
		log.Printf("Touble with Get(key) %s\n", err.Error())
		return err
	}
	if val == nil {
		//ok does not exist in db stop right here
		rcv.newFiles = append(rcv.newFiles, fc.Path)
		return nil
	}
	old, err := FileCheckInfoFromBytes(val)
	if err != nil {
		log.Printf("Touble with FileCheckInfoFromBytes: %s\n", err.Error())
		return err
	}
	if err := fc.CalcDigest(); err != nil {
		log.Printf("Trouble with CalcDigest: %s\n", err.Error())
		return err
	}
	if !fc.Match(old) {
		rcv.changedFiles = append(rcv.changedFiles, fc.Path)
	}
	return err
}

//Stop is a wrapper around underlying DB.Stop that also prints the final report of comparison
func (rcv *Comparator) Stop() error {
	defer rcv.DB.Stop()
	//find the deleted files
	var maperror error
	rcv.Map(func(key, val []byte) {
		if maperror != nil {
			return //noop
		}
		fc, err := FileCheckInfoFromBytes(val)
		if err != nil {
			maperror = err
			return
		}
		if _, perr := os.Lstat(fc.Path); perr != nil {
			rcv.removedFiles = append(rcv.removedFiles, fc.Path)
		}
	})
	if maperror != nil {
		log.Printf("Error in Map: %s", maperror.Error())
		return maperror
	}
	//Print the report
	fmt.Printf("Changed files %d\n", len(rcv.changedFiles))
	for _, v := range rcv.changedFiles {
		fmt.Println(v)
	}
	fmt.Printf("New files %d\n", len(rcv.newFiles))
	for _, v := range rcv.newFiles {
		fmt.Println(v)
	}
	fmt.Printf("Deleted files %d\n", len(rcv.removedFiles))
	for _, v := range rcv.removedFiles {
		fmt.Println(v)
	}
	return nil
}
