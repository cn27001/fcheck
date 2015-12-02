package fcheck

import (
	"bufio"
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

//BufferSize default size of I/O buffers
const BufferSize = 512 * 1024

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

//Put puts an entry in the datastore
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
			if err := encode(r.fout, fc); err != nil {
				log.Print("trouble writing to db file: ", err.Error())
			}
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
		}
		return fmt.Errorf("Expected to read %d bytes but read %d", blen, n)
	}
	return m.UnmarshalBinary(buf)
}

//DBReader is a simple implementation of FileInfoReader
type DBReader struct {
	dbfile string
	index  *PathIndex
	db     PositionalReadCloser
	l      sync.Mutex
}

//ErrNotFound signifies that such FileCheckInfo entry could not be find
var ErrNotFound = errors.New("not found")

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

//Get retrieves an entry from db
func (r *DBReader) Get(key string) (*FileCheckInfo, error) {
	//index reading can be concurrent
	offset, ok := r.index.Get(key)
	if !ok {
		return nil, ErrNotFound
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

//Map maps FileCheckInfo entries in db whose paths match path to DBMapFunc f
func (r *DBReader) Map(path string, f DBMapFunc) error {
	fi, err := os.Open(r.dbfile)
	if err != nil {
		return err
	}
	defer fi.Close()
	bif := bufio.NewReaderSize(fi, BufferSize)
	for {
		var fc FileCheckInfo
		if err = decode(bif, &fc); err != nil {
			if err != io.EOF {
				log.Print("trouble calling decode:", err)
				return err
			}
			break
		}
		if !strings.HasPrefix(fc.Path, path) {
			continue
		}
		f(&fc)
	}
	return nil
}
