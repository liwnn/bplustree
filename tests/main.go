package main

import (
	"fmt"
	"os"

	"github.com/liwnn/bplustree"
)

type bplusTreeConfig struct {
	order   int
	entries int
}

func get_put_test(tree *bplustree.BPlusTree) {
	var i int

	fmt.Fprintf(os.Stderr, "\n> B+tree getter and setter testing...\n")

	tree.Insert(24, 24)
	tree.Insert(72, 72)
	tree.Insert(1, 1)
	tree.Insert(39, 39)
	tree.Insert(53, 53)
	tree.Insert(63, 63)
	tree.Insert(90, 90)
	tree.Insert(88, 88)
	tree.Insert(15, 15)
	tree.Insert(10, 10)
	tree.Insert(44, 44)
	tree.Insert(68, 68)
	tree.Insert(74, 74)
	bplustree.Dump(tree)

	tree.Insert(10, 10)
	tree.Insert(15, 15)
	tree.Insert(18, 18)
	tree.Insert(22, 22)
	tree.Insert(27, 27)
	tree.Insert(34, 34)
	tree.Insert(40, 40)
	tree.Insert(44, 44)
	tree.Insert(47, 47)
	tree.Insert(54, 54)
	tree.Insert(67, 67)
	tree.Insert(72, 72)
	tree.Insert(74, 74)
	tree.Insert(78, 78)
	tree.Insert(81, 81)
	tree.Insert(84, 84)
	bplustree.Dump(tree)

	var nums = []int{24, 72, 1, 39, 53, 63, 90, 88, 15, 10, 44, 68}
	for _, n := range nums {
		data, found := tree.Search(n)
		if !found {
			panic("not found")
		}
		fmt.Fprintf(os.Stderr, "key:%v data:%d\n", n, data)
	}

	/* Not found */
	_, found := tree.Search(100)
	if found {
		panic("100 must not found")
	}
	fmt.Fprintf(os.Stderr, "key:100 found:%v\n", found)

	/* Clear all */
	fmt.Fprintf(os.Stderr, "\n> Clear all...\n")
	for i = 1; i <= 100; i++ {
		tree.Delete(i)
	}
	bplustree.Dump(tree)

	/* Not found */
	_, found = tree.Search(100)
	if found {
		panic("100 must not found")
	}
	fmt.Fprintf(os.Stderr, "key:100 found:%v\n", found)
}

func bplus_tree_insert_delete_test(tree *bplustree.BPlusTree) {
	var i int
	var max_key = 100

	fmt.Fprintf(os.Stderr, "\n> B+tree insertion and deletion testing...\n")

	/* Ordered insertion and deletion */
	fmt.Fprintf(os.Stderr, "\n-- Insert 1 to %d, dump:\n", max_key)
	for i = 1; i <= max_key; i++ {
		tree.Insert(i, i)
	}
	bplustree.Dump(tree)

	fmt.Fprintf(os.Stderr, "\n-- Delete 1 to %d, dump:\n", max_key)
	for i = 1; i <= max_key; i++ {
		tree.Delete(i)
	}
	bplustree.Dump(tree)

	/* Ordered insertion and reversed deletion */
	fmt.Fprintf(os.Stderr, "\n-- Insert 1 to %d, dump:\n", max_key)
	for i = 1; i <= max_key; i++ {
		tree.Insert(i, i)
	}
	bplustree.Dump(tree)

	fmt.Fprintf(os.Stderr, "\n-- Delete %d to 1, dump:\n", max_key)
	for i--; i > 0; i-- {
		tree.Delete(i)
	}
	bplustree.Dump(tree)

	/* Reversed insertion and ordered deletion */
	fmt.Fprintf(os.Stderr, "\n-- Insert %d to 1, dump:\n", max_key)
	for i = max_key; i > 0; i-- {
		tree.Insert(i, i)
	}
	bplustree.Dump(tree)

	fmt.Fprintf(os.Stderr, "\n-- Delete 1 to %d, dump:\n", max_key)
	for i = 1; i <= max_key; i++ {
		tree.Delete(i)
	}
	bplustree.Dump(tree)

	/* Reversed insertion and reversed deletion */
	fmt.Fprintf(os.Stderr, "\n-- Insert %d to 1, dump:\n", max_key)
	for i = max_key; i > 0; i-- {
		tree.Insert(i, i)
	}
	bplustree.Dump(tree)

	fmt.Fprintf(os.Stderr, "\n-- Delete %d to 1, dump:\n", max_key)
	for i = max_key; i > 0; i-- {
		var n = 98
		if i == n {
			bplustree.Dump(tree)
		}
		tree.Delete(i)
		if i == n {
			bplustree.Dump(tree)
		}
	}
	bplustree.Dump(tree)
}

func bplus_tree_normal_test() {
	var tree *bplustree.BPlusTree
	var config bplusTreeConfig

	fmt.Fprintf(os.Stderr, "\n>>> B+tree normal test.\n")

	/* Init b+tree */
	config.order = 7
	config.entries = 10
	tree = bplustree.New(config.order, config.entries)
	if tree == nil {
		fmt.Fprintf(os.Stderr, "Init failure!\n")
		os.Exit(-1)
	}

	/* getter and setter test */
	get_put_test(tree)

	/* insertion and deletion test */
	bplus_tree_insert_delete_test(tree)

	/* Deinit b+tree */
	// bplus_tree_deinit(tree)
}

func main() {
	bplus_tree_normal_test()
	//bplus_tree_abnormal_test();
}
