package main

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"

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

var tSeqNone = tSeqMove(tMoveNone)

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
	require.Equal(t, int64(0), buffer.Lines())
	for i := int64(0); i < tc.ReaderSize; i++ {
		_, err := buffer.Readline()
		require.NoError(t, err)
	}
	require.Nil(t, buffer.cursor)

	var input string
	for i, seq := range tc.Sequences {
		name := fmt.Sprintf("seq_%d", i+1)
		t.Run(name, func(t *testing.T) {
			maxoffset := buffer.Lines() - int64(tc.Height)

			switch v := seq.Sequence.(type) {
			case tSeqAdd:
				t.Logf("adding %d lines to reader", v)
				for i := int64(0); i < int64(v); i++ {
					_, err := buffer.Readline()
					require.NoError(t, err)
				}

			case tSeqMove:
				t.Logf("moving %d, maxoffset: %d", v, maxoffset)
				buffer.MoveCursor(maxoffset, int64(v))
			case tSeqInput:
				t.Logf("update input: %s", v)
				input = string(v)
			}

			matchs := buffer.FilterLines(input, int(tc.Height))
			require.Len(t, matchs.lines, int(tc.Height))
			require.Equal(t, len(seq.ExpectedLine), matchs.size, "should have the expected number of result")

			for i, e := range seq.ExpectedLine {
				expected := strconv.Itoa(e)
				line := matchs.lines[i]
				require.Equal(t, expected, line.String())
			}
		})
	}
}

func TestBufferMove(t *testing.T) {
	cases := []testBufferCase{
		{"simple move",
			100, 100, 10,
			[]testBufferSequence{
				{tSeqMove(tMoveUp), tRange(99, 89)},
				{tSeqMove(tMoveDown), tRange(100, 90)},
				{tSeqMove(tMoveUp * 5), tRange(95, 85)},
				{tSeqMove(tMoveDown * 5), tRange(100, 90)},
			},
		},

		{"complex move",
			100, 100, 10,
			[]testBufferSequence{
				{tSeqMove(tMoveUp * 110), tRange(10, 0)},
				{tSeqMove(tMoveUp * 100), tRange(10, 0)},
				{tSeqMove(tMoveDown * 10), tRange(20, 10)},
				{tSeqMove(tMoveDown * 1000), tRange(100, 90)},
				{tSeqMove(tMoveUp * 1000), tRange(10, 0)},
				{tSeqMove(tMoveDown * 40), tRange(50, 40)},
				{tSeqMove(tMoveDown * 50), tRange(100, 90)},
				{tSeqMove(tMoveDown * 50), tRange(100, 90)},
			},
		},

		{"move with input",
			100, 100, 5,
			[]testBufferSequence{
				{tSeqNone, tRange(100, 95)},
				{tSeqInput("11"), []int{11}},
				{tSeqInput("1"), []int{100, 91, 81, 71, 61}},
				{tSeqMove(tMoveUp), []int{91, 81, 71, 61, 51}},
				{tSeqMove(tMoveUp * 8), []int{19, 18, 17, 16, 15}},
				{tSeqMove(tMoveUp * 100), []int{13, 12, 11, 10, 1}},
				{tSeqInput("foo"), []int{}},
			},
		},

		{"move with dynamic reader",
			0, 100, 10,
			[]testBufferSequence{
				{tSeqNone, []int{}},
				{tSeqAdd(50), tRange(50, 40)},
				{tSeqMove(tMoveUp * 10), tRange(40, 30)},
				{tSeqAdd(50), tRange(40, 30)},
				{tSeqAdd(1000), tRange(1050, 1040)},

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
	}
	r = make([]int, size)

	for i := range r {
		r[i] = start + (i * inc)
	}

	return r
}

func TestBufferReadline(t *testing.T) {
	cases := []struct {
		readerSize int64
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
			reader := &testReader{size: tc.readerSize}
			buffer := newTestBuffer(t, reader, tc.bufferSize)
			require.Equal(t, int64(0), buffer.Lines())
			for i := int64(0); i < tc.readerSize; i++ {
				line, err := buffer.Readline()
				require.NoError(t, err)
				require.True(t, reader.validateCurrenntLine(line))
			}
			require.Equal(t, tc.readerSize, buffer.Lines())
		})
	}

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

func (r *testReader) validateCurrenntLine(line string) bool {
	index := r.Lines()
	return fmt.Sprintf("%d", index) == line
}

func (r *testReader) Readline() (string, error) {
	index := atomic.AddInt64(&r.index, 1)
	return fmt.Sprintf("%d", index), nil
}
