package fcheck

import (
	"bytes"
	"os"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FileCheckInfoSuite struct{}

var _ = Suite(&FileCheckInfoSuite{})

func (s *FileCheckInfoSuite) TestBasicFileCheckInfo(c *C) {
	now := time.Now()
	fc := FileCheckInfo{
		Path:    "/made/up",
		Size:    13,
		Mode:    os.ModeDevice,
		ModTime: now,
		Digest:  []byte("somesuch"),
	}
	data, err := fc.MarshalBinary()
	c.Assert(err, IsNil)
	rfc := &FileCheckInfo{}
	err = rfc.UnmarshalBinary(data)
	c.Assert(err, IsNil)
	c.Assert(rfc.Path, Equals, fc.Path)
	c.Assert(rfc.Size, Equals, fc.Size)
	c.Assert(rfc.Size, Equals, int64(13))
	c.Assert(rfc.Mode&os.ModeDevice, Equals, os.ModeDevice)
	c.Assert(rfc.ModTime, Equals, now)
	c.Assert(bytes.Equal(rfc.Digest, fc.Digest), Equals, true)
}

func (s *FileCheckInfoSuite) TestMatching(c *C) {
	fi, err := os.Lstat("/bin/ls")
	c.Assert(err, IsNil)
	fc := FileCheckInfo{
		Path:    "/bin/ls",
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	fc2 := FileCheckInfo{
		Path:    "/bin/ls",
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
	c.Assert(fc.Match(&fc2), Equals, true)
	fc.Digest = []byte("boo")
	c.Assert(fc.Match(&fc2), Equals, false)
	fc.CalcDigest()
	fc2.CalcDigest()
	c.Assert(len(fc2.HexDigest()) > 10, Equals, true) // "expected hexdigest to be more then 10 hex chars")
	c.Assert(fc.Match(&fc2), Equals, true)            // "expected matching of FileCheckInfos")
	fc2.Size = 1
	c.Assert(fc.Match(&fc2), Equals, false) //, "expected non matching of FileCheckInfos")
	fc2.Size = fc.Size
	c.Assert(fc.Match(&fc2), Equals, true) // "expected matching of FileCheckInfos")
	fc.Mode = fc.Mode | os.ModeDevice
	c.Assert(fc.Match(&fc2), Equals, false)
	fc.Mode = fc2.Mode
	c.Assert(fc.Match(&fc2), Equals, true)
	fc.ModTime = time.Now()
	c.Assert(fc.Match(&fc2), Equals, false)
	fc.ModTime = fc2.ModTime
	c.Assert(fc.Match(&fc2), Equals, true)
	//if not regular file size does not matter
	fc.Mode = fc.Mode | os.ModeDevice
	fc2.Mode = fc.Mode
	fc.Size = 2
	c.Assert(fc.Match(&fc2), Equals, true)
	fc.Size = fc2.Size
	c.Assert(fc.Match(&fc2), Equals, true)
	//if not regular file digest does not matter
	fc.Digest = []byte("boo")
	c.Assert(fc.Match(&fc2), Equals, true)
}
