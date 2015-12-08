package fcheck

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"sync"

	. "gopkg.in/check.v1"
)

type TestSuite struct {
	testDBName string
	testPath   string
}

var _ = Suite(&TestSuite{
	testDBName: "fcheck_test.db",
	testPath:   "/bin"})

//TearDownTest is called after each test completes
func (s *TestSuite) TearDownSuite(c *C) {
	os.Remove(s.testDBName)
}

func (s *TestSuite) Test1Generator(c *C) {
	var g Walker = NewGenerator(s.testDBName, 2, false)
	exclude := make(StringSet)
	err := g.Start()
	c.Assert(err, IsNil)
	err = g.StartWalking(s.testPath, exclude)
	c.Assert(err, IsNil)
	err = g.Stop()
	c.Assert(err, IsNil)
}

func (s *TestSuite) Test2Printer(c *C) {
	var p Walker = NewPrinter(s.testDBName)
	exclude := make(StringSet)
	exclude.Add("/bin/ps")
	var buf bytes.Buffer
	err := p.Start()
	c.Assert(err, IsNil)
	rawp := p.(*Printer)
	rawp.console = &buf
	err = p.StartWalking(s.testPath, exclude)
	c.Assert(err, IsNil)
	err = p.Stop()
	c.Assert(err, IsNil)
	//examine buffer
	foundLS := false
	foundPS := false
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		if strings.Index(scanner.Text(), "/bin/ls") > -1 {
			foundLS = true
		}
		if strings.Index(scanner.Text(), "/bin/ps") > -1 {
			foundPS = true
		}
	}
	c.Assert(foundLS, Equals, true)  //"expected to found /bin/ls")
	c.Assert(foundPS, Equals, false) //"expected to NOT found /bin/ps")
}

func (s *TestSuite) Test3Comparator(c *C) {
	var cm Walker = NewComparator(s.testDBName, 2, false)
	rawcm := cm.(*Comparator)
	var buf bytes.Buffer
	rawcm.console = &buf
	exclude := make(StringSet)
	err := cm.Start()
	c.Assert(err, IsNil)
	err = cm.StartWalking(s.testPath, exclude)
	c.Assert(err, IsNil)
	err = cm.Stop()
	c.Assert(err, IsNil)
	c.Assert(rawcm.newFiles, HasLen, 0)
	c.Assert(rawcm.changedFiles, HasLen, 0)
	c.Assert(rawcm.removedFiles, HasLen, 0)
}

func (s *TestSuite) Test4ComparatorNoPath(c *C) {
	var cm Walker = NewComparator(s.testDBName, 2, false)
	rawcm := cm.(*Comparator)
	var buf bytes.Buffer
	rawcm.console = &buf
	exclude := make(StringSet)
	err := cm.Start()
	c.Assert(err, IsNil)
	//non-exist path
	err = cm.StartWalking("/foobardubar23256646", exclude)
	c.Assert(err, IsNil)
	err = cm.Stop()
	c.Assert(err, IsNil)
	c.Assert(rawcm.newFiles, HasLen, 0)
	c.Assert(rawcm.changedFiles, HasLen, 0)
	c.Assert(rawcm.removedFiles, HasLen, 0)
}

func (s *TestSuite) Test5ComparatorNoPathInDB(c *C) {
	var cm Walker = NewComparator(s.testDBName, 2, false)
	rawcm := cm.(*Comparator)
	var buf bytes.Buffer
	rawcm.console = &buf
	exclude := make(StringSet)
	err := cm.Start()
	c.Assert(err, IsNil)
	//permission errors as ordinary user plus new files
	err = cm.StartWalking("/etc", exclude)
	c.Assert(err, IsNil)
	err = cm.Stop()
	c.Assert(err, IsNil)
	c.Assert(len(rawcm.newFiles) > 0, Equals, true) // "expected to be all new files in /etc")
	c.Assert(rawcm.changedFiles, HasLen, 0)
	c.Assert(rawcm.removedFiles, HasLen, 0)
}

func (s *TestSuite) Test6PrinterNoPath(c *C) {
	var cm Walker = NewPrinter(s.testDBName)
	exclude := make(StringSet)
	rawcm := cm.(*Printer)
	var buf bytes.Buffer
	rawcm.console = &buf
	err := cm.Start()
	c.Assert(err, IsNil)
	//non-exist path
	err = cm.StartWalking("/foobardubar23256646", exclude)
	c.Assert(err, IsNil)
	err = cm.Stop()
	c.Assert(err, IsNil)
}

func (s *TestSuite) Test7Get(c *C) {
	d := NewDBReader(s.testDBName)
	err := d.Start()
	c.Assert(err, IsNil)
	err = d.GenerateIndex()
	c.Assert(err, IsNil)
	fi, err := d.Get("/skart/12412415145464634633463464")
	c.Assert(err, NotNil)
	c.Assert(err, Equals, ErrNotFound)
	//assume /bin/ls exists !
	p := "/bin/ls"
	fi, err = d.Get(p)
	c.Assert(err, IsNil)
	c.Assert(fi, NotNil)
	//many concurrent gets
	var wg sync.WaitGroup
	mgf := func(even bool) {
		p := p
		if even {
			p = "/bin/ps"
		}
		for i := 0; i < 1000; i++ {
			fi, err := d.Get(p)
			c.Assert(err, IsNil)
			c.Assert(fi, NotNil)
			c.Assert(p, Equals, fi.Path)
		}
		wg.Done()
	}
	for i := 0; i < 500; i++ {
		wg.Add(1)
		even := i%2 == 0
		go mgf(even)
	}
	wg.Wait()
	err = d.Stop()
	c.Assert(err, IsNil)
}
