package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type FooterComponent struct {
	s   tcell.Screen
	pos *position
}

func NewFooterComponent(s tcell.Screen, pos *position) *FooterComponent {
	return &FooterComponent{
		s:   s,
		pos: pos,
	}
}

func (i *FooterComponent) Redraw(x, y, width, height int) {
	w, h := i.s.Size()

	cursor, offset := i.pos.Cursor(), i.pos.Offset()
	line := fmt.Sprintf("cursor: %d, offset: %d, maxCursor: %d, maxOffset: %d, height: %d, width: %d",
		cursor, offset, i.pos.maxCursor, i.pos.maxOffset, h, w)
	emitStr(i.s, x, y, tcell.StyleDefault, line)
}
