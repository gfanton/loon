package main

import "github.com/gdamore/tcell/v2"

type RawLine struct {
	line string
}

func ParseRawLine(line string) *RawLine {
	return &RawLine{line}
}

func (l *RawLine) Print(p Printer, x, y, width, offset int) {
	if offset >= len(l.line) {
		return
	}
	p.Print(x, y, tcell.StyleDefault, l.line[offset:])
}

func (l *RawLine) String() string {
	return l.line
}

func (l *RawLine) Len() int {
	return len(l.line)
}

type RawParser struct{}

func (*RawParser) Parse(line string) Line {
	return ParseRawLine(line)
}