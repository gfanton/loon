package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type Printer interface {
	Print(x, y int, style tcell.Style, str string) int
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) int {
	var comb []rune

	for _, c := range str {
		w := runewidth.RuneWidth(c)
		if w == 0 {
			c = 0
			w = 0
		}

		s.SetContent(x, y, c, comb, style)
		x += w
	}

	return x
}

func fillUpLine(printer Printer, startx, y, width int, style tcell.Style) {
	// fillup line with empty space
	if startx < width {
		printer.Print(startx, y, style, strings.Repeat(" ", width-startx))
	}
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

func (rp *RawPrinter) Print(x, y int, style tcell.Style, str string) int {
	return emitStr(rp.s, x, y, tcell.StyleDefault, str)
}

// func emitTruncateStr(s tcell.Screen, x, y int, style tcell.Style, str string) (int, int) {
// 	sw, sh := s.Size()

// 	var comb []rune

// 	for _, c := range str {
// 		w := runewidth.RuneWidth(c)
// 		if w == 0 {
// 			c = 0
// 			w = 0
// 		}

// 		if x >= sw {
// 			s.SetContent(x, y, tcell.RuneDegree, comb, style)
// 			x = 0
// 			y += 1
// 		}

// 		s.SetContent(x, y, c, comb, style)
// 		x += w

// 		if y > sh {
// 			return x, y
// 		}
// 	}

// 	return x, y
// }
