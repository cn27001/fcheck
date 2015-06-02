package fcheck

import "io"

//StringSet represents a simple set of strings
type StringSet map[string]int8

//Add a string to a set
func (r *StringSet) Add(s string) {
	(*r)[s] = 1
}

//Has returns true if string is present in the set
func (r *StringSet) Has(s string) bool {
	_, ok := (*r)[s]
	return ok
}

//Del removes a string from the set
func (r *StringSet) Del(s string) {
	delete((*r), s)
}

//Items returns all strings in a set as a slice
func (r *StringSet) Items() []string {
	v := make([]string, 0, len(*r))
	for k := range *r {
		v = append(v, k)
	}
	return v
}

//Walker represents an object that can be initialized/destroyed before/after a filepath is walked
type Walker interface {
	StartWalking(path string, exclude StringSet) error
	StartStopper
}

//StartStopper represents an object that can be initialized/destroyed by calling Start and Stop
type StartStopper interface {
	Start() error
	Stop() error
}

//FileInfoWriter is an interface for writing FileCheckInfo records to DB
type FileInfoWriter interface {
	StartStopper
	Put(fc *FileCheckInfo) error
}

//FileInfoReader is an interface for reading FileCheckInfo records from DB
type FileInfoReader interface {
	StartStopper
	Get(path string) (*FileCheckInfo, error)
	Map(path string, callback DBMapFunc) error
}

//DBMapFunc is the callback function definition used by Map
type DBMapFunc func(value *FileCheckInfo) error

//PositionalWriteCloser is minimum subset of io interfaces used by the DB code
type PositionalWriteCloser interface {
	io.WriteCloser
	io.Seeker
}

//PositionalReadCloser is minimum subset of io interfaces used by the DB code
type PositionalReadCloser interface {
	io.ReadCloser
	io.Seeker
}
