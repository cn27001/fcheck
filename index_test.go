package fcheck

import (
	"bytes"
	. "gopkg.in/check.v1"
)

type IndexSuite struct{}

var _ = Suite(&IndexSuite{})

func (s *IndexSuite) TestPathSplit(c *C) {
	parts := splitPath("bar")
	c.Assert(parts, HasLen, 1)
	c.Assert(parts[0], Equals, "bar")
	parts = splitPath("bar/shoe")
	c.Assert(parts, HasLen, 2)
	c.Assert(parts[1], Equals, "shoe")
	parts = splitPath("/bar/shoe")
	c.Assert(parts, HasLen, 3)
	c.Assert(parts[0], Equals, "")
	c.Assert(parts[1], Equals, "bar")
	parts = splitPath("/bar/shoe/")
	c.Assert(parts, HasLen, 3)
	c.Assert(parts[0], Equals, "")
	c.Assert(parts[1], Equals, "bar")
	c.Assert(parts[2], Equals, "shoe")
}

func (s *IndexSuite) TestBasic(c *C) {
	pi := NewPathIndex()
	c.Assert(pi.Size(), Equals, int64(1))
	pi.Set("/foo", 2)
	c.Assert(pi.Size(), Equals, int64(2))
	pi.Set("/bar", 3)
	c.Assert(pi.Size(), Equals, int64(3))
	pi.Set("/bar/shoe", 4)
	c.Assert(pi.Size(), Equals, int64(4))
	pi.Set("/bar/shoe/top/up/high/stuff", 11)
	c.Assert(pi.Size(), Equals, int64(8))
	v, getok := pi.Get("/bar")
	c.Assert(getok, Equals, true)
	c.Assert(v, Equals, int64(3))
	v, getok = pi.Get("/bar/shoe/top/up/high/stuff")
	c.Assert(getok, Equals, true)
	c.Assert(v, Equals, int64(11))
}

func (s *IndexSuite) TestIndexStorage(c *C) {
	pi := NewPathIndex()
	pi.Set("/foo", 2)
	pi.Set("/bar", 3)
	c.Assert(pi.Size(), Equals, int64(3))
	pi.Set("/bar/shoe", 4)
	c.Assert(pi.Size(), Equals, int64(4))
	var buf bytes.Buffer
	pi.Save(&buf)
	idx := NewPathIndex()
	err := idx.Load(&buf)
	c.Assert(err, IsNil)
	c.Assert(idx.Size(), Equals, int64(4))
	v, getok := pi.Get("/bar")
	c.Assert(getok, Equals, true)
	c.Assert(v, Equals, int64(3))
}

func (s *IndexSuite) TestIndexDots(c *C) {
	pi := NewPathIndex()
	pi.Set("/foo", 2)
	pi.Set("/foo/bar.txt", 33)
	c.Assert(pi.Size(), Equals, int64(3))
	v, getok := pi.Get("/foo")
	c.Assert(getok, Equals, true)
	c.Assert(v, Equals, int64(2))
	v, getok = pi.Get("/foo/bar.txt")
	c.Assert(getok, Equals, true)
	c.Assert(v, Equals, int64(33))
}
