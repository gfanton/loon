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

	muPosi sync.RWMutex

	posi  *position
	input *Input

	ring *Ring

	header *InputComponent
	file   *FileComponent
	footer *FooterComponent
}

func NewScreen(lcfg *LoonConfig, ring *Ring) (*Screen, error) {
	encoding.Register()
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("unable to create new screen: %w", err)
	}

	// create input
	input := &Input{}

	posi := &position{}
	posi.SetMaxCursor(ring.lines)

	var printer Printer
	if lcfg.NoColor {
		printer = &RawPrinter{s}
	} else {
		printer = &ColorPrinter{s}
	}

	filec := NewFileComponent(printer, input, posi, ring)
	inputc := NewInputComponent(printer, input, 1, 0)
	footerc := NewFooterComponent(s, printer, posi)
	return &Screen{
		cupdate: make(chan struct{}, 1),
		ts:      s,
		ring:    ring,

		input: input,
		posi:  posi,

		header: inputc,
		file:   filec,
		footer: footerc,
	}, nil
}

func (s *Screen) readfile() {
	for {
		_, err := s.ring.Readline()
		if err != nil {
			return
		}

		_, h := s.ts.Size()
		s.posi.SetMaxCursor(s.ring.Lines())
		if s.posi.Cursor() > 0 {
			s.posi.AddCursor(1)
			if s.ring.Lines() > int64(h) {
				s.file.IncWindow(1)
			}
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
	var cursor, offset int64

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
		s.posi.Add(cursor, offset)
	}

	s.Redraw()
}

func (s *Screen) handleEventKey(ev *tcell.EventKey) error {
	switch ev.Key() {
	case tcell.KeyUp:
		s.posi.AddCursor(1)
	case tcell.KeyUpRight:
		s.posi.Add(1, 1)
	case tcell.KeyRight:
		s.posi.AddOffset(2)
	case tcell.KeyDownRight:
		s.posi.Add(-1, 1)
	case tcell.KeyDown:
		s.posi.AddCursor(-1)
	case tcell.KeyDownLeft:
		s.posi.Add(-1, -1)
	case tcell.KeyLeft:
		s.posi.AddOffset(-2)
	case tcell.KeyUpLeft:
		s.posi.Add(1, -1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		s.input.DeleteBackward()
	case tcell.KeyEnter:
		s.posi.SetCursor(0)
	default:
		if r := ev.Rune(); r >= 41 && r <= 176 {
			s.input.Add(r)
			// reset cursor
			s.posi.SetCursor(0)
		} else {
			return nil
		}
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
