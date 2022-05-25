// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type animal struct {
	species string
	name    string
}

func TestPartitioner(t *testing.T) {
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

	t1 := p.Take(input1...)
	t2 := p.Take(input2...)
	t1.Wait()
	t2.Wait()

	expected1 := map[string][]animal{
		"dog": {
			{"dog", "name1"},
			{"dog", "name4"},
			{"dog", "name3"},
		},
		"snail": {
			{"snail", "name2"},
		},
		"cat": {
			{"cat", "name5"},
			{"cat", "name4"},
		},
	}

	expected2 := map[string][]animal{
		"dog": {
			{"dog", "name3"},
			{"dog", "name1"},
			{"dog", "name4"},
		},
		"snail": {
			{"snail", "name2"},
		},
		"cat": {
			{"cat", "name4"},
			{"cat", "name5"},
		},
	}

	res := p.Outcome()
	assert.True(t, reflect.DeepEqual(expected1, res) || reflect.DeepEqual(expected2, res))
}

func ExamplePartitioner() {
	partitionFunc := func(data animal) (string, bool) {
		if data.species == "" {
			return "", false
		}

		return data.species, true
	}

	p := NewPartitioner(context.Background(), partitionFunc)

	input := []animal{
		{"dog", "name1"},
		{"snail", "name2"},
		{"dog", "name4"},
		{"cat", "name5"},
	}

	t := p.Take(input...)
	t.Wait()

	res := p.Outcome()
	first := res["dog"]
	fmt.Println(first[0])
	fmt.Println(first[1])

	// Output:
	// {dog name1}
	// {dog name4}
}
