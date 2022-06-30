package main

import (
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tMoveUp   = 1
	tMoveDown = -1
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
	ReaderSize, BufferSize, Height int64
	Sequences                      []testBufferSequence
}

func testBufferMoveCase(t *testing.T, tc *testBufferCase) {
	reader := &testReader{size: tc.ReaderSize}
	buffer := newTestBuffer(t, reader, int(tc.BufferSize))

	require.Equal(t, uint(0), buffer.Lines())
	for i := int64(0); i < tc.ReaderSize; i++ {
		_, err := buffer.Readline()
		require.NoError(t, err)
	}

	var input string
	filter := func(v interface{}) bool {
		if v == nil {
			return false
		}

		node := v.(*Node)
		line := node.Line.String()
		return simpleFilter(input, line)
	}

	window := NewBufferWindow(buffer, filter, int(tc.Height))

	for i, seq := range tc.Sequences {
		name := fmt.Sprintf("seq_%d", i+1)
		t.Run(name, func(t *testing.T) {

			switch v := seq.Sequence.(type) {
			case tSeqAdd:
				t.Logf("adding %d lines to reader", v)
				for i := int64(0); i < int64(v); i++ {
					_, err := buffer.Readline()
					require.NoError(t, err)

					window.Move(-1)
				}

			case tSeqMove:
				if i := int(v); i != 0 {
					t.Logf("moving %d", i)
					window.Move(i)
				}
			case tSeqResize:
				if i := int(v); i != 0 {
					t.Logf("resize %d", i)
					window.Resize(i)
				}
			case tSeqInput:
				t.Logf("update input: %s", v)
				input = string(v)
				window.Update()
			}

			matchs := []string{}
			window.DoFork(func(index int, v interface{}) {
				if v == nil {
					matchs = append(matchs, "-")
					return
				}

				node := v.(*Node)
				line := node.Line.String()
				matchs = append(matchs, line)
			})

			log.Printf("lines: %v\n", matchs)

			// matchs := buffer.FilterLines(input, int(tc.Height))
			// require.Len(t, matchs, int(tc.Height))
			// require.Equal(t, len(seq.ExpectedLine), len(matchs), "should have the expected number of result")

			for i, e := range seq.ExpectedLine {
				expected := strconv.Itoa(e)
				line := matchs[i]
				require.Equal(t, expected, line)
			}
		})
	}
}

func TestBufferMove(t *testing.T) {
	cases := []testBufferCase{
		{"simple move",
			100, 100, 10,
			[]testBufferSequence{
				{tSeqMove(tMoveNone), tRange(90, 100)},
				{tSeqMove(tMoveUp), tRange(89, 99)},
				{tSeqMove(tMoveDown), tRange(90, 100)},
				{tSeqMove(tMoveUp * 5), tRange(85, 95)},
				{tSeqMove(tMoveDown * 5), tRange(90, 100)},
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
				{tSeqResize(10), tRange(0, 10)},
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
				{tSeqAdd(50), tRange(80, 90)},
				{tSeqAdd(1000), tRange(1080, 1090)},

				// {tSeqMove(tMoveUp), tRange(99, 89)},
				// {tSeqMove(tMoveDown), tRange(100, 90)},
				// {tSeqMove(tMoveUp * 5), tRange(95, 85)},
				// {tSeqMove(tMoveDown * 5), tRange(100, 90)},
			},
		},

		{"move with dynamic reader and input",
			100, 1000, 5,
			[]testBufferSequence{
				{tSeqInput("1"), []int{61, 71, 81, 91, 100}},
				{tSeqMove(tMoveUp * 110), []int{1, 10, 11, 12, 13}},
				{tSeqAdd(1), []int{10, 11, 12, 13, 14}},
				{tSeqAdd(50), []int{132, 133, 134, 135, 136}},
				{tSeqMove(tMoveUp * 1), []int{131, 132, 133, 134, 135}},
				{tSeqMove(tMoveDown * 10), []int{141, 142, 143, 144, 145}},
				{tSeqInput("10"), []int{106, 107, 108, 109, 110}},
				{tSeqInput("12"), []int{12, 112, 120, 121, 122}},
				{tSeqInput(""), []int{118, 119, 120, 121, 122}},
				{tSeqInput("15"), []int{15, 115, 150, 151}},
				{tSeqInput("17"), []int{17, 117}},
				// {tSeqInput("2"), []int{17, 117}},

				//  51, 41, 31, 21, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 1
				// {tSeqAdd(1001), tRange(50, 40)},
				// {tSeqMove(tMoveUp * 10), tRange(40, 30)},
				// {tSeqAdd(50), tRange(40, 30)},
				// {tSeqAdd(1000), tRange(1050, 1040)},

				// {tSeqMove(tMoveUp), tRange(99, 89)},
				// {tSeqMove(tMoveDown), tRange(100, 90)},
				// {tSeqMove(tMoveUp * 5), tRange(95, 85)},
				// {tSeqMove(tMoveDown * 5), tRange(100, 90)},
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

func TestBufferReadline(t *testing.T) {
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
			reader := &testReader{size: int64(tc.readerSize)}
			buffer := newTestBuffer(t, reader, tc.bufferSize)

			for i := uint(0); i < tc.readerSize; i++ {
				line, err := buffer.Readline()
				require.NoError(t, err)
				require.True(t, reader.validateCurrentLine(line))
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

	reader := &testReader{size: int64(readerSize)}
	buffer := newTestBuffer(t, reader, bufferSize)

	assert.Equal(t, uint(0), buffer.Lines())

	halfsize := windowSize / 2
	cursor := uint(0)
	for ; cursor < uint(halfsize); cursor++ {
		line, err := buffer.Readline()
		require.NoError(t, err)
		require.True(t, reader.validateCurrentLine(line))
	}
	require.Equal(t, uint(5), buffer.Lines())

	window := NewBufferWindow(buffer, noopFilter, windowSize)
	l, s := window.Size()
	assert.Equal(t, 10, s)
	assert.Equal(t, 5, l)

	for ; cursor < readerSize; cursor++ {
		line, err := buffer.Readline()
		require.NoError(t, err)
		require.True(t, reader.validateCurrentLine(line))
	}
	require.Equal(t, uint(readerSize), buffer.Lines())

	l, s = window.Size()
	require.Equal(t, 10, s)
	require.Equal(t, 5, l) // FIXME: not true
}

func newTestBuffer(t *testing.T, reader Reader, size int) *Buffer {
	t.Helper()
	parser := &testParser{}
	buffer := NewBuffer(size, parser, reader)
	require.NotNil(t, buffer)
	return buffer
}

type testParser struct{}

type noopLine struct {
	line string
}

func (*noopLine) Print(p Printer, x, y, width, offset int) {}
func (n *noopLine) String() string                         { return n.line }
func (n *noopLine) Len() int                               { return len(n.line) }

func (*testParser) Parse(line string) Line {
	return &noopLine{line}
}

var _ Reader = (*testReader)(nil)

var _ Reader = (*testReader)(nil)

type testReader struct {
	size  int64
	index int64
}

func (r *testReader) Lines() int64 {
	return atomic.LoadInt64(&r.index)
}

func (r *testReader) ResetLines() {
	r.index = 0
}

func (r *testReader) validateCurrentLine(line string) bool {
	index := r.Lines()
	return fmt.Sprintf("%d", index) == line
}

func (r *testReader) Readline() (string, error) {
	index := atomic.AddInt64(&r.index, 1)
	return fmt.Sprintf("%d", index), nil
}
