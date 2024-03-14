package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunk(t *testing.T) {
	type args struct {
		slice []int
		size  int
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "empty",
			args: args{
				slice: []int{},
				size:  1,
			},
			want: nil,
		},
		{
			name: "one",
			args: args{
				slice: []int{1},
				size:  1,
			},
			want: [][]int{{1}},
		},
		{
			name: "two",
			args: args{
				slice: []int{1, 2},
				size:  1,
			},
			want: [][]int{{1}, {2}},
		},
		{
			name: "two",
			args: args{
				slice: []int{1, 2},
				size:  2,
			},
			want: [][]int{{1, 2}},
		},
		{
			name: "three",
			args: args{
				slice: []int{1, 2, 3},
				size:  2,
			},
			want: [][]int{{1, 2}, {3}},
		},
		{
			name: "four",
			args: args{
				slice: []int{1, 2, 3, 4},
				size:  2,
			},
			want: [][]int{{1, 2}, {3, 4}},
		},
		{
			name: "five",
			args: args{
				slice: []int{1, 2, 3, 4, 5},
				size:  2,
			},
			want: [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name: "six",
			args: args{
				slice: []int{1, 2, 3, 4, 5, 6},
				size:  2,
			},
			want: [][]int{{1, 2}, {3, 4}, {5, 6}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Chunk(tt.args.slice, tt.args.size)
			assert.Equal(t, tt.want, got)
		})
	}
}
