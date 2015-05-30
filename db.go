package fcheck

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

//DBWriter represents the underlying datastore that stores the actual filesystem entries
type DBWriter struct {
	dbfile   string
	wChan    chan *FileCheckInfo
	quitChan chan bool
	fout     PositionalWriteCloser
	index    *PathIndex
}

//NewDBWriter returns new instance of DBWriter
func NewDBWriter(dbfname string) *DBWriter {
	return &DBWriter{dbfile: dbfname}
}

//Start performs any needed initialization
func (r *DBWriter) Start() error {
	f, err := os.Create(r.dbfile)
	if err != nil {
		return err
	}
	//make channel
	r.wChan = make(chan *FileCheckInfo)
	r.quitChan = make(chan bool)
	r.fout = f
	go r.writer()
	r.index = NewPathIndex()
	return nil
}

//Stop performs any needed cleanup
func (r *DBWriter) Stop() error {
	numWorkers := 1
	for i := 0; i < numWorkers; i++ {
		r.quitChan <- true
	}
	if r.fout != nil {
		r.fout.Close() //ignore error
	}
	f, err := os.Create(r.dbfile + ".index")
	if err != nil {
		return err
	}
	defer f.Close()
	return r.index.Save(f)
}

//Set puts an entry in the datastore
func (r *DBWriter) Put(fc *FileCheckInfo) error {
	r.wChan <- fc
	return nil
}

func (r *DBWriter) writer() {
	for {
		select {
		case fc := <-r.wChan:
			curPos, err := r.fout.Seek(0, os.SEEK_CUR)
			if err != nil {
				log.Fatal(err)
			}
			r.index.Set(fc.Path, curPos)
			encode(r.fout, fc)
		case <-r.quitChan:
			return
		}
	}
}

func encode(out io.Writer, m encoding.BinaryMarshaler) error {
	data, err := m.MarshalBinary()
	if err != nil {
		return err
	}
	blen := uint16(len(data))
	err = binary.Write(out, binary.LittleEndian, blen)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

func decode(in io.Reader, m encoding.BinaryUnmarshaler) error {
	var blen uint16
	err := binary.Read(in, binary.LittleEndian, &blen)
	if err != nil {
		return err
	}
	buf := make([]byte, blen)
	if n, err := in.Read(buf); n != len(buf) {
		if err != nil {
			return err
		} else {
			return fmt.Errorf("Expected to read %d bytes but read %d", blen, n)
		}
	}
	return m.UnmarshalBinary(buf)
}

type DBReader struct {
	dbfile string
	index  *PathIndex
	db     PositionalReadCloser
	l      sync.Mutex
}

var NotFoundErr = errors.New("not found")

//NewDBReader returns new instance of DBReader
func NewDBReader(dbfname string) *DBReader {
	return &DBReader{dbfile: dbfname}
}

//Start performs any needed initialization
func (r *DBReader) Start() error {
	rs, err := os.Open(r.dbfile)
	if err != nil {
		return err
	}
	r.db = rs
	f, err := os.Open(r.dbfile + ".index")
	if err != nil {
		return err
	}
	defer f.Close()
	r.index = NewPathIndex()
	err = r.index.Load(f)
	return err
}

//Stop performs any needed cleanup
func (r *DBReader) Stop() error {
	return r.db.Close()
}

//Set puts an entry in the datastore
func (r *DBReader) Get(key string) (*FileCheckInfo, error) {
	//index reading can be concurrent
	offset, ok := r.index.Get(key)
	if !ok {
		return nil, NotFoundErr
	}
	//lock db file
	r.l.Lock()
	defer r.l.Unlock()
	var fc FileCheckInfo
	//seek to where our record is at
	_, err := r.db.Seek(offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	//actual read
	err = decode(r.db, &fc)
	if fc.Path != key {
		log.Fatalf("Something went terribly wrong key(%s) does not equal path(%s) at %d", key, fc.Path, offset)
	}
	return &fc, err
}

func (r *DBReader) Map(path string, f DBMapFunc) error {
	fi, err := os.Open(r.dbfile)
	if err != nil {
		return err
	}
	defer fi.Close()
	//ok traverse index to get min and max offset of records in db file
	var (
		maxPos, minPos int64
		node           *PEntry
		ok             bool
	)
	node, ok = r.index.GetNode(path)
	if !ok {
		return nil //error if not found ?
	}
	node.Traverse(func(x *PEntry) {
		if x.Pos < minPos {
			minPos = x.Pos
		}
		if x.Pos > maxPos {
			maxPos = x.Pos
		}
	})
	//seek to min pos
	if _, serr := fi.Seek(minPos, os.SEEK_SET); serr != nil {
		return serr
	}
	for {
		var fc FileCheckInfo
		if err := decode(fi, &fc); err != nil {
			break
		}
		if !strings.HasPrefix(fc.Path, path) {
			continue
		}
		f(&fc)
		curPos, err := fi.Seek(0, os.SEEK_CUR)
		if err != nil {
			break
		}
		if maxPos > 0 && maxPos < curPos {
			break
		}
	}
	return err
}
