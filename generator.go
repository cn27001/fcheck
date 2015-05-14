package fcheck

import (
	"log"
	"os"
)

//Generator represents a file system walker that generates meta db of files it sees on the system
type Generator struct {
	*DB
}

//NewGenerator returns new Generator instance backed by the DB in dbfname
func NewGenerator(dbfname string) *Generator {
	if err := os.Remove(dbfname); err != nil {
		log.Printf("Trouble removing old db file %s: %s\n", dbfname, err)
	}
	return &Generator{NewDB(dbfname)}
}

//Walk is the implemention of filepath.WalkFunc meant to be passed to filepath.Walk
func (g *Generator) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Printf("Trouble in Generator.Walk: %s\n", err)
		return nil
	}
	fc := &FileCheckInfo{
		Path:    path,
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
	}
	return g.saveFc(fc)
}

func (g *Generator) saveFc(fc *FileCheckInfo) error {
	if err := fc.CalcDigest(); err != nil {
		log.Printf("Trouble calculating digest: %s\n", err)
		return err
	}
	key := fc.Key()
	val, err := fc.ToBytes()
	if err != nil {
		log.Printf("Trouble with ToBytes: %s\n", err)
		return err
	}
	err = g.Set(key, val)
	if err != nil {
		log.Printf("Trouble with Set: %s\n", err)
		return err
	}
	return nil
}
