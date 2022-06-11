package main

import (
	"container/ring"
	"sync"
)

type Ring struct {
	ring       *ring.Ring
	ringSize   int
	updateRing chan struct{}
	muRing     sync.RWMutex
	lines      int64

	reader Reader
}

type LineMatching struct {
	lines           []string
	size, maxoffset int
}

func (r *Ring) FindLine(max int, cursor int64, match func(line string) bool) *LineMatching {
	buffer := make([]string, max)

	var seek int64
	var maxoffset, size int
	r.muRing.RLock()
	for p := r.ring.Prev(); p != r.ring && size < max; p = p.Prev() {
		if p.Value == nil {
			break
		}

		line := p.Value.(string)
		if match(line) {
			if seek < cursor {
				seek++
			} else {

				buffer[size] = line
				size++

				if len(line) > maxoffset {
					maxoffset = len(line)
				}
			}
		}
	}
	r.muRing.RUnlock()

	return &LineMatching{
		lines:     buffer,
		size:      size,
		maxoffset: maxoffset,
	}

}

func (r *Ring) Readline() (string, error) {
	line, err := r.reader.Readline()
	if err != nil {
		return "", err
	}

	r.muRing.Lock()
	r.ring.Value = line
	r.ring = r.ring.Next()
	r.lines++
	r.muRing.Unlock()

	return line, nil
}

func (r *Ring) Size() (l int64) {
	r.muRing.RLock()
	if l = r.lines; l > int64(r.ringSize) {
		l = int64(r.ringSize)
	}

	r.muRing.RUnlock()
	return
}

func (r *Ring) Lines() (l int64) {
	r.muRing.RLock()
	l = r.lines
	r.muRing.RUnlock()
	return
}
