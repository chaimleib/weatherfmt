package typedmap_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chaimleib/weatherfmt/typedmap"
)

func TestTypedMap_IntMap(t *testing.T) {
	type Case struct {
		name   string
		modify func(m *typedmap.TypedMap[int])
		want   map[string]int
	}
	cases := []Case{
		{
			name: "empty",
		},
		{
			name: "store a:1",
			modify: func(m *typedmap.TypedMap[int]) {
				m.Store("a", 1)
			},
			want: map[string]int{"a": 1},
		},
		{
			name: "swap b:2 for empty",
			modify: func(m *typedmap.TypedMap[int]) {
				m.SwapFunc("b", func(key string, old int, ok bool) (int, error) {
					return 2, nil
				})
			},
			want: map[string]int{"b": 2},
		},
		{
			name: "swap c:3 for present",
			modify: func(m *typedmap.TypedMap[int]) {
				m.SwapFunc("c", func(key string, old int, ok bool) (int, error) {
					return 3, nil
				})
			},
			want: map[string]int{"c": 3},
		},
		{
			name: "swap d:4 for present",
			modify: func(m *typedmap.TypedMap[int]) {
				m.SwapFunc("c", func(key string, old int, ok bool) (int, error) {
					return 3, nil
				})
			},
			want: map[string]int{"c": 3},
		},
		{
			name: "cancel swap with error",
			modify: func(m *typedmap.TypedMap[int]) {
				m.SwapFunc("c", func(key string, old int, ok bool) (int, error) {
					return 0, errors.New("cancel swap")
				})
			},
			want: nil,
		},
		{
			name: "cancel swap with error, even with nonzero return",
			modify: func(m *typedmap.TypedMap[int]) {
				m.SwapFunc("c", func(key string, old int, ok bool) (int, error) {
					return 1, errors.New("cancel swap")
				})
			},
			want: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := typedmap.New[int]()
			if c.modify != nil {
				c.modify(m)
			}

			keys := m.Keys()
			if len(c.want) == 0 {
				assert.Empty(t, keys)
				return
			}

			gotMap := make(map[string]int)
			for _, k := range keys {
				v, ok := m.Load(k)
				if !ok {
					t.Fatalf("failed to Load(%q)", k)
				}
				gotMap[k] = v
			}
			assert.Equal(t, c.want, gotMap)
		})
	}
}
