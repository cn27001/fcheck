package fcheck

import (
	"bytes"
	"testing"
)

func TestPathSplit(t *testing.T) {
	parts := splitPath("bar")
	equals(t, len(parts), 1)
	equals(t, parts[0], "bar")
	parts = splitPath("bar/shoe")
	equals(t, len(parts), 2)
	equals(t, parts[1], "shoe")
	parts = splitPath("/bar/shoe")
	equals(t, len(parts), 3)
	equals(t, "", parts[0])
	equals(t, "bar", parts[1])
	parts = splitPath("/bar/shoe/")
	equals(t, len(parts), 3)
	equals(t, "", parts[0])
	equals(t, "shoe", parts[2])
}

func TestBasic(t *testing.T) {
	pi := NewPathIndex()
	equals(t, int64(1), pi.Size())
	pi.Set("/foo", 2)
	equals(t, int64(2), pi.Size())
	pi.Set("/bar", 3)
	equals(t, pi.Size(), int64(3))
	pi.Set("/bar/shoe", 4)
	equals(t, pi.Size(), int64(4))
	pi.Set("/bar/shoe/top/up/high/stuff", 11)
	equals(t, int64(8), pi.Size())
	v, getok := pi.Get("/bar")
	equals(t, getok, true)
	equals(t, v, int64(3))
	equals(t, getok, true)
	v, getok = pi.Get("/bar/shoe/top/up/high/stuff")
	equals(t, v, int64(11))
}

func TestIndexStorage(t *testing.T) {
	pi := NewPathIndex()
	pi.Set("/foo", 2)
	pi.Set("/bar", 3)
	equals(t, pi.Size(), int64(3))
	pi.Set("/bar/shoe", 4)
	equals(t, pi.Size(), int64(4))
	var buf bytes.Buffer
	pi.Save(&buf)
	idx := NewPathIndex()
	err := idx.Load(&buf)
	ok(t, err)
	equals(t, idx.Size(), int64(4))
	v, getok := pi.Get("/bar")
	equals(t, getok, true)
	equals(t, v, int64(3))
}

func TestIndexDots(t *testing.T) {
	pi := NewPathIndex()
	pi.Set("/foo", 2)
	pi.Set("/foo/bar.txt", 33)
	equals(t, pi.Size(), int64(3))
	v, getok := pi.Get("/foo")
	equals(t, getok, true)
	equals(t, v, int64(2))
	v, getok = pi.Get("/foo/bar.txt")
	equals(t, getok, true)
	equals(t, v, int64(33))
}
