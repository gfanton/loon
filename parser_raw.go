package main

import "github.com/gdamore/tcell/v2"

type rawMark struct {
	off, len int
}

type RawLine struct {
	sid  SourceID
	line string
}

func ParseRawLine(sid SourceID, line string) *RawLine {
	return &RawLine{sid, line}
}

func (l *RawLine) Print(p Printer, x, y, width, offset int) {
	if offset >= len(l.line) {
		return
	}

	str := l.line[offset:]
	p.Print(x, y, tcell.StyleDefault, str)

	if len(str) < width {
		fillUpLine(p, len(str), y, width, tcell.StyleDefault)
	}
}

func (l *RawLine) String() string {
	return l.line
}

func (l *RawLine) Source() SourceID {
	return l.sid
}

func (l *RawLine) SetMarks(marks ...Mark) {
}

func (l *RawLine) Len() int {
	return len(l.line)
}

type RawParser struct{}

func (*RawParser) Parse(sid SourceID, line string) Line {
	return ParseRawLine(sid, line)
}
