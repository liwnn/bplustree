package bplustree

import (
	"testing"
)

func TestBplusTree(t *testing.T) {
	tree := New(3, 3)
	tree.Insert(3, 30)
	tree.Insert(4, 40)
	tree.Insert(6, 60)
	tree.Insert(7, 70)
	tree.Insert(5, 50)
	tree.Insert(8, 80)
	tree.Insert(2, 20)

	x, found := tree.Search(6)
	if !found || x != 60 {
		panic("Search")
	}

	tree.GetRange(0, 7)
	tree.Delete(5)
}

func BenchmarkInsert(b *testing.B) {
	testCount := 1000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bt := New(255, 511)
		for i := testCount; i > 0; i-- {
			bt.Insert(i, 1)
		}
	}
}

func BenchmarkSearch(b *testing.B) {
	testCount := 1000000
	bt := New(255, 511)
	for i := testCount; i > 0; i-- {
		bt.Insert(i, 1)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < testCount; j++ {
			bt.Search(j)
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	testCount := b.N
	bt := New(255, 511)
	for i := testCount; i > 0; i-- {
		bt.Insert(i, 1)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bt.Delete(i)
	}
}
