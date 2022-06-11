package main

import "sync"

type Input struct {
	muRunes sync.RWMutex
	runes   string
}

func (i *Input) Get() string {
	i.muRunes.RLock()
	input := i.runes
	i.muRunes.RUnlock()

	return input
}

func (i *Input) Add(r rune) {
	i.muRunes.Lock()
	i.runes += string(r)
	i.muRunes.Unlock()
}

func (i *Input) DeleteBackward() {
	i.muRunes.Lock()
	if size := len(i.runes); size > 0 {
		i.runes = i.runes[:size-1]
	}
	i.muRunes.Unlock()
}
