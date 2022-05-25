package async

import (
	"context"
	"testing"
)

func BenchmarkName(b *testing.B) {
	partitionFunc := func(data animal) (string, bool) {
		if data.species == "" {
			return "", false
		}

		return data.species, true
	}

	p := NewPartitioner(context.Background(), partitionFunc)

	input1 := []animal{
		{"dog", "name1"},
		{"snail", "name2"},
		{"dog", "name4"},
		{"cat", "name5"},
	}

	input2 := []animal{
		{"dog", "name3"},
		{"cat", "name4"},
	}

	for i := 0; i < b.N; i++ {
		p.Take(input1...)
		p.Take(input2...)

		p.Outcome()
	}
}
