package fcheck

import (
	"os"
	"testing"
	"time"
)

func TestBasicFileCheckInfo(t *testing.T) {
	now := time.Now()
	fc := FileCheckInfo{
		Path:    "/bin/smas/x/p/skot/perhaps/ls",
		Size:    13,
		Mode:    os.ModeDevice,
		ModTime: now,
		Digest:  []byte("somesuch"),
	}
	data, err := fc.MarshalBinary()
	ok(t, err)
	rfc := FileCheckInfo{}
	err = rfc.UnmarshalBinary(data)
	ok(t, err)
	equals(t, fc.Path, rfc.Path)
	equals(t, fc.Size, rfc.Size)
	assert(t, rfc.Size == 13, "expected 13 for size")
	assert(t, rfc.Mode&os.ModeDevice == os.ModeDevice, "expected os.ModeDevice on file")
	equals(t, now, rfc.ModTime)
	assert(t, rfc.ModTime.Equal(time.Now()) || rfc.ModTime.Before(time.Now()), "expected mod time to be less or equal than time now")
	assert(t, rfc.ModTime.After(time.Now().Add(-1*time.Hour)), "expected mod time after one hour ago")
	equals(t, []byte("somesuch"), rfc.Digest)
}

func TestMatching(t *testing.T) {
	fi, err := os.Lstat("/bin/ls")
	ok(t, err)
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
	assert(t, fc.Match(&fc2), "expected matching of FileCheckInfos")
	fc.Digest = []byte("boo")
	assert(t, fc.Match(&fc2) == false, "expected non matching of FileCheckInfos")
	fc.CalcDigest()
	fc2.CalcDigest()
	assert(t, len(fc2.HexDigest()) > 10, "expected hexdigest to be more then 10 hex chars")
	assert(t, fc.Match(&fc2), "expected matching of FileCheckInfos")
	fc2.Size = 1
	assert(t, fc.Match(&fc2) == false, "expected non matching of FileCheckInfos")
	fc2.Size = fc.Size
	assert(t, fc.Match(&fc2), "expected matching of FileCheckInfos")
	fc.Mode = fc.Mode | os.ModeDevice
	assert(t, fc.Match(&fc2) == false, "expected non matching of FileCheckInfos")
	fc.Mode = fc2.Mode
	assert(t, fc.Match(&fc2), "expected matching of FileCheckInfos")
	fc.ModTime = time.Now()
	assert(t, fc.Match(&fc2) == false, "expected non matching of FileCheckInfos")
	fc.ModTime = fc2.ModTime
	assert(t, fc.Match(&fc2), "expected matching of FileCheckInfos")
	//if not regular file size does not matter
	fc.Mode = fc.Mode | os.ModeDevice
	fc2.Mode = fc.Mode
	fc.Size = 2
	assert(t, fc.Match(&fc2), "expected matching of FileCheckInfos")
	fc.Size = fc2.Size
	//if not regular file digest does not matter
	fc.Digest = []byte("boo")
	assert(t, fc.Match(&fc2), "expected matching of FileCheckInfos")
}
