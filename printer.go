package fcheck

import "fmt"

//Comparator represents a file system walker that checks previously generated db and compares it to a filesystem
type Printer struct {
	FileInfoReader
}

//NewComparator returns new Comparator instance backed by the DB in dbfname
func NewPrinter(dbfname string) *Printer {
	return &Printer{NewDBReader(dbfname)}
}

func (r *Printer) StartWalking(path string) error {
	const layout = "2006-01-02 15:04:05 (MST)"
	return r.Map(path, func(fc *FileCheckInfo) error {
		fmt.Printf("%s\t%s\t%s\t%s\n", fc.Mode.String(), fc.ModTime.Format(layout), fc.HexDigest(), fc.Path)
		return nil
	})
}
