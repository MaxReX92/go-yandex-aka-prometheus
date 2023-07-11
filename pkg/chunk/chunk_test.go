package chunk

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

func TestChunk_SliceToChunks(t *testing.T) {
	tests := []struct {
		name      string
		chunkSize int
		slice     []int
		expected  [][]int
	}{
		{
			name:      "empty",
			chunkSize: 2,
			slice:     []int{},
			expected:  nil,
		},
		{
			name:      "success",
			chunkSize: 2,
			slice:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expected: [][]int{
				{0, 1},
				{2, 3},
				{4, 5},
				{6, 7},
				{8, 9},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SliceToChunks(tt.slice, tt.chunkSize)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestChunk_ChanToChunks(t *testing.T) {
	tests := []struct {
		name      string
		chunkSize int
		slice     []int
		expected  [][]int
	}{
		{
			name:      "empty",
			chunkSize: 2,
			slice:     []int{},
			expected:  nil,
		},
		{
			name:      "success",
			chunkSize: 2,
			slice:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expected: [][]int{
				{0, 1},
				{2, 3},
				{4, 5},
				{6, 7},
				{8, 9},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := test.ArrayToChan(tt.slice)
			actual := ChanToChunks(ch, tt.chunkSize)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
