package fcheck

import "io"

//Walker represents an object that can be initialized/destroyed before/after a filepath is walked
type Walker interface {
	StartWalking(path string) error
	StartStopper
}

//StartStopper represents an object that can be initialized/destroyed by calling Start and Stop
type StartStopper interface {
	Start() error
	Stop() error
}

type FileInfoWriter interface {
	StartStopper
	Put(fc *FileCheckInfo) error
}

type FileInfoReader interface {
	StartStopper
	Get(path string) (*FileCheckInfo, error)
	Map(path string, callback DBMapFunc) error
}

//DBMapFunc is the callback function definition used by Map
type DBMapFunc func(value *FileCheckInfo) error

type PositionalWriteCloser interface {
	io.WriteCloser
	io.Seeker
}

type PositionalReadCloser interface {
	io.ReadCloser
	io.Seeker
}
