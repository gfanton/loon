package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWindowSize(t *testing.T) {
	t.Run("0 size", func(t *testing.T) {
		w := NewWindow[int](0)

		size, length := w.Size()
		require.Equal(t, 0, size)
		require.Equal(t, 0, length)
	})

	t.Run("various size update up", func(t *testing.T) {
		w := NewWindow[int](5)

		size, length := w.Size()
		require.Equal(t, 5, size)
		require.Equal(t, 0, length)

		w.PushFront(1)
		size, length = w.Size()
		require.Equal(t, 5, size)
		require.Equal(t, 1, length)

		for i := 0; i < 20; i++ {
			w.PushFront(i)
		}

		size, length = w.Size()
		require.Equal(t, 5, size)
		require.Equal(t, 5, length)
	})

	t.Run("various size update down", func(t *testing.T) {
		w := NewWindow[int](5)

		size, length := w.Size()
		require.Equal(t, 5, size)
		require.Equal(t, 0, length)

		w.PushBack(1)
		size, length = w.Size()
		require.Equal(t, 5, size)
		require.Equal(t, 1, length)

		for i := 0; i < 20; i++ {
			w.PushBack(i)
		}

		size, length = w.Size()
		require.Equal(t, 5, size)
		require.Equal(t, 5, length)
	})
}

func TestWindowReset(t *testing.T) {
	w := NewWindow[int](5)
	for i := 0; i < 20; i++ {
		w.PushFront(i)
	}

	require.Equal(t, []int{15, 16, 17, 18, 19}, w.Slice())
	w.Reset()
	require.Equal(t, []int{}, w.Slice())
}

func TestWindowFull(t *testing.T) {
	w := NewWindow[int](5)

	w.PushFront(1)
	require.False(t, w.IsFull())

	w.PushBack(2)
	require.False(t, w.IsFull())

	w.PushFront(3)
	require.False(t, w.IsFull())

	w.PushFront(4)
	w.PushFront(5)
	require.True(t, w.IsFull())

	w.PushBack(6)
	w.PushBack(7)
	require.True(t, w.IsFull())
}

func TestWindowResize(t *testing.T) {
	cases := []struct {
		Start, Resize int
	}{
		{0, 10},
		{10, 20},
		{20, 10},
		{10, 10},
		{9, 10},
		{10, 9},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("resize %d to %d", tc.Start, tc.Resize)
		t.Run(name, func(t *testing.T) {
			r := NewWindow[int](tc.Start)
			for i := 0; i < tc.Start; i++ {
				r.PushFront(i)
			}

			size, length := r.Size()
			require.Equal(t, tc.Start, size)
			require.Equal(t, tc.Start, length)

			r.Resize(tc.Resize)

			size, length = r.Size()
			require.Equal(t, tc.Resize, size)

			expectedLen := tc.Start
			if tc.Resize < tc.Start {
				expectedLen = tc.Resize
			}
			require.Equal(t, expectedLen, length)
		})
	}
}

type tPush int

const (
	tPushUp tPush = iota
	tPushDown
)

type tWindowMove struct {
	kind tPush
	v    int
}

func TestWindowPush(t *testing.T) {
	cases := []struct {
		Name     string
		Size     int
		Moves    []tWindowMove
		Expected []int
	}{
		{"no size", 0, []tWindowMove{
			{tPushUp, 1},
			{tPushUp, 2},
		}, []int{}},

		{"push up", 5, []tWindowMove{
			{tPushUp, 1},
			{tPushUp, 2},
		}, []int{1, 2}},

		{"push down", 5, []tWindowMove{
			{tPushDown, 2},
			{tPushDown, 1},
		}, []int{1, 2}},

		{"push up and down", 5, []tWindowMove{
			{tPushDown, 2},
			{tPushDown, 1},
			{tPushUp, 3},
			{tPushUp, 4},
			{tPushUp, 5},
		}, []int{1, 2, 3, 4, 5}},

		{"oversize push", 5, []tWindowMove{
			{tPushDown, 2},
			{tPushDown, 1},
			{tPushUp, 3},
			{tPushUp, 4},
			{tPushUp, 5},
		}, []int{1, 2, 3, 4, 5}},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			w := NewWindow[int](tc.Size)
			for i, move := range tc.Moves {
				switch move.kind {
				case tPushUp:
					t.Logf("%d: push up %d", i, move.v)
					w.PushFront(move.v)
				case tPushDown:
					t.Logf("%d: push down %d", i, move.v)
					w.PushBack(move.v)
				default:
					require.FailNow(t, "unknown message type: %v", move.kind)
				}
			}
			result := w.Slice()
			require.Equal(t, tc.Expected, result)
		})
	}
}

type tSlide int

const (
	tSlideUp tSlide = iota
	tSlideDown
)

func TestWindowSlide(t *testing.T) {
	cases := []struct {
		Name       string
		Size, Fill int
		Slide      []tSlide
		Expected   []int
	}{
		// slide front
		{"slide front no values", 5, 0, []tSlide{
			tSlideUp,
		}, []int{}},

		{"slide front one value", 5, 1, []tSlide{
			tSlideUp,
		}, []int{}},

		{"slide front two values", 5, 2, []tSlide{
			tSlideUp,
		}, []int{1}},

		{"slide front three values", 5, 3, []tSlide{
			tSlideUp,
		}, []int{1, 2}},

		{"slide front full values", 5, 5, []tSlide{
			tSlideUp,
		}, []int{1, 2, 3, 4}},

		// slide back
		{"slide back no values", 5, 0, []tSlide{
			tSlideDown,
		}, []int{}},

		{"slide back one value", 5, 1, []tSlide{
			tSlideDown,
		}, []int{}},

		{"slide back two values", 5, 2, []tSlide{
			tSlideDown,
		}, []int{0}},

		{"slide back three values", 5, 3, []tSlide{
			tSlideDown,
		}, []int{0, 1}},

		{"slide back full values", 5, 5, []tSlide{
			tSlideDown,
		}, []int{0, 1, 2, 3}},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			w := NewWindow[int](tc.Size)
			for i := 0; i < tc.Fill; i++ {
				w.PushFront(i)
			}

			for i, slide := range tc.Slide {
				switch slide {
				case tSlideUp:
					t.Logf("%d: slide up", i)
					w.SlideFront()
				case tSlideDown:
					t.Logf("%d: slide down", i)
					w.SlideBack()
				default:
					require.FailNow(t, "unknown message type: %+v", slide)
				}
			}
			result := w.Slice()
			require.Equal(t, tc.Expected, result)

			s, l := w.Size()
			require.Equal(t, len(result), l)
			require.Equal(t, tc.Size, s)
		})
	}
}
