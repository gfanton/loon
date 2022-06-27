package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRingResize(t *testing.T) {
	cases := []struct {
		Start, Resize int
	}{
		{10, 20},
		{20, 10},
		{10, 10},
		{9, 10},
		{10, 9},
	}

	for _, tc := range cases {
		name := fmt.Sprintf("resize %d to %d", tc.Start, tc.Resize)
		t.Run(name, func(t *testing.T) {
			r := NewLoop(tc.Start)
			require.Equal(t, tc.Start, r.root.Len())
			require.Equal(t, tc.Start, r.size)
			r.Resize(tc.Resize)
			require.Equal(t, tc.Resize, r.root.Len())
			require.Equal(t, tc.Resize, r.size)
		})
	}
}

// func TestRingDo(t *testing.T) {
// 	r := NewLoop(10)
// 	for i := 0; i < 10; i ++ {
// 		root := r.Root()
// 		root.Value =

// 	}

// 	require.Equal(t, tc.Start, r.root.Len())
// 	require.Equal(t, tc.Start, r.size)
// 	r.Resize(tc.Resize)
// 	require.Equal(t, tc.Resize, r.root.Len())
// 	require.Equal(t, tc.Resize, r.size)
// }

// func TestLoopMove(t *testing.T) {
// 	cases := []struct {
// 		Size, Window int
// 		Move         []int
// 	}{
// 		{10, 10, []int{0, 10}},
// 	}

// 	for _, tc := range cases {
// 		name := fmt.Sprintf("move size: %d, windown %d", tc.Size, tc.Window)
// 		t.Run(name, func(t *testing.T) {
// 			loop := NewLoop(tc.Size)
// 			for i := 0; i < tc.Size; i++ {
// 				loop.PushFrontValue(i + 1)
// 			}

// 			// var input string
// 			// filter := func (v interface{} {

// 			// })

// 			loop.Do(func(r *ring.Ring) {
// 				fmt.Println("loop:", r.Value)
// 			}, tc.Size)

// 			w := NewBufferWindow(loop, tc.Window)
// 			w.DoFork(func(v interface{}) {
// 				fmt.Println("fork loop:", v)
// 			})

// 			for _, m := range tc.Move {
// 				w.Move(m)
// 				fmt.Println("---")
// 				w.DoFork(func(v interface{}) {
// 					fmt.Println(v)
// 				})
// 				// w.DoFork(func(v interface{}) {
// 				// 	require.Equal(t, expected interface{}, actual interface{}, msgAndArgs ...interface{})
// 				// })
// 			}

// 			// require.Equal(t, tc.Start, r.root.Len())
// 			// require.Equal(t, tc.Start, r.size)
// 			// r.Resize(tc.Resize)
// 			// require.Equal(t, tc.Resize, r.root.Len())
// 			// require.Equal(t, tc.Resize, r.size)
// 		})
// 	}
// }
