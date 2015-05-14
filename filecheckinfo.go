package fcheck

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"time"
)

//FileCheckInfo represent the FileCheckInfo structure which captures metadata about a single file on a filesystem
type FileCheckInfo struct {
	Path    string      // full path of the file
	Name    string      // base name of the file
	Size    int64       // length in bytes for regular files; system-dependent for others
	Mode    os.FileMode // file mode bits
	ModTime time.Time   // modification time
	Digest  []byte      //sha1 sum
}

//FileCheckInfoFromBytes returns new instance of FileCheckInfo after deserializing it from []byte
func FileCheckInfoFromBytes(data []byte) (*FileCheckInfo, error) {
	var fc FileCheckInfo
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&fc)
	return &fc, err
}

//CalcDigest performs a SHA1 checksum on a file in question if it's a regular file
func (fc *FileCheckInfo) CalcDigest() error {
	if !fc.Mode.IsRegular() {
		return nil
	}
	file, err := os.Open(fc.Path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()
	h := sha1.New()
	if _, err := io.Copy(h, file); err != nil {
		return err
	}
	fc.Digest = h.Sum(nil)
	return nil
}

//HexDigest returns the Digest (checksum) as hexadecimal string
func (fc *FileCheckInfo) HexDigest() string {
	if !fc.Mode.IsRegular() {
		return ""
	}
	return fmt.Sprintf("%x", fc.Digest)
}

//ToBytes serializes this instance of FileCheckInfo into a []byte
func (fc *FileCheckInfo) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(fc)
	return buf.Bytes(), err
}

//Key returns unique identifier for this FileCheckInfo (currently the full path is used)
func (fc *FileCheckInfo) Key() []byte {
	return []byte(fc.Path)
}

//Match returns true if this instance of FileCheckInfo equals the Other
//that is both the os.FileMode and (checksum if applicable have to much)
func (fc *FileCheckInfo) Match(ot *FileCheckInfo) bool {
	if ot.Mode != fc.Mode {
		return false
	} else if ot.HexDigest() != fc.HexDigest() {
		return false
	}
	return true
}
