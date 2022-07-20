package main

import (
	"bytes"
	"hash/fnv"
	"strings"

	"github.com/gdamore/tcell/v2"
	ansi "github.com/leaanthony/go-ansi-parser"
)

type lineSequence struct {
	Style       tcell.Style
	Index, Size int
}

type ANSILine struct {
	content bytes.Buffer
	sid     SourceID
	bgcol   tcell.Color
	seqs    []*lineSequence
	marks   []Mark
}

func ParseANSILine(line string, color bool) *ANSILine {
	var l ANSILine
	if st, err := ansi.Parse(line); err == nil {
		l.seqs = make([]*lineSequence, len(st))
		var index int
		for i, s := range st {
			l.seqs[i] = &lineSequence{}
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

func (l *ANSILine) printSeqs(p Printer, content []byte, x, y, width, offset int) (pl int) {
	for _, s := range l.seqs {
		from, to := s.Index, s.Index+s.Size
		if offset >= to || from > (width+offset) {
			continue
		}

		if offset >= from {
			from = offset
		}

		str := content[from:to]
		p.Print(x, y, s.Style.Background(l.bgcol), string(str))

		x += len(str) // x offset
	}

	fillUpLine(p, x, y, width, tcell.StyleDefault.Background(l.bgcol))
	return
}

func (l *ANSILine) printMarks(p Printer, content []byte, x, y, width, offset int) {
	if offset < 0 {
		offset = 0
	}

	for _, m := range l.marks {
		from, to := m.Off, m.Off+m.Len
		if offset >= to || from > (width+offset) {
			continue
		}

		if offset >= from {
			from = offset
		}

		str := content[from:to]
		style := tcell.StyleDefault.Background(getMarkColor(m.N)).Reverse(true).Bold(true)

		p.Print(x+from-offset, y, style, string(str))
	}
}

func (l *ANSILine) Source() SourceID {
	return l.sid
}

func (l *ANSILine) Print(p Printer, x, y, width, offset int) {
	content := l.content.Bytes()
	l.printSeqs(p, content, x, y, width, offset)
	l.printMarks(p, content, x, y, width, offset)
}

func (l *ANSILine) SetMarks(marks ...Mark) {
	l.marks = marks
}

func (l *ANSILine) String() string {
	return l.content.String()
}

func (l *ANSILine) Len() int {
	return l.content.Len()
}

type ANSIParser struct {
	NoColor     bool
	SourceColor bool
}

func (p *ANSIParser) Parse(sid SourceID, line string) Line {
	ansiline := ParseANSILine(line, !p.NoColor)
	ansiline.sid = sid
	if p.SourceColor {
		ansiline.bgcol = sid.Color(0.75)
	}
	return ansiline
}

func styledcell(as *ansi.StyledText) (ts tcell.Style) {
	ts = tcell.StyleDefault
	if as == nil {
		return
	}

	if as.BgCol != nil {
		name := strings.TrimRight(strings.ToLower(as.BgCol.Name), "123456789")
		if c, ok := tcell.ColorNames[name]; ok {
			ts = ts.Background(c)
		} else {
			ts = ts.Background(tcell.NewRGBColor(
				int32(as.BgCol.Rgb.R), int32(as.BgCol.Rgb.G), int32(as.BgCol.Rgb.B),
			))
		}
	}

	if as.FgCol != nil {
		name := strings.TrimRight(strings.ToLower(as.FgCol.Name), "123456789")
		if c, ok := tcell.ColorNames[name]; ok {
			ts = ts.Foreground(c)
		} else {
			ts = ts.Foreground(tcell.NewRGBColor(
				int32(as.FgCol.Rgb.R), int32(as.FgCol.Rgb.G), int32(as.FgCol.Rgb.B),
			))
		}
	}

	return ts.
		Italic(as.Italic()).
		Bold(as.Bold()).
		Underline(as.Underlined()).
		Reverse(as.Inversed()).
		Blink(as.Blinking()).
		StrikeThrough(as.Strikethrough())
}

var markColors = []tcell.Color{
	tcell.ColorDefault,
	tcell.ColorDarkRed,
	tcell.ColorDarkBlue,
	tcell.ColorDarkGreen,
	tcell.ColorDarkOrange,
	tcell.ColorDarkMagenta,
	tcell.ColorDarkOliveGreen,
	tcell.ColorDarkCyan,
	tcell.ColorDarkTurquoise,
	tcell.ColorDarkViolet,
}

func getMarkColor(n int) tcell.Color {
	return markColors[n%len(markColors)]
}

func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func init() {
	defaultParseOptions = append(defaultParseOptions,
		ansi.WithDefaultBackgroundColor("black"),
	)
	defaultParseOptions = append(defaultParseOptions,
		ansi.WithDefaultForegroundColor("white"),
	)
}
