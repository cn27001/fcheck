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

//DBWriter represents the underlying datastore that stores the actual filesystem entries
type DBWriter struct {
	dbfile   string
	wChan    chan *FileCheckInfo
	quitChan chan bool
	fout     io.WriteCloser
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
	return nil
}

//Stop performs any needed cleanup
func (r *DBWriter) Stop() error {
	numWorkers := 1
	for i := 0; i < numWorkers; i++ {
		r.quitChan <- true
	}
	if r.fout != nil {
		return r.fout.Close()
	}
	return nil
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
	written := 0
	for err == nil && written < len(data) {
		if written > 0 {
			log.Print("More than one write needed to finish! This should never happen.")
		}
		var num int
		num, err = out.Write(data[written:])
		written += num
	}
	return err
}

func decode(in io.Reader, m encoding.BinaryUnmarshaler) error {
	var bytesToRead uint16
	err := binary.Read(in, binary.LittleEndian, &bytesToRead)
	if err != nil {
		return err
	}
	buffer := make([]byte, bytesToRead)
	bytesRead := 0
	for err == nil && bytesRead < int(bytesToRead) {
		var lenBytes int
		buf := buffer[bytesRead:]
		lenBytes, err = in.Read(buf)
		bytesRead += lenBytes
	}
	if err != nil && err != io.EOF {
		return err
	}
	if bytesRead < int(bytesToRead) {
		return fmt.Errorf("Expected to read %d bytes but read %d, err is %v read so far %v", bytesToRead, bytesRead, err, buffer[0:bytesRead])
	}
	return m.UnmarshalBinary(buffer)
}

//DBReader is a simple implementation of FileInfoReader
type DBReader struct {
	dbfile string
	index  *PathIndex
	db     *os.File
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
	return nil
}

//GenerateIndex will generate in memory index for faster record seeks from DB file
func (r *DBReader) GenerateIndex() error {
	log.Println("Generating Index")
	idx := NewPathIndex()
	fi, err := os.Open(r.dbfile)
	if err != nil {
		return err
	}
	defer fi.Close()
	bif := bufio.NewReader(fi)
	in := NewPositionReader(bif)
	var (
		pos int64
		fc  FileCheckInfo
	)
	for {
		pos = in.Position()
		if err = decode(in, &fc); err != nil {
			if err != io.EOF {
				log.Printf("trouble in decode: %s\n", err.Error())
				return err
			}
			break
		}
		idx.Set(fc.Path, pos)
	}
	r.index = idx
	log.Println("Done generating Index")
	return nil
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
	bif := bufio.NewReader(fi)
	for {
		var fc FileCheckInfo
		if err = decode(bif, &fc); err != nil {
			if err != io.EOF {
				log.Printf("trouble calling decode for %s: %s\n", path, err.Error())
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

//PositionReader keeps track of how many bytes it has read so far
type PositionReader struct {
	r   io.Reader
	pos int64
}

//NewPositionReader returns new reader using r as the source reader
func NewPositionReader(r io.Reader) *PositionReader {
	return &PositionReader{
		r: r}
}

//Position returns how many bytes were read so far
func (pr *PositionReader) Position() int64 {
	return pr.pos
}

//Read implements io.Reader
func (pr *PositionReader) Read(buf []byte) (int, error) {
	bl, err := pr.r.Read(buf)
	pr.pos += int64(bl)
	return bl, err
}
