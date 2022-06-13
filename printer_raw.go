package main

import "github.com/gdamore/tcell/v2"

type RawPrinter struct {
	s tcell.Screen
}

func (cp *RawPrinter) Print(x, y int, style tcell.Style, str string) {
	emitStr(cp.s, x, y, tcell.StyleDefault, str)
}
