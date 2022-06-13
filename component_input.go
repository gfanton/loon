package main

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type InputComponent struct {
	input *Input

	x, y    int
	printer Printer
}

func NewInputComponent(p Printer, input *Input, xpos, ypos int) *InputComponent {
	return &InputComponent{
		input:   input,
		printer: p,
		x:       xpos, y: ypos,
	}
}

func (i *InputComponent) Redraw(x, y, width, height int) {
	input := i.input.Get()
	var offset int
	if size := len(input); size > width {
		offset = size - width
	}

	i.printer.Print(x, y, tcell.StyleDefault, input[offset:])
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
