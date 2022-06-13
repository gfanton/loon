package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type FooterComponent struct {
	s       tcell.Screen
	printer Printer
	pos     *position
}

func NewFooterComponent(s tcell.Screen, p Printer, pos *position) *FooterComponent {
	return &FooterComponent{
		s:       s,
		printer: p,
		pos:     pos,
	}
}

func (i *FooterComponent) Redraw(x, y, width, height int) {
	w, h := i.s.Size()

	cursor, offset := i.pos.Cursor(), i.pos.Offset()
	line := fmt.Sprintf("cursor: %d, offset: %d, maxCursor: %d, maxOffset: %d, height: %d, width: %d",
		cursor, offset, i.pos.maxCursor, i.pos.maxOffset, h, w)

	i.printer.Print(x, y, tcell.StyleDefault, line)
}
