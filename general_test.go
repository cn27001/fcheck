package fcheck

import (
	"bufio"
	"bytes"
	"strings"
	"sync"
	"testing"
)

var (
	testDBName string = "fcheck_test.db"
	testPath   string = "/bin"
)

func TestGenerator(t *testing.T) {
	var g Walker = NewGenerator(testDBName)
	exclude := make(StringSet)
	err := g.Start()
	ok(t, err)
	err = g.StartWalking(testPath, exclude)
	ok(t, err)
	err = g.Stop()
	ok(t, err)
}

func TestPrinter(t *testing.T) {
	var p Walker = NewPrinter(testDBName)
	exclude := make(StringSet)
	exclude.Add("/bin/ps")
	var buf bytes.Buffer
	err := p.Start()
	ok(t, err)
	rawp := p.(*Printer)
	rawp.console = &buf
	err = p.StartWalking(testPath, exclude)
	ok(t, err)
	err = p.Stop()
	ok(t, err)
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
	assert(t, foundLS, "expected to found /bin/ls")
	assert(t, foundPS == false, "expected to NOT found /bin/ps")
}

func TestComparator(t *testing.T) {
	var cm Walker = NewComparator(testDBName)
	rawcm := cm.(*Comparator)
	var buf bytes.Buffer
	rawcm.console = &buf
	exclude := make(StringSet)
	err := cm.Start()
	ok(t, err)
	err = cm.StartWalking(testPath, exclude)
	ok(t, err)
	err = cm.Stop()
	ok(t, err)
	equals(t, 0, len(rawcm.newFiles))
	equals(t, 0, len(rawcm.changedFiles))
	equals(t, 0, len(rawcm.removedFiles))
}

func TestComparatorNoPath(t *testing.T) {
	var cm Walker = NewComparator(testDBName)
	rawcm := cm.(*Comparator)
	var buf bytes.Buffer
	rawcm.console = &buf
	exclude := make(StringSet)
	err := cm.Start()
	ok(t, err)
	//non-exist path
	err = cm.StartWalking("/foobardubar23256646", exclude)
	ok(t, err)
	err = cm.Stop()
	ok(t, err)
	equals(t, 0, len(rawcm.newFiles))
	equals(t, 0, len(rawcm.changedFiles))
	equals(t, 0, len(rawcm.removedFiles))
}

func TestComparatorNoPathInDB(t *testing.T) {
	var cm Walker = NewComparator(testDBName)
	rawcm := cm.(*Comparator)
	var buf bytes.Buffer
	rawcm.console = &buf
	exclude := make(StringSet)
	err := cm.Start()
	ok(t, err)
	//permission errors as ordinary user plus new files
	err = cm.StartWalking("/etc", exclude)
	ok(t, err)
	err = cm.Stop()
	ok(t, err)
	assert(t, len(rawcm.newFiles) > 0, "expected to be all new files in /etc")
	equals(t, 0, len(rawcm.changedFiles))
	equals(t, 0, len(rawcm.removedFiles))
}

func TestPrinterNoPath(t *testing.T) {
	var cm Walker = NewPrinter(testDBName)
	exclude := make(StringSet)
	rawcm := cm.(*Printer)
	var buf bytes.Buffer
	rawcm.console = &buf
	err := cm.Start()
	ok(t, err)
	//non-exist path
	err = cm.StartWalking("/foobardubar23256646", exclude)
	ok(t, err)
	err = cm.Stop()
	ok(t, err)
}

func TestGet(t *testing.T) {
	d := NewDBReader(testDBName)
	err := d.Start()
	ok(t, err)
	fi, err := d.Get("/skart/12412415145464634633463464")
	assert(t, err != nil, "Expected an error")
	assert(t, err == ErrNotFound, "Expected not found on funny key")
	//assume /bin/ls exists !
	p := "/bin/ls"
	fi, err = d.Get(p)
	ok(t, err)
	assert(t, fi != nil, "expect fi not nil")
	//many concurrent gets
	var wg sync.WaitGroup
	mgf := func(even bool) {
		p := p
		if even {
			p = "/bin/ps"
		}
		for i := 0; i < 1000; i++ {
			fi, err := d.Get(p)
			ok(t, err)
			assert(t, fi != nil, "expect fi not nil")
			equals(t, fi.Path, p)
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
	ok(t, err)
}
