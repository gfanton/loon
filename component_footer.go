package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type FooterComponent struct {
	s       tcell.Screen
	printer Printer
	ring    *Buffer
}

func NewFooterComponent(s tcell.Screen, p Printer, ring *Buffer) *FooterComponent {
	return &FooterComponent{
		s:       s,
		printer: p,
		ring:    ring,
	}
}

func (i *FooterComponent) Redraw(x, y, width, height int) {
	w, h := i.s.Size()

	lines, size, offset := i.ring.Infos()
	line := fmt.Sprintf("offset: %d, size: %d, lines: %d, height: %d, width: %d",
		offset, size, lines, h, w)

	i.printer.Print(x, y, tcell.StyleDefault, line)
}
