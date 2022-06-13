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

	parser Parser
	reader Reader
}

type LineMatching struct {
	lines           []Line
	size, maxoffset int
}

func (r *Ring) Clear() {
	r.muRing.Lock()
	r.lines = 0
	r.ring.Value = nil
	r.ring = r.ring.Next()
	r.muRing.Unlock()
}

func (r *Ring) FindLine(nlimit int, cursor int64, match func(string) bool) *LineMatching {
	Line := make([]Line, nlimit)

	var seek int64
	var maxoffset, n int
	r.muRing.RLock()
	for p := r.ring.Prev(); p != r.ring && n < nlimit; p = p.Prev() {
		if p.Value == nil {
			break
		}

		line := p.Value.(string)
		if match(line) {
			if seek < cursor {
				seek++
			} else {
				pline := r.parser.Parse(line)
				Line[n] = pline
				n++
				if pline.Len() > maxoffset {
					maxoffset = pline.Len()
				}
			}
		}
	}
	r.muRing.RUnlock()

	return &LineMatching{
		lines:     Line,
		size:      n,
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
