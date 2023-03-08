package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/liwnn/bplustree"
)

type bplusTreeConfig struct {
	order   int
	entries int
}

func bplus_tree_setting(config *bplusTreeConfig) int {
	var ret int
	var again = true

	fmt.Fprintf(os.Stderr, "\n-- B+tree setting...\n")
	fmt.Fprintf(os.Stderr, "Set b+tree non-leaf order (3 < order <= %d e.g. 7): ", bplustree.BPLUS_MAX_ORDER)

	br := bufio.NewReader(os.Stdin)
	for again {
		i, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(os.Stderr, "\n")
				return -1
			}
		}
		switch i {
		case 'q':
			return -1
		case '\n':
			config.order = 7
			again = false
		default:
			if err := br.UnreadByte(); err != nil {
				panic(err)
			}
			_, err := fmt.Fscanf(br, "%d", &config.order)
			if err != nil {
				_, _ = br.ReadBytes('\n')
				again = true
			} else {
				c, err := br.ReadByte()
				if err != nil {
					panic(err)
				}
				if c != '\n' {
					_, _ = br.ReadBytes('\n')
					again = true
				} else if config.order < 3 || config.order > bplustree.BPLUS_MAX_ORDER {
					again = true
				} else {
					again = false
				}
			}
		}
	}

	again = true
	fmt.Fprintf(os.Stderr, "Set b+tree leaf entries (<= %d e.g. 10): ", bplustree.BPLUS_MAX_ENTRIES)
	for again {
		i, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(os.Stderr, "\n")
				return -1
			}
		}
		switch i {
		case 'q':
			return -1
		case '\n':
			config.entries = 10
			again = false
		default:
			if err := br.UnreadByte(); err != nil {
				panic(err)
			}
			_, err := fmt.Fscanf(br, "%d", &config.entries)
			if err != nil {
				_, _ = br.ReadBytes('\n')
				again = true
			} else {
				c, err := br.ReadByte()
				if err != nil {
					panic(err)
				}
				if c != '\n' {
					_, _ = br.ReadBytes('\n')
					again = true
				} else if config.entries > bplustree.BPLUS_MAX_ENTRIES {
					again = true
				} else {
					again = false
				}
			}
		}
	}

	return ret
}

func _proc(tree *bplustree.BPlusTree, op byte, n int) {
	switch op {
	case 'i':
		tree.Insert(n, n)
	case 'r':
		tree.Delete(n)
	case 's':
		data, found := tree.Search(n)
		fmt.Fprintf(os.Stderr, "key:%d data:%d(found=%v)\n", n, data, found)
	default:
	}
}

func number_process(br *bufio.Reader, tree *bplustree.BPlusTree, op byte) int {
	var n int
	var start, end int

	for {
		c, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		if c == ' ' || c == '\t' || c == '\n' {
			if start != 0 {
				if n >= 0 {
					end = n
				} else {
					n = 0
				}
			}

			if start != 0 && end != 0 {
				if start <= end {
					for n = start; n <= end; n++ {
						_proc(tree, op, n)
					}
				} else {
					for n = start; n >= end; n-- {
						_proc(tree, op, n)
					}
				}
			} else {
				if n != 0 {
					_proc(tree, op, n)
				}
			}

			n = 0
			start = 0
			end = 0

			if c == '\n' {
				return 0
			} else {
				continue
			}
		}

		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else if c == '-' && n != 0 {
			start = n
			n = 0
		} else {
			n = 0
			start = 0
			end = 0
			for {
				c, err := br.ReadByte()
				if err != nil {
					break
				}
				if c != ' ' && c != '\t' && c != '\n' {
					continue
				} else {
					break
				}
			}
			_ = br.UnreadByte()
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
	return -1
}

func command_tips() {
	fmt.Fprintf(os.Stderr, "i: Insert key. e.g. i 1 4-7 9\n")
	fmt.Fprintf(os.Stderr, "r: Remove key. e.g. r 1-100\n")
	fmt.Fprintf(os.Stderr, "s: Search by key. e.g. s 41-60\n")
	fmt.Fprintf(os.Stderr, "d: Dump the tree structure.\n")
	fmt.Fprintf(os.Stderr, "q: quit.\n")
}

func command_process(tree *bplustree.BPlusTree) {
	br := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "Please input command (Type 'h' for help): ")
	for {
		c, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(os.Stderr, "\n")
			} else {
				panic(err)
			}
		}
		switch c {
		case 'q':
			return
		case 'h':
			command_tips()
		case 'd':
			bplustree.Dump(tree)
		case 'i', 'r', 's':
			if number_process(br, tree, c) < 0 {
				return
			}
		case '\n':
			fmt.Fprintf(os.Stderr, "Please input command (Type 'h' for help): ")
		default:
		}
	}
}

func main() {
	var tree *bplustree.BPlusTree
	var config bplusTreeConfig

	/* B+tree default setting */
	if bplus_tree_setting(&config) < 0 {
		return
	}

	/* Init b+tree */
	tree = bplustree.New(config.order, config.entries)
	if tree == nil {
		fmt.Fprintf(os.Stderr, "Init failure!\n")
		os.Exit(-1)
	}

	/* Operation process */
	command_process(tree)

	/* Deinit b+tree */
	// bplus_tree_deinit(tree)
}
