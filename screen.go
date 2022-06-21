package main

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

type Screen struct {
	ts      tcell.Screen
	cupdate chan struct{}

	muScreen sync.RWMutex

	input *Input

	ring *Buffer

	header *InputComponent
	file   *FileComponent
	footer *FooterComponent
}

func NewScreen(lcfg *LoonConfig, ring *Buffer) (*Screen, error) {
	encoding.Register()
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("unable to create new screen: %w", err)
	}

	// create input
	input := &Input{}

	// posi := &position{}
	// posi.SetMaxCursor(ring.Lines())

	var printer Printer
	if lcfg.NoColor {
		printer = &RawPrinter{s}
	} else {
		printer = &ColorPrinter{s}
	}

	filec := NewFileComponent(printer, input, ring)
	inputc := NewInputComponent(printer, input, 1, 0)
	footerc := NewFooterComponent(s, printer, ring)
	return &Screen{
		cupdate: make(chan struct{}, 1),
		ts:      s,
		ring:    ring,

		input:  input,
		header: inputc,
		file:   filec,
		footer: footerc,
	}, nil
}

func (s *Screen) Clear() {
	s.muScreen.Lock()
	s.ring.Clear()
	s.Redraw()
	s.muScreen.Unlock()

}

func (s *Screen) readfile() {
	for {
		_, err := s.ring.Readline()
		if err != nil {
			return
		}

		s.Redraw()
	}
}

func (s *Screen) Run() error {
	// init screen
	if err := s.ts.Init(); err != nil {
		return fmt.Errorf("unable to init new screen: %w", err)
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)

	s.ts.SetStyle(defStyle)

	s.ts.EnableMouse()

	go s.redrawLoop()
	go s.readfile()

	for {
		switch ev := s.ts.PollEvent().(type) {
		case *tcell.EventError:
			return fmt.Errorf("interrupted: %w", ev)
		case *tcell.EventResize:
			s.ts.Sync()
			s.Redraw()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				s.ts.Fini()
				return nil
			}

			go s.handleEventKey(ev)
		case *tcell.EventMouse:
			go s.handleEventMouse(ev)
		}

	}

}

func (s *Screen) handleEventMouse(ev *tcell.EventMouse) {
	var cursor, offset int

	button := ev.Buttons()
	if button&tcell.WheelLeft != 0 {
		offset += 1
	} else if button&tcell.WheelRight != 0 {
		offset -= 1
	}

	if button&tcell.WheelDown != 0 {
		cursor -= 1
	} else if button&tcell.WheelUp != 0 {
		cursor += 1
	}

	if cursor != 0 || offset != 0 {
		s.file.MoveAdd(offset, cursor)
	}

	s.Redraw()
}

func (s *Screen) handleEventKey(ev *tcell.EventKey) error {
	var factor int
	switch {
	case ev.Modifiers()&tcell.ModCtrl != 0:
		return s.handleCtrlCommand(ev)
	case ev.Modifiers()&tcell.ModAlt != 0:
		factor = 5
	default:
		factor = 1
	}

	switch ev.Key() {
	case tcell.KeyUp:
		s.file.CursorAdd(1 * factor)
	case tcell.KeyRight:
		s.file.OffsetAdd(2 * factor)
	case tcell.KeyDown:
		s.file.CursorAdd(-1 * factor)
	case tcell.KeyLeft:
		s.file.OffsetAdd(-2 * factor)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		s.input.DeleteBackward()
	case tcell.KeyEnter:
		s.file.ResetPosition()
	default:
		if r := ev.Rune(); (r >= 41 && r <= 176) || r == ' ' {
			s.input.Add(r)
			s.file.ResetPosition()
		} else {
			return nil
		}
	}

	s.Redraw()
	return nil
}

func (s *Screen) handleCtrlCommand(ev *tcell.EventKey) error {
	switch ev.Key() {
	case tcell.KeyCtrlE:
		s.file.OffsetSet(s.file.MaxOffset())
	case tcell.KeyCtrlA:
		s.file.OffsetSet(0)
	case tcell.KeyCtrlL:
		s.Clear()
	default:
	}

	s.Redraw()
	return nil
}

func (s *Screen) Redraw() {
	select {
	case s.cupdate <- struct{}{}:
	default:
	}
}

func (s *Screen) redrawLoop() {
	for range s.cupdate {
		s.redraw()
	}
}

func (s *Screen) redraw() {
	w, h := s.ts.Size()

	s.ts.Clear()

	// header start at x:1,y:0
	s.header.Redraw(1, 0, w, 1)

	// file start at x:1, y:1
	s.file.Redraw(0, 1, w, h-2)

	// file start at x:1, y:1
	s.footer.Redraw(1, h-1, w, 1)

	s.ts.Show()

}
