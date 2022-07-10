package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type Printer interface {
	Print(x, y int, style tcell.Style, str string) int
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) int {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}

	return x
}

// printer color, default

type ColorPrinter struct {
	s tcell.Screen
}

func (cp *ColorPrinter) Print(x, y int, style tcell.Style, str string) int {
	return emitStr(cp.s, x, y, style, str)
}

// raw printer ignore color style

type RawPrinter struct {
	s tcell.Screen
}

func (cp *RawPrinter) Print(x, y int, style tcell.Style, str string) int {
	return emitStr(cp.s, x, y, tcell.StyleDefault, str)
}
