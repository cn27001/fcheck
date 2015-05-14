package fcheck

import "os"

//Walker represents an object that can be initialized/destroyed before/after a filepath is walked
type Walker interface {
	Walk(string, os.FileInfo, error) error
	Start() error
	Stop() error
}
