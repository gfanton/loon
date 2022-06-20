package main

import (
	"bytes"

	"github.com/gdamore/tcell/v2"
	ansi "github.com/leaanthony/go-ansi-parser"
)

type lineSequence struct {
	Style       tcell.Style
	Index, Size int
}

type ANSILine struct {
	seqs    []lineSequence
	content bytes.Buffer
}

func ParseANSILine(line string, color bool) *ANSILine {
	var l ANSILine
	if st, err := ansi.Parse(line); err == nil {
		l.seqs = make([]lineSequence, len(st))
		var index int
		for i, s := range st {
			l.content.WriteString(s.Label)
			if color {
				l.seqs[i].Style = styledcell(s)
			} else {
				l.seqs[i].Style = tcell.StyleDefault
			}
			l.seqs[i].Size = len(s.Label)
			l.seqs[i].Index = index
			index += len(s.Label)
		}
	} else {
		l.content.WriteString(line)
	}

	return &l
}

func (l *ANSILine) Print(p Printer, x, y, width, offset int) {
	content := l.content.Bytes()
	for _, s := range l.seqs {
		// if s.Index > width {
		// 	break
		// }

		to := s.Index + s.Size
		if offset >= to {
			continue
		}

		from := s.Index
		if offset >= from && offset < to {
			from = offset
		}

		str := content[from:to]
		p.Print(x, y, s.Style, string(str))
		x += len(str)
	}
}

func (l *ANSILine) String() string {
	return l.content.String()
}

func (l *ANSILine) Len() int {
	return l.content.Len()
}

type ANSIParser struct {
	NoColor bool
}

func (p *ANSIParser) Parse(line string) Line {
	return ParseANSILine(line, !p.NoColor)
}
