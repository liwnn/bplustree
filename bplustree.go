package bplustree

import (
	"fmt"
)

const (
	BPLUS_MAX_ORDER   = 256
	BPLUS_MAX_ENTRIES = 512
	BPLUS_MAX_LEVEL   = 10
)

const (
	LEFT_SIBLING  = 0
	RIGHT_SIBLING = 1
)

const (
	BPLUS_TREE_LEAF     = 0
	BPLUS_TREE_NON_LEAF = 1
)

type key_t = int
type value_t = int

type bplusNode struct {
	typ          int           // BPLUS_TREE_LEAF is 0 and  BPLUS_TREE_NON_LEAF is 1
	parentKeyIdx int           // index of parent node
	parent       *bplusNonLeaf // piointer to parent node
}

type node interface{}

func getNode(n node) *bplusNode {
	if v, ok := n.(*bplusLeaf); ok {
		return &v.bplusNode
	} else {
		return &n.(*bplusNonLeaf).bplusNode
	}
}

type bplusNonLeaf struct {
	bplusNode
	prev, next *bplusNonLeaf
	/**  number of child node */
	children int
	/**  key array */
	key [BPLUS_MAX_ORDER - 1]key_t
	/** pointers to child node */
	subPtr [BPLUS_MAX_ORDER]node
}

func (nl *bplusNonLeaf) keySearch(target key_t) (int, bool) {
	i, j := 0, nl.children-1
	for i < j {
		h := int(uint(i+j) >> 1)
		if nl.key[h] <= target {
			i = h + 1
		} else {
			j = h
		}
	}

	if i > 0 && nl.key[i-1] == target {
		return i - 1, true
	}
	return i, false
}

func (nl *bplusNonLeaf) listAdd(link *bplusNonLeaf, next *bplusNonLeaf) {
	link.next = next
	link.prev = nl
	next.prev = link
	nl.next = link
}

func (nl *bplusNonLeaf) simpleRemove(remove int) {
	assert(nl.children >= 2)
	copy(nl.key[remove:], nl.key[remove+1:nl.children-1])
	copy(nl.subPtr[remove+1:], nl.subPtr[remove+2:nl.children])
	nl.children--
	// for gc
	nl.subPtr[nl.children] = nil

	if _, ok := nl.subPtr[0].(*bplusLeaf); ok {
		for i := remove + 1; i < nl.children; i++ {
			nl.subPtr[i].(*bplusLeaf).parentKeyIdx = i - 1
		}
	} else {
		for i := remove + 1; i < nl.children; i++ {
			nl.subPtr[i].(*bplusNonLeaf).parentKeyIdx = i - 1
		}
	}
}

func (nl *bplusNonLeaf) simpleInsert(lch node, rch node, key key_t, insert int) {
	copy(nl.key[insert+1:], nl.key[insert:nl.children-1])
	copy(nl.subPtr[insert+2:], nl.subPtr[insert+1:nl.children])
	nl.key[insert] = key
	nl.subPtr[insert] = lch
	nl.subPtr[insert+1] = rch
	nl.children++

	if _, ok := lch.(*bplusNonLeaf); ok {
		for i := insert; i < nl.children; i++ {
			nl.subPtr[i].(*bplusNonLeaf).parentKeyIdx = i - 1
		}
	} else {
		for i := insert; i < nl.children; i++ {
			nl.subPtr[i].(*bplusLeaf).parentKeyIdx = i - 1
		}
	}
}

func (nl *bplusNonLeaf) shiftFromLeft(left *bplusNonLeaf, parentKeyIndex int, remove int) {
	/* node's elements right shift */
	copy(nl.key[1:remove+1], nl.key[0:remove])
	copy(nl.subPtr[1:], nl.subPtr[0:remove+1])
	/* parent key right rotation */
	nl.key[0] = nl.parent.key[parentKeyIndex]
	nl.parent.key[parentKeyIndex] = left.key[left.children-2]
	/* borrow the last sub-node from left sibling */
	nl.subPtr[0] = left.subPtr[left.children-1]
	left.children--

	for i := remove + 1; i > 0; i-- {
		getNode(nl.subPtr[i]).parentKeyIdx = i - 1
	}
	bn := getNode(nl.subPtr[0])
	bn.parent = nl
	bn.parentKeyIdx = -1
}

func (nl *bplusNonLeaf) shiftFromRight(right *bplusNonLeaf, parentKeyIndex int) {
	/* parent key left rotation */
	nl.key[nl.children-1] = nl.parent.key[parentKeyIndex]
	nl.parent.key[parentKeyIndex] = right.key[0]
	/* borrow the first sub-node from right sibling */
	nl.subPtr[nl.children] = right.subPtr[0]
	bn := getNode(right.subPtr[0])
	bn.parent = nl
	bn.parentKeyIdx = nl.children - 1
	nl.children++
	/* left shift in right sibling */
	copy(right.key[0:], right.key[1:right.children-1])
	copy(right.subPtr[0:], right.subPtr[1:right.children])
	for i := 0; i < right.children-1; i++ {
		getNode(right.subPtr[i]).parentKeyIdx = i - 1
	}
	right.children--
}

func (nl *bplusNonLeaf) mergeFromRight(right *bplusNonLeaf, parentKeyIndex int) {
	/* move parent key down */
	nl.key[nl.children-1] = nl.parent.key[parentKeyIndex]
	/* merge from right sibling */
	copy(nl.key[nl.children:], right.key[:right.children-1])
	copy(nl.subPtr[nl.children:], right.subPtr[:right.children])
	for i, j := nl.children, 0; j < right.children; j++ {
		bn := getNode(nl.subPtr[i])
		bn.parent = nl
		bn.parentKeyIdx = i - 1
		i++
	}
	nl.children += right.children
	/* delete empty right sibling */
	right.delete()
}

func (nl *bplusNonLeaf) mergeIntoLeft(left *bplusNonLeaf, parentKeyIndex int, remove int) {
	/* move parent key down */
	left.key[left.children-1] = nl.parent.key[parentKeyIndex]
	/* merge into left sibling */
	copy(left.key[left.children:], nl.key[:remove])
	copy(left.key[left.children+remove:], nl.key[remove+1:nl.children-1])

	copy(left.subPtr[left.children:], nl.subPtr[:remove+1])
	copy(left.subPtr[left.children+remove+1:], nl.subPtr[remove+2:nl.children])

	var i, j int
	for i, j = left.children, 0; j < nl.children-1; j++ {
		bn := getNode(left.subPtr[i])
		bn.parent = left
		bn.parentKeyIdx = i - 1
		i++
	}
	left.children = i
	/* delete empty node */
	nl.delete()
}

func (nl *bplusNonLeaf) delete() {
	nl.prev.next = nl.next
	nl.next.prev = nl.prev
	// TODO: free node
}

func (nl *bplusNonLeaf) splitLeft(left *bplusNonLeaf, lCh node, rCh node, key key_t, insert int, split int) key_t {
	var order = nl.children
	var split_key key_t
	/* split as left sibling */
	nl.prev.listAdd(left, nl)
	/* replicate from sub[0] to sub[split] */
	copy(left.subPtr[:], nl.subPtr[:insert])
	copy(left.subPtr[insert+2:], nl.subPtr[insert+1:split+1])
	left.subPtr[insert] = lCh
	left.subPtr[insert+1] = rCh
	left.children = split + 1
	for i := 0; i < left.children; i++ {
		bn := getNode(left.subPtr[i])
		bn.parent = left
		bn.parentKeyIdx = i - 1
	}
	/* replicate from key[0] to key[split - 1] */
	copy(left.key[:], nl.key[:insert])
	copy(left.key[insert+1:], nl.key[insert:split-1])
	left.key[insert] = key

	var i, j int
	if insert == split {
		left.key[insert] = key
		left.subPtr[insert] = lCh
		lbn := getNode(lCh)
		lbn.parent = left
		lbn.parentKeyIdx = j - 1
		nl.subPtr[0] = rCh
		split_key = key
	} else {
		nl.subPtr[0] = nl.subPtr[split]
		split_key = nl.key[split]
	}
	sbn := getNode(nl.subPtr[0])
	sbn.parent = nl
	sbn.parentKeyIdx = -1
	/* left shift for right node from split to children - 1 */
	for i, j = split, 0; i < order-1; {
		nl.key[j] = nl.key[i]
		nl.subPtr[j+1] = nl.subPtr[i+1]
		bn := getNode(nl.subPtr[j+1])
		bn.parent = nl
		bn.parentKeyIdx = j

		i++
		j++
	}
	nl.subPtr[j] = nl.subPtr[i]
	nl.children = j + 1
	return split_key
}

func (nl *bplusNonLeaf) splitRight1(right *bplusNonLeaf, l_ch, r_ch node, key key_t, insert int, split int) key_t {
	var i, j int
	var order = nl.children
	var split_key key_t
	/* split as right sibling */
	nl.listAdd(right, nl.next)
	/* split key is key[split - 1] */
	split_key = nl.key[split-1]
	/* left node's children always be [split] */
	nl.children = split
	/* right node's first sub-node */
	right.key[0] = key
	right.subPtr[0] = l_ch
	lbn := getNode(l_ch)
	lbn.parent = right
	lbn.parentKeyIdx = -1
	right.subPtr[1] = r_ch
	rbn := getNode(r_ch)
	rbn.parent = right
	rbn.parentKeyIdx = 0
	/* insertion point is split point, replicate from key[split] */
	for i, j = split, 1; i < order-1; {
		right.key[j] = nl.key[i]
		right.subPtr[j+1] = nl.subPtr[i+1]
		rcbn := getNode(right.subPtr[j+1])
		rcbn.parent = right
		rcbn.parentKeyIdx = j

		i++
		j++
	}
	right.children = j + 1
	return split_key
}

func (nl *bplusNonLeaf) splitRight2(right *bplusNonLeaf, lCh, rCh node, key key_t, insert int, split int) key_t {
	var i, j int
	var order = nl.children
	var split_key key_t
	/* left node's children always be [split + 1] */
	nl.children = split + 1
	/* split as right sibling */
	nl.listAdd(right, nl.next)
	/* split key is key[split] */
	split_key = nl.key[split]
	/* right node's first sub-node */
	right.subPtr[0] = nl.subPtr[split+1]
	sn := getNode(right.subPtr[0])
	sn.parent = right
	sn.parentKeyIdx = -1
	/* replicate from key[split + 1] to key[order - 1] */
	for i, j = split+1, 0; i < order-1; j++ {
		if j != insert-split-1 {
			right.key[j] = nl.key[i]
			right.subPtr[j+1] = nl.subPtr[i+1]

			bn := getNode(right.subPtr[j+1])
			bn.parent = right
			bn.parentKeyIdx = j
			i++
		}
	}
	/* reserve a hole for insertion */
	if j > insert-split-1 {
		right.children = j + 1
	} else {
		assert(j == insert-split-1)
		right.children = j + 2
	}
	/* insert new key and sub-node */
	j = insert - split - 1
	right.key[j] = key
	right.subPtr[j] = lCh
	lbn := getNode(lCh)
	lbn.parent = right
	lbn.parentKeyIdx = j - 1
	right.subPtr[j+1] = rCh
	rbn := getNode(rCh)
	rbn.parent = right
	rbn.parentKeyIdx = j
	return split_key
}

type bplusLeaf struct {
	bplusNode
	/** pointer to first node(head) in leaf linked list
	 */
	prev, next *bplusLeaf
	/** number of actual key-value pairs in leaf node */
	entries int
	/**  key array */
	kvs [BPLUS_MAX_ENTRIES]struct {
		key   key_t
		value value_t
	}
}

func (leaf *bplusLeaf) keySearch(target key_t) (int, bool) {
	i, j := 0, leaf.entries
	for i < j {
		h := int(uint(i+j) >> 1)
		if leaf.kvs[h].key <= target {
			i = h + 1
		} else {
			j = h
		}
	}

	if i > 0 && leaf.kvs[i-1].key == target {
		return i - 1, true
	}
	return i, false
}

func (leaf *bplusLeaf) listAdd(link *bplusLeaf, next *bplusLeaf) {
	link.next = next
	link.prev = leaf
	next.prev = link
	leaf.next = link
}

func (leaf *bplusLeaf) simpleInsert(key key_t, data value_t, insert int) {
	copy(leaf.kvs[insert+1:], leaf.kvs[insert:leaf.entries])
	leaf.kvs[insert].key = key
	leaf.kvs[insert].value = data
	leaf.entries++
}

func (leaf *bplusLeaf) simpleRemove(remove int) {
	copy(leaf.kvs[remove:], leaf.kvs[remove+1:leaf.entries])
	leaf.entries--
}

func (leaf *bplusLeaf) mergeFromRight(right *bplusLeaf) {
	/* merge from right sibling */
	copy(leaf.kvs[leaf.entries:], right.kvs[:right.entries])
	leaf.entries += right.entries
	/* delete right sibling */
	leafDelete(right)
}

func (leaf *bplusLeaf) splitLeft(left *bplusLeaf, key key_t, data value_t, insert int) {
	/* split = [m/2] */
	split := (leaf.entries + 1) / 2
	/* split as left sibling */
	leaf.prev.listAdd(left, leaf)
	/* replicate from 0 to key[split - 2] */
	copy(left.kvs[:], leaf.kvs[:insert])
	left.kvs[insert].key = key
	left.kvs[insert].value = data
	copy(left.kvs[insert+1:], leaf.kvs[insert:split-1])
	left.entries = split

	/* left shift for right node */
	leaf.entries = copy(leaf.kvs[:], leaf.kvs[split-1:leaf.entries])
}

func (leaf *bplusLeaf) shiftFromRight(right *bplusLeaf, parentKeyIndex int) {
	/* borrow the first element from right sibling */
	leaf.kvs[leaf.entries] = right.kvs[0]
	leaf.entries++
	/* left shift in right sibling */
	copy(right.kvs[0:], right.kvs[1:right.entries])
	right.entries--
	/* update parent key */
	leaf.parent.key[parentKeyIndex] = right.kvs[0].key
}

func (leaf *bplusLeaf) shiftFromLeft(left *bplusLeaf, parentKeyIndex int, remove int) {
	/* right shift in leaf node */
	copy(leaf.kvs[1:remove+1], leaf.kvs[0:remove])
	/* borrow the last element from left sibling */
	left.entries--
	leaf.kvs[0] = left.kvs[left.entries]
	/* update parent key */
	leaf.parent.key[parentKeyIndex] = leaf.kvs[0].key
}

func (leaf *bplusLeaf) mergeIntoLeft(left *bplusLeaf, remove int) {
	/* merge into left sibling */
	left.entries += copy(left.kvs[left.entries:], leaf.kvs[0:remove])
	left.entries += copy(left.kvs[left.entries:], leaf.kvs[remove+1:leaf.entries])
	/* delete merged leaf */
	leafDelete(leaf)
}

func (leaf *bplusLeaf) splitRight(right *bplusLeaf, key key_t, data value_t, insert int) {
	/* split = [m/2] */
	split := (leaf.entries + 1) / 2
	/* split as right sibling */
	leaf.listAdd(right, leaf.next)
	/* replicate from key[split] */
	j := insert - split

	copy(right.kvs[:], leaf.kvs[split:insert])
	copy(right.kvs[j+1:], leaf.kvs[insert:leaf.entries])

	/* reserve a hole for insertion */
	right.entries = leaf.entries - split + 1

	/* insert new key */
	right.kvs[j].key = key
	right.kvs[j].value = data
	/* left leaf number */
	leaf.entries = split
}

type BPlusTree struct {
	/**  The actual number of children for a node, referred to here as order */
	order int
	/** number of actual key-value pairs in tree */
	entries int
	/** height of the tree */
	level int
	root  node

	firstLeaf    *bplusLeaf
	firstNonLeaf [BPLUS_MAX_LEVEL]*bplusNonLeaf
}

func assert(ok bool) {
	if !ok {
		panic("ok")
	}
}

func New(order int, entries int) *BPlusTree {
	/* The max order of non leaf nodes must be more than two */
	assert(order <= BPLUS_MAX_ORDER && entries <= BPLUS_MAX_ENTRIES)

	tree := new(BPlusTree)
	tree.root = nil
	tree.order = order
	tree.entries = entries
	return tree
}

func leafNew() *bplusLeaf {
	leaf := new(bplusLeaf)
	leaf.prev = leaf
	leaf.next = leaf
	leaf.typ = BPLUS_TREE_LEAF
	leaf.parentKeyIdx = -1
	return leaf
}

func nonLeafNew() *bplusNonLeaf {
	nonLeaf := new(bplusNonLeaf)
	nonLeaf.prev = nonLeaf
	nonLeaf.next = nonLeaf
	nonLeaf.typ = BPLUS_TREE_NON_LEAF
	nonLeaf.parentKeyIdx = -1
	return nonLeaf
}

func nonLeafSiblingSelect(lsib *bplusNonLeaf, rsib *bplusNonLeaf, parent *bplusNonLeaf, i int) int {
	if i == -1 {
		/* the frist sub-node, no left sibling, choose the right one */
		return RIGHT_SIBLING
	} else if i == parent.children-2 {
		/* the last sub-node, no right sibling, choose the left one */
		return LEFT_SIBLING
	} else {
		/* if both left and right sibling found, choose the one with more entries */
		if lsib.children >= rsib.children {
			return LEFT_SIBLING
		}
		return RIGHT_SIBLING
	}
}

func leafDelete(node *bplusLeaf) {
	node.prev.next = node.next
	node.next.prev = node.prev

	// TODO: free node
}

func leafSiblingSelect(lsib *bplusLeaf, rsib *bplusLeaf, parent *bplusNonLeaf, i int) int {
	if i == -1 {
		/* the first sub-node, no left sibling, choose the right one */
		return RIGHT_SIBLING
	} else if i == parent.children-2 {
		/* the last sub-node, no right sibling, choose the left one */
		return LEFT_SIBLING
	} else {
		/* if both left and right sibling found, choose the one with more entries */
		if lsib.entries >= rsib.entries {
			return LEFT_SIBLING
		}
		return RIGHT_SIBLING
	}
}

func (tree *BPlusTree) parentNodeBuild(left node, right node, key key_t, level int) int {
	ln := getNode(left)
	rn := getNode(right)
	if ln.parent == nil && rn.parent == nil {
		/* new parent */
		parent := nonLeafNew()
		parent.key[0] = key
		parent.subPtr[0] = left
		ln.parent = parent
		ln.parentKeyIdx = -1
		parent.subPtr[1] = right
		rn.parent = parent
		rn.parentKeyIdx = 0
		parent.children = 2
		/* update root */
		tree.root = parent
		tree.level++

		tree.firstNonLeaf[tree.level] = parent
		return 0
	} else if rn.parent == nil {
		/* trace upwards */
		rn.parent = ln.parent
		return tree.nonLeafInsert(ln.parent, left, right, key, level+1)
	} else {
		/* trace upwards */
		ln.parent = rn.parent
		return tree.nonLeafInsert(rn.parent, left, right, key, level+1)
	}
}

func (tree *BPlusTree) nonLeafInsert(node *bplusNonLeaf, lCh node, rCh node, key key_t, level int) int {
	/* search key location */
	insert, ok := node.keySearch(key)
	assert(!ok)

	/* node is full */
	if node.children == tree.order {
		/* split = [m/2] */
		var splitKey key_t
		split := node.children / 2
		sibling := nonLeafNew()
		if insert < split {
			splitKey = node.splitLeft(sibling, lCh, rCh, key, insert, split)
		} else if insert == split {
			splitKey = node.splitRight1(sibling, lCh, rCh, key, insert, split)
		} else {
			splitKey = node.splitRight2(sibling, lCh, rCh, key, insert, split)
		}
		/* build new parent */
		if insert < split {
			return tree.parentNodeBuild(sibling, node, splitKey, level)
		} else {
			return tree.parentNodeBuild(node, sibling, splitKey, level)
		}
	} else {
		node.simpleInsert(lCh, rCh, key, insert)
	}
	return 0
}

func (tree *BPlusTree) leafInsert(leaf *bplusLeaf, key key_t, data value_t) int {
	/* search key location */
	insert, ok := leaf.keySearch(key)
	if ok {
		/* Already exists */
		return -1
	}

	/* node full */
	if leaf.entries == tree.entries {
		/* split = [m/2] */
		split := (tree.entries + 1) / 2
		/* splited sibling node */
		sibling := leafNew()
		/* sibling leaf replication due to location of insertion */
		if insert < split {
			leaf.splitLeft(sibling, key, data, insert)
		} else {
			leaf.splitRight(sibling, key, data, insert)
		}
		/* build new parent */
		if insert < split {
			return tree.parentNodeBuild(sibling, leaf, leaf.kvs[0].key, 0)
		} else {
			return tree.parentNodeBuild(leaf, sibling, sibling.kvs[0].key, 0)
		}
	} else {
		leaf.simpleInsert(key, data, insert)
	}
	return 0
}

func (tree *BPlusTree) leafRemove(leaf *bplusLeaf, key key_t) int {
	remove, ok := leaf.keySearch(key)
	if !ok {
		/* Not exist */
		return -1
	}

	if leaf.entries <= (tree.entries+1)/2 {
		parent := leaf.parent
		l_sib := leaf.prev
		r_sib := leaf.next
		if parent != nil {
			/* decide which sibling to be borrowed from */
			i := leaf.parentKeyIdx
			if leafSiblingSelect(l_sib, r_sib, parent, i) == LEFT_SIBLING {
				if l_sib.entries > (tree.entries+1)/2 {
					leaf.shiftFromLeft(l_sib, i, remove)
				} else {
					leaf.mergeIntoLeft(l_sib, remove)
					/* trace upwards */
					tree.nonLeafRemove(parent, i)
				}
			} else {
				/* remove first in case of overflow during merging with sibling */
				leaf.simpleRemove(remove)
				if r_sib.entries > (tree.entries+1)/2 {
					leaf.shiftFromRight(r_sib, i+1)
				} else {
					leaf.mergeFromRight(r_sib)
					/* trace upwards */
					tree.nonLeafRemove(parent, i+1)
				}
			}
		} else {
			if leaf.entries == 1 {
				/* delete the only last node */
				assert(key == leaf.kvs[0].key)
				tree.root = nil
				leafDelete(leaf)
				return 0
			} else {
				leaf.simpleRemove(remove)
			}
		}
	} else {
		leaf.simpleRemove(remove)
	}

	return 0
}

func (tree *BPlusTree) Insert(key key_t, data value_t) int {
	node := tree.root
	for node != nil {
		if ln, ok := node.(*bplusLeaf); ok {
			return tree.leafInsert(ln, key, data)
		} else {
			nln := node.(*bplusNonLeaf)
			if i, found := nln.keySearch(key); found {
				node = nln.subPtr[i+1]
			} else {
				node = nln.subPtr[i]
			}
		}
	}

	/* new root */
	root := leafNew()
	root.kvs[0].key = key
	root.kvs[0].value = data
	root.entries = 1
	tree.root = root

	tree.firstLeaf = root
	return 0
}

func (tree *BPlusTree) Search(key key_t) (ret value_t, ok bool) {
	node := tree.root
	for node != nil {
		if ln, success := node.(*bplusLeaf); success {
			i, found := ln.keySearch(key)
			if found {
				ok = true
				ret = ln.kvs[i].value
			}
			break
		} else {
			nln := node.(*bplusNonLeaf)
			i, found := nln.keySearch(key)
			if found {
				node = nln.subPtr[i+1]
			} else {
				node = nln.subPtr[i]
			}
		}
	}
	return
}

func (tree *BPlusTree) nonLeafRemove(node *bplusNonLeaf, remove int) {
	if node.children <= (tree.order+1)/2 {
		l_sib := node.prev
		r_sib := node.next
		parent := node.parent
		if parent != nil {
			/* decide which sibling to be borrowed from */
			i := node.parentKeyIdx
			if nonLeafSiblingSelect(l_sib, r_sib, parent, i) == LEFT_SIBLING {
				if l_sib.children > (tree.order+1)/2 {
					node.shiftFromLeft(l_sib, i, remove)
				} else {
					node.mergeIntoLeft(l_sib, i, remove)
					/* trace upwards */
					tree.nonLeafRemove(parent, i)
				}
			} else {
				/* remove first in case of overflow during merging with sibling */
				node.simpleRemove(remove)
				if r_sib.children > (tree.order+1)/2 {
					node.shiftFromRight(r_sib, i+1)
				} else {
					node.mergeFromRight(r_sib, i+1)
					/* trace upwards */
					tree.nonLeafRemove(parent, i+1)
				}
			}
		} else {
			if node.children == 2 {
				/* delete old root node */
				assert(remove == 0)
				sbn := getNode(node.subPtr[0])
				sbn.parent = nil
				tree.root = node.subPtr[0]
				node.delete()
				tree.level--
			} else {
				node.simpleRemove(remove)
			}
		}
	} else {
		node.simpleRemove(remove)
	}
}

func (tree *BPlusTree) Delete(key key_t) int {
	node := tree.root
	for node != nil {
		if ln, ok := node.(*bplusLeaf); ok {
			return tree.leafRemove(ln, key)
		} else {
			nln := node.(*bplusNonLeaf)
			i, found := nln.keySearch(key)
			if found {
				node = nln.subPtr[i+1]
			} else {
				node = nln.subPtr[i]
			}
		}
	}
	return -1
}

func (tree *BPlusTree) listIsLastLeaf(link *bplusLeaf) bool {
	return link == tree.firstLeaf
}

func (tree *BPlusTree) GetRange(key1 key_t, key2 key_t) (value_t, bool) {
	var data value_t
	var min, max key_t
	if key1 <= key2 {
		min = key1
		max = key2
	} else {
		min = key2
		max = key1
	}
	node := tree.root
	for node != nil {
		if ln, ok := node.(*bplusLeaf); ok {
			i, found := ln.keySearch(min)
			if !found {
				if i >= ln.entries {
					if tree.listIsLastLeaf(ln) {
						return data, false
					}
					ln = ln.next
				}
			}
			for ln.kvs[i].key <= max {
				data = ln.kvs[i].value
				i++
				if i >= ln.entries {
					if tree.listIsLastLeaf(ln) {
						return data, false
					}
					ln = ln.prev
					i = 0
				}
			}
			break
		} else {
			nln := node.(*bplusNonLeaf)
			i, found := nln.keySearch(min)
			if found {
				node = nln.subPtr[i+1]
			} else {
				node = nln.subPtr[i]
			}
		}
	}

	return data, true
}

func Dump(tree *BPlusTree) {
	type nodeBacklog struct {
		/* Node backlogged */
		node node
		/* The index next to the backtrack point, must be >= 1 */
		next_sub_idx int
	}

	var level = 0
	var nbl *nodeBacklog
	var nbl_stack [BPLUS_MAX_LEVEL]nodeBacklog
	var topIndex int
	top := &nbl_stack[topIndex]

	node := tree.root
	for {
		if node != nil {
			/* non-zero needs backward and zero does not */
			var sub_idx int
			if nbl != nil {
				sub_idx = nbl.next_sub_idx
			}
			/* Reset each loop */
			nbl = nil

			/* Backlog the path */
			nl, ok := node.(*bplusNonLeaf)
			if !ok || sub_idx+1 >= nl.children { // leaf or no children
				top.node = nil
				top.next_sub_idx = 0
			} else {
				top.node = node
				top.next_sub_idx = sub_idx + 1
			}
			topIndex++
			top = &nbl_stack[topIndex]

			level++

			/* Draw the whole node when the first entry is passed through */
			if sub_idx == 0 {
				for i := 1; i < level; i++ {
					if i == level-1 {
						fmt.Printf("%-8s", "+-------")
					} else {
						if nbl_stack[i-1].node != nil {
							fmt.Printf("%-8s", "|")
						} else {
							fmt.Printf("%-8s", " ")
						}
					}
				}
				key_print(node)
			}

			/* Move deep down */
			if nln, ok := node.(*bplusNonLeaf); ok {
				node = nln.subPtr[sub_idx]
			} else {
				node = nil
			}
		} else {
			if topIndex == 0 {
				nbl = nil
			} else {
				topIndex--
				top = &nbl_stack[topIndex]
				nbl = top
			}
			if nbl == nil {
				/* End of traversal */
				break
			}
			node = nbl.node
			level--
		}
	}
}

func key_print(node node) {
	if leaf, ok := node.(*bplusLeaf); ok {
		fmt.Printf("leaf:")
		for i := 0; i < leaf.entries; i++ {
			fmt.Printf(" %d", leaf.kvs[i].key)
		}
	} else {
		non_leaf := node.(*bplusNonLeaf)
		fmt.Printf("node:")
		for i := 0; i < non_leaf.children-1; i++ {
			fmt.Printf(" %d", non_leaf.key[i])
		}
	}
	println()
}
