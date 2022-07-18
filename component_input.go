package main

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
)

var simpleFilter = func(input string, value string) bool {
	inputs := strings.Split(input, " ")
	for _, in := range inputs {
		if strings.Contains(value, in) {
			return true
		}
	}

	return false
}

type InputComponent struct {
	input *Input

	x, y    int
	printer Printer
}

func NewInputComponent(lcfg *LoonConfig, p Printer, input *Input, xpos, ypos int) *InputComponent {
	return &InputComponent{
		input:   input,
		printer: p,
		x:       xpos, y: ypos,
	}
}

func (i *InputComponent) Redraw(x, y, width, height int) {
	style := tcell.StyleDefault
	input := i.input.Get()
	if input == "" {
		input = ":"
		style = style.Blink(true)
	}

	var offset int
	if size := len(input); size > width {
		offset = size - width
	}

	xoffset := i.printer.Print(x, y, style, input[offset:])
	fillUpLine(i.printer, xoffset, y, width, tcell.StyleDefault)
}

type Input struct {
	muRunes sync.RWMutex
	runes   string
}

func (i *Input) Get() string {
	i.muRunes.RLock()
	input := i.runes
	i.muRunes.RUnlock()

	return input
}

func (i *Input) Add(r rune) {
	i.muRunes.Lock()
	i.runes += string(r)
	i.muRunes.Unlock()
}

func (i *Input) DeleteBackward() {
	i.muRunes.Lock()
	if size := len(i.runes); size > 0 {
		i.runes = i.runes[:size-1]
	}
	i.muRunes.Unlock()
}
