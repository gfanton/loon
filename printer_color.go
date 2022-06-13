package main

import (
	"github.com/gdamore/tcell/v2"
)

type ColorPrinter struct {
	s tcell.Screen
}

func (cp *ColorPrinter) Print(x, y int, style tcell.Style, str string) {
	emitStr(cp.s, x, y, style, str)
}
