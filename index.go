package fcheck

import (
	"encoding/gob"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

func splitPath(path string) []string {
	return strings.Split(filepath.Clean(path), string(filepath.Separator))
}

//PathIndex holds our path index structure
type PathIndex struct {
	root *PEntry
}

//NewPathIndex returns new instance of PathIndex
func NewPathIndex() *PathIndex {
	return &PathIndex{NewPEntry()}
}

//Get returns the value from PathIndex stored under key k
func (pi *PathIndex) Get(k string) (int64, bool) {
	var node *PEntry
	var ok bool
	node, ok = pi.GetNode(k)
	if !ok {
		return -1, ok
	}
	return node.Pos, ok
}

//GetNode returns the PEntry from PathIndex stored under key k
func (pi *PathIndex) GetNode(k string) (*PEntry, bool) {
	pathParts := splitPath(k)
	if len(pathParts) == 1 {
		return pi.root, true
	}
	if pe := pi.root.get(pathParts[1:]); pe != nil {
		return pe, true
	}
	return nil, false
}

//Set sets value to v under key k, creating any parent entries in the process as needed
func (pi *PathIndex) Set(k string, v int64) {
	//if not slash return error ?
	pathParts := splitPath(k)
	if len(pathParts) == 1 {
		pi.root.Pos = v
		return
	}
	pe := pi.root.getOrCreate(pathParts[1:])
	pe.Pos = v
}

//Save stores the PathIndex to disk
func (pi *PathIndex) Save(f io.Writer) error {
	enc := gob.NewEncoder(f)
	return enc.Encode(pi.root)
}

//Load re-stores the PathIndex from disk
func (pi *PathIndex) Load(f io.Reader) error {
	dec := gob.NewDecoder(f)
	pe := NewPEntry()
	if err := dec.Decode(pe); err != nil {
		return err
	}
	pi.root = pe
	return nil
}

//Size returns the number of entries in PathIndex
func (pi *PathIndex) Size() int64 {
	return pi.root.size()
}

type nodeStepF func(node *PEntry)

//PEntry represents actual path index entry
type PEntry struct {
	Name     string
	Pos      int64
	Children []*PEntry
}

//NewPEntry returns a new PEntry instance
func NewPEntry() *PEntry {
	return &PEntry{"", -1, nil}
}

func (pe *PEntry) get(parts []string) *PEntry {
	head := parts[0]
	tail := parts[1:]
	if idx, exists := pe.idxChild(head); exists {
		if len(tail) == 0 {
			return pe.Children[idx]
		}
		return pe.Children[idx].get(tail)
	}
	return nil
}

func (pe *PEntry) getOrCreate(parts []string) *PEntry {
	if len(parts) == 0 {
		return pe
	}
	head := parts[0]
	tail := parts[1:]
	idx, exists := pe.idxChild(head)
	if exists {
		if len(tail) == 0 {
			return pe.Children[idx]
		}
		return pe.Children[idx].getOrCreate(tail)
	}
	ch := pe.addChild(idx, head)
	if len(tail) == 0 {
		return ch
	}
	return ch.getOrCreate(tail)
}

func (pe *PEntry) idxChild(childName string) (int, bool) {
	x := childName
	data := pe.Children
	i := sort.Search(len(data), func(i int) bool { return data[i].Name >= x })
	if i < len(data) && data[i].Name == x {
		// x is present at data[i]
		return i, true
	}
	// else
	// x is not present in data,
	// but i is the index where it would be inserted.
	return i, false
}

func (pe *PEntry) addChild(idx int, name string) *PEntry {
	ch := NewPEntry()
	ch.Name = name
	pe.Children = append(pe.Children, nil)
	copy(pe.Children[idx+1:], pe.Children[idx:])
	pe.Children[idx] = ch
	return ch
}

func (pe *PEntry) size() int64 {
	var n int64 = 1
	for _, v := range pe.Children {
		n = n + v.size()
	}
	return n
}

//Traverse will travers the PEntry tree and call f on each entry
func (pe *PEntry) Traverse(f nodeStepF) {
	f(pe)
	for _, v := range pe.Children {
		v.Traverse(f)
	}
}
