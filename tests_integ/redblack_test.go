package memtable

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

func assert(t *testing.T, condi bool, msg string) {
	if !condi {
		t.Error(msg)
		t.Fail()
	}
}

func checkValidRecur(t *testing.T, bst *rbt.Tree, node *rbt.Node) {
	if node != nil && node.Key != nil {
		if node.Left != nil && node.Left.Key != nil {
			assert(t, node.Key.(int) >= node.Left.Key.(int), "key >= left key")
		}
		if node.Right != nil && node.Right.Key != nil {
			assert(t, node.Key.(int) <= node.Right.Key.(int), "key >= right key")
		}
	}

	if node == nil || (node.Left == nil && node.Right == nil) {
		return
	}

	checkValidRecur(t, bst, node.Left)
	checkValidRecur(t, bst, node.Right)
}

func TestBuildTree(t *testing.T) {
	file, err := os.Open("./memtable/data/large_input.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tree := rbt.NewWithIntComparator() // empty (keys are of type int)

	i := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		slime := strings.Split(scanner.Text(), " ")
		cmd := slime[0]

		value, err := strconv.ParseFloat(slime[1], 64)

		if err != nil {
			log.Fatal(err)
		}

		key := int(value)

		if cmd == "a" {
			tree.Put(key, key)
		} else if cmd == "r" {
			tree.Remove(key)
		}

		checkValidRecur(t, tree, tree.Root)

		i = i + 1
	}

	if tree.Size() != 1004 {
		t.Fail()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// func BenchmarkIntMin(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		IntMin(1, 2)
// 	}
// }
