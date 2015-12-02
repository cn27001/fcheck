package fcheck

import (
	"fmt"
	"io"
	"os"
	"strings"
)

//Printer represents a DB only walker that displays entries in previously generated db (flag show)
type Printer struct {
	FileInfoReader
	console io.Writer
}

//NewPrinter returns new Printer instance backed by the DB in dbfname
func NewPrinter(dbfname string) *Printer {
	return &Printer{NewDBReader(dbfname), os.Stdout}
}

//StartWalking does the actual display of requested (flag -path) it respect excludes (flag -exclude_from)
func (r *Printer) StartWalking(path string, exclude StringSet) error {
	const layout = "2006-01-02 15:04:05 (MST)"
	excludes := exclude.Items()
	return r.Map(path, func(fc *FileCheckInfo) error {
		for _, v := range excludes {
			if strings.HasPrefix(fc.Path, v) {
				//path is excluded
				return nil
			}
		}
		fmt.Fprintf(r.console, "%s %s %s %s\n", fc.Mode.String(), fc.ModTime.Format(layout), fc.HexDigest(), fc.Path)
		return nil
	})
}
