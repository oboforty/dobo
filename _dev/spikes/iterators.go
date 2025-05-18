package main

import (
	"iter"
	"strconv"
)

func main() {
	input := []string{"1", "2", "3", "5", "some word", "8"}

	it := newAtoiIterator(input)

	var nums []int
	for num := range it {
		nums = append(nums, num)
	}
}

func newAtoiIterator(vals []string) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, v := range vals {
			num, _ := strconv.Atoi(v)
			if !yield(num) {
				return
			}
		}
	}
}
