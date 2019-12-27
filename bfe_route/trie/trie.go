package trie

import (
	"errors"
)

type trieChildren map[string]*Trie

type Trie struct {
	Entry      interface{}
	SplatEntry interface{}
	Children   trieChildren
}

func NewTrie() *Trie {
	return &Trie{
		Children: make(trieChildren),
	}
}

func (t *Trie) Get(path []string) (entry interface{}, ok bool) {
	if len(path) == 0 {
		return t.getEntry()
	}

	key := path[0]
	newPath := path[1:]

	if res, ok := t.Children[key]; ok {
		entry, ok = res.Get(newPath)
	}

	if entry == nil && t.SplatEntry != nil {
		entry = t.SplatEntry
		ok = true
	}

	return
}

func (t *Trie) Set(path []string, value interface{}) error {
	if len(path) == 0 {
		t.setEntry(value)
		return nil
	}

	if path[0] == "*" {
		if len(path) != 1 {
			return errors.New("* should be last element")
		}
		t.SplatEntry = value
	}

	key := path[0]
	newPath := path[1:]

	res, ok := t.Children[key]
	if !ok {
		res = NewTrie()
		t.Children[key] = res
	}

	return res.Set(newPath, value)
}

func (t *Trie) setEntry(value interface{}) {
	t.Entry = value
}

func (t *Trie) getEntry() (entry interface{}, ok bool) {
	return t.Entry, t.Entry != nil
}
