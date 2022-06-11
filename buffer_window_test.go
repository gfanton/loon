package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	tMoveUp   = -1
	tMoveDown = 1
	tMoveNone = 0
)

type tSeqMove int
type tSeqInput string
type tSeqAdd int64
type tSeqResize int64

var tSeqNone = tSeqMove(tMoveNone)

var noopFilter = func(v interface{}) bool {
	return true
}

type testBufferSequence struct {
	Sequence     interface{}
	ExpectedLine []int
}

type testBufferCase struct {
	Name                           string
	ReaderSize, BufferSize, Height int
	Sequences                      []testBufferSequence
}

func testBufferMoveCase(t *testing.T, tc *testBufferCase) {
	var input string
	filter := func(n int) bool {
		return strings.Contains(fmt.Sprintf("%d", n), input)
	}

	bw := newTestBufferWindow[int](t, &testParser{}, filter, tc.BufferSize, tc.Height)

	require.Equal(t, uint(0), bw.Lines())
	for i := 0; i < tc.ReaderSize; i++ {
		_, err := bw.Readline()
		require.NoError(t, err)
	}
	require.Equal(t, uint(tc.ReaderSize), bw.Lines())
	size, _ := bw.WindowSize()
	require.Equal(t, tc.Height, size)
	for i, seq := range tc.Sequences {
		switch v := seq.Sequence.(type) {
		case tSeqAdd:
			t.Logf("step_%d: adding %d lines to reader", i+1, v)
			for i := int64(0); i < int64(v); i++ {
				_, err := bw.Readline()
				require.NoError(t, err)

				// buffer.MoveWindow(-1)
			}

		case tSeqMove:
			t.Logf("step_%d: moving %d", i+1, int(v))
			if j := int(v); j != 0 {
				bw.Move(j)
			}
		case tSeqResize:
			t.Logf("step_%d: resizing %d", i+1, int(v))
			bw.Resize(int(v))
			bw.Refresh()
		case tSeqInput:
			t.Logf("step_%d: update input: '%s'", i+1, v)
			input = string(v)
			bw.Refresh()
		}

		_, length := bw.WindowSize()
		matchs := bw.Slice()
		// log.Printf("lines: %v\n", matchs)

		// matchs := buffer.FilterLines(input, int(tc.Height))
		require.Len(t, matchs, length)
		require.Equal(t, len(seq.ExpectedLine), len(matchs), "should have the expected number of result")

		require.Equal(t, seq.ExpectedLine, matchs)
		for i, e := range seq.ExpectedLine {
			line := matchs[i]
			require.Equal(t, e, line)
		}
	}
}

func TestBufferMove(t *testing.T) {
	cases := []testBufferCase{
		{"simple move",
			100, 100, 10,
			[]testBufferSequence{
				// upper bound
				{tSeqMove(tMoveNone), tRange(90, 100)},
				{tSeqMove(tMoveUp), tRange(89, 99)},
				{tSeqMove(tMoveDown), tRange(90, 100)},
				{tSeqMove(tMoveUp * 5), tRange(85, 95)},
				{tSeqMove(tMoveDown * 5), tRange(90, 100)},
				// lower bound
				{tSeqMove(tMoveUp) * 90, tRange(0, 10)},
				{tSeqMove(tMoveUp), tRange(0, 10)},
				{tSeqMove(tMoveDown), tRange(1, 11)},
			},
		},

		{"small buffer",
			100, 5, 10,
			[]testBufferSequence{
				{tSeqNone, []int{96, 97, 98, 99, 100}},
				{tSeqMove(tMoveUp), []int{96, 97, 98, 99, 100}},
				{tSeqMove(tMoveDown), []int{96, 97, 98, 99, 100}},
			},
		},

		{"complex move",
			100, 100, 10,
			[]testBufferSequence{
				{tSeqMove(tMoveUp * 110), tRange(0, 10)},
				{tSeqMove(tMoveUp * 100), tRange(0, 10)},
				{tSeqMove(tMoveDown * 10), tRange(10, 20)},
				{tSeqMove(tMoveDown * 1000), tRange(90, 100)},
				{tSeqMove(tMoveUp * 1000), tRange(0, 10)},
				{tSeqMove(tMoveDown * 40), tRange(40, 50)},
				{tSeqMove(tMoveDown * 50), tRange(90, 100)},
				{tSeqMove(tMoveDown * 50), tRange(90, 100)},
			},
		},

		{"resize",
			100, 100, 0,
			[]testBufferSequence{
				{tSeqNone, []int{}},
				{tSeqResize(10), tRange(90, 100)},
			},
		},

		{"move with input",
			100, 100, 5,
			[]testBufferSequence{
				{tSeqNone, tRange(95, 100)},
				{tSeqInput("11"), []int{11}},
				{tSeqInput("1"), []int{1, 10, 11, 12, 13}},
				{tSeqMove(tMoveUp), []int{1, 10, 11, 12, 13}},
				{tSeqMove(tMoveDown * 5), []int{14, 15, 16, 17, 18}},
				{tSeqMove(tMoveDown * 100), []int{61, 71, 81, 91, 100}},
				{tSeqInput("foo"), []int{}},
			},
		},

		{"move with dynamic reader",
			0, 100, 10,
			[]testBufferSequence{
				{tSeqNone, []int{}},
				{tSeqAdd(50), tRange(40, 50)},
				{tSeqMove(tMoveUp * 10), tRange(30, 40)},
				{tSeqAdd(50), tRange(30, 40)},
				// {tSeqAdd(1000), tRange(1080, 1090)},
			},
		},

		{"move with dynamic reader and input",
			100, 1000, 5,
			[]testBufferSequence{
				{tSeqNone, []int{96, 97, 98, 99, 100}},
				{tSeqInput("1"), []int{61, 71, 81, 91, 100}},
				{tSeqMove(tMoveUp * 110), []int{1, 10, 11, 12, 13}},
				{tSeqAdd(10), []int{1, 10, 11, 12, 13}},
				{tSeqMove(tMoveDown * 110), []int{106, 107, 108, 109, 110}},
				{tSeqAdd(50), []int{156, 157, 158, 159, 160}},
				{tSeqInput("10"), []int{106, 107, 108, 109, 110}},
				{tSeqInput("12"), []int{12, 112, 120, 121, 122}},
				{tSeqInput(""), []int{118, 119, 120, 121, 122}},
				{tSeqInput("15"), []int{15, 115, 150, 151, 152}},
				{tSeqInput("20"), []int{20, 120}},
				{tSeqAdd(1000), []int{720, 820, 920, 1020, 1120}},
				{tSeqMove(tMoveUp), []int{620, 720, 820, 920, 1020}},
				{tSeqAdd(100), []int{620, 720, 820, 920, 1020}},
				{tSeqInput(""), []int{1016, 1017, 1018, 1019, 1020}},
				{tSeqAdd(1), []int{1016, 1017, 1018, 1019, 1020}},
			},
		},

		{"slide full buffer",
			100, 10, 5,
			[]testBufferSequence{
				{tSeqNone, []int{96, 97, 98, 99, 100}},
				{tSeqMove(tMoveUp), []int{95, 96, 97, 98, 99}},
				{tSeqAdd(5), []int{96, 97, 98, 99, 100}},
				{tSeqAdd(5), []int{101, 102, 103, 104, 105}},
				{tSeqInput("10"), []int{101, 102, 103, 104, 105}},
				{tSeqAdd(5), []int{106, 107, 108, 109, 110}},
				{tSeqMove(tMoveUp * 5), []int{106, 107, 108, 109, 110}},
				{tSeqAdd(2), []int{108, 109, 110}},
				{tSeqAdd(10), []int{}},
				{tSeqInput(""), []int{123, 124, 125, 126, 127}},
				{tSeqMove(tMoveDown * 5), []int{123, 124, 125, 126, 127}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testBufferMoveCase(t, &tc)
		})
	}

}

func tRange(start, end int) (r []int) {
	size := end - start

	var inc int
	if size < 0 {
		inc = -1
		size = -size
	} else {
		inc = 1
		start++
	}
	r = make([]int, size)

	for i := range r {
		r[i] = start + (i * inc)
	}

	return r
}

func TestBufferWindowReadline(t *testing.T) {
	cases := []struct {
		readerSize uint
		bufferSize int
	}{
		{100, 1000},
		{1000, 100},
		{1000, 1},
		{1, 1000},
		{1000000, 100},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("reader_%d_buffer_%d", tc.readerSize, tc.bufferSize)
		t.Run(name, func(t *testing.T) {
			parser := &testParser{}
			filter := func(v int) bool { return true }
			buffer := newTestBufferWindow[int](t, parser, filter, tc.bufferSize, 10)

			for i := uint(0); i < tc.readerSize; i++ {
				_, err := buffer.Readline()
				require.NoError(t, err)
			}

			require.Equal(t, tc.readerSize, buffer.Lines())
		})
	}

}

func TestBufferWindowSize(t *testing.T) {
	const (
		readerSize = 100
		bufferSize = 1000
		windowSize = 10
	)

	parser := &testParser{}
	filter := func(v int) bool { return true }
	buffer := newTestBufferWindow[int](t, parser, filter, bufferSize, windowSize)

	require.Equal(t, uint(0), buffer.Lines())

	halfsize := windowSize / 2
	cursor := uint(0)
	for ; cursor < uint(halfsize); cursor++ {
		_, err := buffer.Readline()
		require.NoError(t, err)
	}
	require.Equal(t, uint(halfsize), buffer.Lines())

	view := buffer.Slice()
	require.Equal(t, tRange(0, 5), view)

	for ; cursor < readerSize; cursor++ {
		_, err := buffer.Readline()
		require.NoError(t, err)
	}
	require.Equal(t, uint(readerSize), buffer.Lines())

	view = buffer.Slice()
	require.Equal(t, tRange(90, 100), view)

	// require.Equal(t, 10, s)
	// require.Equal(t, 5, l) // FIXME: not true
}

func newTestBufferWindow[T any](t *testing.T, parser Parser[T], filter Filter[T], bsize, wsize int) *BufferWindow[T] {
	t.Helper()
	buffer := NewBuffer[T](bsize)
	reader := &testReader{}

	opts := BufferWindowOptions[T]{
		Parser: parser,
		Reader: reader,
		Filter: filter,
		Buffer: buffer,
	}

	require.NotNil(t, buffer)
	bw := NewBufferWindow[T](wsize, &opts)
	bw.sync = true
	return bw
}

type testReader struct {
	size  int
	index int32
}

func (r *testReader) Lines() int {
	return int(atomic.LoadInt32(&r.index))
}

func (r *testReader) ResetLines() {
	r.index = 0
}

func (r *testReader) Readline() (string, error) {
	index := atomic.AddInt32(&r.index, 1)
	return strconv.Itoa(int(index)), nil
}

type testParser struct{}

func (p *testParser) Parse(line string) (i int) {
	var err error
	if i, err = strconv.Atoi(line); err != nil {
		i = 0
	}
	return
}
