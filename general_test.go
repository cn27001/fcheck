package fcheck

import (
	"sync"
	"testing"
)

var (
	testDBName string = "fcheck_test.db"
	testPath   string = "/bin"
)

func TestGenerator(t *testing.T) {
	g := NewGenerator(testDBName)
	err := g.Start()
	ok(t, err)
	err = g.StartWalking(testPath)
	ok(t, err)
	err = g.Stop()
	ok(t, err)
}

func TestPrinter(t *testing.T) {
	p := NewPrinter(testDBName)
	err := p.Start()
	ok(t, err)
	err = p.StartWalking(testPath)
	ok(t, err)
	err = p.Stop()
	ok(t, err)
}

func TestComparator(t *testing.T) {
	cm := NewComparator(testDBName)
	err := cm.Start()
	ok(t, err)
	err = cm.StartWalking(testPath)
	ok(t, err)
	err = cm.Stop()
	ok(t, err)
	equals(t, 0, len(cm.newFiles))
	equals(t, 0, len(cm.changedFiles))
	equals(t, 0, len(cm.removedFiles))
}

func TestComparatorNoPath(t *testing.T) {
	cm := NewComparator(testDBName)
	err := cm.Start()
	ok(t, err)
	//non-exist path
	err = cm.StartWalking("/foobardubar23256646")
	ok(t, err)
	err = cm.Stop()
	ok(t, err)
	equals(t, 0, len(cm.newFiles))
	equals(t, 0, len(cm.changedFiles))
	equals(t, 0, len(cm.removedFiles))
}

func TestComparatorNoPathInDB(t *testing.T) {
	cm := NewComparator(testDBName)
	err := cm.Start()
	ok(t, err)
	//permission errors as ordinary user plus new files
	err = cm.StartWalking("/etc")
	ok(t, err)
	err = cm.Stop()
	ok(t, err)
	assert(t, len(cm.newFiles) > 0, "expected to be all new files in /etc")
	equals(t, 0, len(cm.changedFiles))
	equals(t, 0, len(cm.removedFiles))
}

func TestPrinterNoPath(t *testing.T) {
	cm := NewPrinter(testDBName)
	err := cm.Start()
	ok(t, err)
	//non-exist path
	err = cm.StartWalking("/foobardubar23256646")
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
	assert(t, err == NotFoundErr, "Expected not found on funny key")
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
