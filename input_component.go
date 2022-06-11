package main

import (
	"github.com/gdamore/tcell/v2"
)

type InputComponent struct {
	input *Input

	x, y int
	s    tcell.Screen
}

func NewInputComponent(s tcell.Screen, input *Input, xpos, ypos int) *InputComponent {
	return &InputComponent{
		input: input,
		s:     s,
		x:     xpos, y: ypos,
	}
}

func (i *InputComponent) Redraw(x, y, width, height int) {
	input := i.input.Get()
	var offset int
	if size := len(input); size > width {
		offset = size - width
	}

	emitStr(i.s, x, y, tcell.StyleDefault, input[offset:])
}
