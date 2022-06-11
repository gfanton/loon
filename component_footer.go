package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type FooterComponent struct {
	s       tcell.Screen
	printer Printer
	buffer  *BufferWindow[Line]
}

func NewFooterComponent(s tcell.Screen, p Printer, buffer *BufferWindow[Line]) *FooterComponent {
	return &FooterComponent{
		s:       s,
		printer: p,
		buffer:  buffer,
	}
}

func (i *FooterComponent) Redraw(x, y, width, height int) {
	w, h := i.s.Size()

	lines := i.buffer.Lines()
	line := fmt.Sprintf("height: %d, width: %d, lines: %d", h, w, lines)

	i.printer.Print(x, y, tcell.StyleDefault, line)
}
