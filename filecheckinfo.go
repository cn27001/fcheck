package fcheck

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

var emptyDigestString = ""

//FileCheckInfo represent the FileCheckInfo structure which captures metadata about a single file on a filesystem
type FileCheckInfo struct {
	Path    string      // full path of the file
	Size    int64       // length in bytes for regular files; system-dependent for others
	Mode    os.FileMode // file mode bits
	ModTime time.Time   // modification time
	Digest  []byte      // checksum
}

//CalcDigest performs a SHA512 checksum on a file in question if it's a regular file
func (fc *FileCheckInfo) CalcDigest() error {
	if !fc.Mode.IsRegular() || fc.Size == 0 {
		//only calc regular files
		//do not calc empty (sometimes special files)
		return nil
	}
	file, err := os.Open(fc.Path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()
	h := sha512.New()
	if _, err := io.Copy(h, file); err != nil {
		return err
	}
	fc.Digest = h.Sum(nil)
	return nil
}

//HexDigest returns the Digest (checksum) as hexadecimal string
func (fc *FileCheckInfo) HexDigest() string {
	if !fc.Mode.IsRegular() {
		return emptyDigestString
	} else if len(fc.Digest) == 0 {
		return emptyDigestString
	}
	return fmt.Sprintf("%x", fc.Digest)
}

func (fc *FileCheckInfo) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	bw := &binaryWriter{}
	//encode path
	bp := []byte(fc.Path)
	blen := uint16(len(bp))
	bw.Write(&buf, blen)
	buf.Write(bp)
	//size
	bw.Write(&buf, fc.Size)
	//mode
	bw.Write(&buf, fc.Mode)
	//modtime
	sertime, err := fc.ModTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	blen = uint16(len(sertime))
	bw.Write(&buf, blen)
	buf.Write(sertime)
	//digest
	blen = uint16(len(fc.Digest))
	bw.Write(&buf, blen)
	buf.Write(fc.Digest)
	return buf.Bytes(), bw.Err()
}

func (fc *FileCheckInfo) UnmarshalBinary(data []byte) error {
	var blen uint16
	var pos, nextpos int
	var rawmode uint32
	br := &binaryReader{}
	byr := bytes.NewReader(data)
	br.Read(byr, &blen)
	pos = pos + 2 // two bytes read for unit16
	nextpos = pos + int(blen)
	fc.Path = string(br.Slice(data, pos, nextpos)) // casting to string does copy of the data []byte
	pos = nextpos
	byr.Seek(int64(pos), 0)
	//size
	br.Read(byr, &fc.Size)
	pos = pos + 8 // 8 for int64
	//mode
	br.Read(byr, &rawmode)
	fc.Mode = os.FileMode(rawmode)
	pos = pos + 4 // 4 for unit32
	//modtime
	br.Read(byr, &blen)
	pos = pos + 2 // two bytes read for unit16
	nextpos = pos + int(blen)
	if err := (&fc.ModTime).UnmarshalBinary(br.Slice(data, pos, nextpos)); err != nil {
		return err
	}
	pos = nextpos
	byr.Seek(int64(pos), 0)
	//digest
	br.Read(byr, &blen)
	pos = pos + 2 // two bytes read for unit16
	nextpos = pos + int(blen)
	fc.Digest = br.Slice(data, pos, nextpos) // TODO do i need to copy?
	return br.Err()
}

func (fc *FileCheckInfo) LiteMatch(ot *FileCheckInfo) bool {
	if ot.Mode != fc.Mode {
		return false
	}
	switch {
	case fc.Mode.IsRegular():
		return fc.Size == ot.Size && fc.ModTime.Equal(ot.ModTime)
	default:
		return fc.ModTime.Equal(ot.ModTime)
	}
}

//Match returns true if this instance of FileCheckInfo equals the Other
//that is both the os.FileMode and (checksum if applicable have to much)
func (fc *FileCheckInfo) Match(ot *FileCheckInfo) bool {
	ok := fc.LiteMatch(ot)
	if ok {
		return ot.HexDigest() == fc.HexDigest()
	}
	return ok
}

type binaryWriter struct {
	err error
}

func (r *binaryWriter) Write(w io.Writer, data interface{}) {
	if r.err != nil {
		return //noop
	}
	r.err = binary.Write(w, binary.LittleEndian, data)
}

func (r *binaryWriter) Err() error {
	return r.err
}

type binaryReader struct {
	err error
}

func (r *binaryReader) Read(rdr io.Reader, data interface{}) {
	if r.err != nil {
		return //noop
	}
	r.err = binary.Read(rdr, binary.LittleEndian, data)
}

func (r *binaryReader) Err() error {
	return r.err
}

func (r *binaryReader) Slice(data []byte, from int, to int) []byte {
	if r.err != nil {
		return []byte(nil)
	}
	if from < 0 || to > len(data) {
		r.err = fmt.Errorf("Index out of bounds when calling Slice from:%d to:%d", from, to)
		return []byte(nil)
	}
	return data[from:to]
}

func init() {
	var buf bytes.Buffer
	//128 hexachars for sha512
	for i := 0; i < 128; i++ {
		buf.WriteString(" ")
	}
	emptyDigestString = buf.String()
}
