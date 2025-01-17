package tests_integ

import (
	"bufio"
	"encoding/binary"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	rbt "github.com/emirpasic/gods/v2/trees/redblacktree"
)

func assert(t *testing.B, condi bool, msg string) {
	if !condi {
		t.Error(msg)
		t.Fail()
	}
}

func checkValidRecur(t *testing.B, bst *rbt.Tree[int, []byte], node *rbt.Node[int, []byte]) {
	if node != nil {
		if node.Left != nil {
			assert(t, node.Key >= node.Left.Key, "key >= left key")
		}
		if node.Right != nil {
			assert(t, node.Key <= node.Right.Key, "key >= right key")
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

	tree := rbt.New[int, []byte]()

	i := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		slime := strings.Split(scanner.Text(), " ")
		cmd := slime[0]

		fval, err := strconv.ParseFloat(slime[1], 64)

		if err != nil {
			log.Fatal(err)
		}

		key := int(fval)
		var value [8]byte
		binary.BigEndian.PutUint64(value[:], math.Float64bits(fval))

		if cmd == "a" {
			tree.Put(key, value[:])
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

	tree := rbt.New[int, int]()

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
