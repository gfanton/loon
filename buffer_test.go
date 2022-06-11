package main

import (
	"container/ring"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBufferAddValue(t *testing.T) {
	cases := []struct {
		Name             string
		Size             int
		Values, Expected []int
	}{
		{"0 value : size 1", 1, []int{}, []int{}},
		{"0 value : size 10", 10, []int{}, []int{}},
		{"5 values : size 1", 1, []int{1, 2, 3, 4, 5}, []int{5}},

		{"add 1 value : size 10", 10,
			[]int{1},
			[]int{1},
		},

		{"add 5 values : size 5", 5,
			[]int{1, 2, 3, 4, 5},
			[]int{1, 2, 3, 4, 5},
		},

		{"add 7 values : size : 5", 5,
			[]int{1, 2, 3, 4, 5, 6, 7},
			[]int{3, 4, 5, 6, 7},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			buff := NewBuffer[*int](tc.Size)
			for _, v := range tc.Values {
				i := v
				buff.AddValue(&i)
			}

			var nempty int
			if tc.Size > len(tc.Expected) {
				nempty = tc.Size - len(tc.Expected)
			}

			values := []int{}
			buff.ring.Do(func(v any) {
				if v == nil {
					nempty--
				} else {
					values = append(values, *(v.(*int)))
				}
			})

			require.Equal(t, 0, nempty)
			require.Equal(t, tc.Expected, values)
			require.Equal(t, uint(len(tc.Values)), buff.Lines())
		})
	}
}

func TestBufferDoPrev(t *testing.T) {
	t.Run("buffer no values", func(t *testing.T) {
		buff := NewBuffer[int](10)

		i := 0
		buff.DoPrev(func(r *ring.Ring, j int) bool {
			i++
			return true
		})
		require.Equal(t, 0, i)
	})

	t.Run("buffer half full", func(t *testing.T) {
		const size = 10
		hsize := size / 2

		buff := NewBuffer[int](10)

		expected := []int{}
		for i := 0; i < hsize; i++ {
			buff.AddValue(i)
			expected = append(expected, hsize-i-1)
		}

		values := []int{}
		buff.DoPrev(func(r *ring.Ring, j int) bool {
			require.NotNil(t, r)
			require.NotNil(t, r.Value)
			values = append(values, j)
			return true
		})
		require.Equal(t, expected, values)
		require.Equal(t, hsize, len(values))
		require.Equal(t, uint(hsize), buff.Lines())
	})

	t.Run("buffer full", func(t *testing.T) {
		const size = 10

		buff := NewBuffer[int](10)

		expected := []int{}
		for i := 0; i < size; i++ {
			buff.AddValue(i)
			expected = append(expected, size-i-1)
		}

		values := []int{}
		buff.DoPrev(func(r *ring.Ring, j int) bool {
			require.NotNil(t, r)
			require.NotNil(t, r.Value)
			values = append(values, j)
			return true
		})

		require.Equal(t, expected, values)
		require.Equal(t, size, len(values))
		require.Equal(t, uint(size), buff.Lines())
	})
}
