package main

import "github.com/gdamore/tcell/v2"

type rawMark struct {
	off, len int
}

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

	str := l.line[offset:]
	p.Print(x, y, tcell.StyleDefault, str)

	if len(str) < width {
		fillUpLine(p, len(str), y, width)
	}
}

func (l *RawLine) String() string {
	return l.line
}

func (l *RawLine) SetMarks(marks ...Mark) {
}

func (l *RawLine) Len() int {
	return len(l.line)
}

type RawParser struct{}

func (*RawParser) Parse(line string) Line {
	return ParseRawLine(line)
}
