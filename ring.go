// Package ring implements operations on circular lists.
package main

import "container/ring"

func DoRingNext(r *ring.Ring, f func(r *ring.Ring) bool) {
	if r != nil {
		if !f(r) {
			return
		}
		for n := r.Next(); n != r; n = n.Next() {
			if !f(n) {
				return
			}
		}
	}
}

// Prev returns the previous ring element. r must not be empty.
func DoRingPrev(r *ring.Ring, f func(r *ring.Ring) bool) {
	if r != nil {
		var p *ring.Ring
		for p = r.Prev(); p != r; p = p.Prev() {
			if !f(p) {
				return
			}
		}

		f(p)
	}
}
