package main

type Parser interface {
	Parse(line string) Line
}

type Line interface {
	Print(p Printer, x, y, width, offset int)
	String() string
	Len() int
}
