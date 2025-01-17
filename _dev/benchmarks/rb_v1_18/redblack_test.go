package tests_integ

import (
	"bufio"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

func assert(t *testing.B, condi bool, msg string) {
	if !condi {
		t.Error(msg)
		t.FailNow()
	}
}

func checkValidRecur(t *testing.B, bst *rbt.Tree, node *rbt.Node) {
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

func BenchmarkBuildRBTree(b *testing.B) {
	file, err := os.Open("../data/memtable/large_input.txt")
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

		checkValidRecur(b, tree, tree.Root)

		i = i + 1
	}

	if tree.Size() != 1004 {
		b.FailNow()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func TestBuildLargeRBTree(t *testing.T) {
	// let's store a tree of 128MB
	unitSize := int(unsafe.Sizeof(5) * 2)
	numberOfItems := (1.28e+8) / (unitSize)

	tree := rbt.NewWithIntComparator()

	// Act 1. Build Tree
	start := time.Now()
	var firstItem int
	for i := range numberOfItems {
		key := rand.IntN(9999999)

		if firstItem == 0 {
			firstItem = key
		}

		tree.Put(key, i)
	}
	t.Logf("Build Tree took %s", time.Since(start))

	// Act 2. Get from tree
	git := time.Now()
	val := tree.GetNode(firstItem)
	t.Logf("Get Tree took %s", time.Since(git))

	if val == nil {
		t.FailNow()
	}

	// Act 3. Put 1 item into tree
	pit := time.Now()
	tree.Put(rand.IntN(9999999), 123456)
	t.Logf("Put Tree took %s", time.Since(pit))
}
