package geonames

//
//	go test ./pkg/storage/geonames/ -bench=BenchmarkNameIndex -benchmem
//
// Sample results (linux/amd64, Intel i5-11400 @ 2.60GHz, go test -count=1):
//
//	BenchmarkNameIndexBuild/Radix-12                 123 ms/op    36 MB/op   1.02M allocs/op
//	BenchmarkNameIndexBuild/RWay-12                  211 ms/op    98 MB/op   1.47M allocs/op
//
//	BenchmarkNameIndexPrefixSearch/Radix/a-12        895 µs/op   358 KB/op      19 allocs/op
//	BenchmarkNameIndexPrefixSearch/RWay/a-12          10 ms/op   247 KB/op       7 allocs/op
//
//	BenchmarkNameIndexPrefixSearch/Radix/min-12      666 µs/op   358 KB/op      19 allocs/op
//	BenchmarkNameIndexPrefixSearch/RWay/min-12        12 ms/op   247 KB/op       5 allocs/op
//
//	BenchmarkNameIndexPrefixSearch/Radix/mosc-12      29 ns/op     0 B/op       0 allocs/op
//	BenchmarkNameIndexPrefixSearch/RWay/mosc-12       54 ns/op     0 B/op       0 allocs/op
//
//	BenchmarkNameIndexPrefixSearch/Radix/zzzz-12     8.9 ns/op     0 B/op       0 allocs/op
//	BenchmarkNameIndexPrefixSearch/RWay/zzzz-12       27 ns/op     0 B/op       0 allocs/op
//
//	BenchmarkNameIndexRetainedHeap/Radix-12           ~34 MB retained
//	BenchmarkNameIndexRetainedHeap/RWay-12            ~97 MB retained
//

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/derekparker/trie"
	"github.com/gammazero/radixtree"
)

// Cities500 is ~2e5 rows; keep the default close to that.
const benchNameCount = 200_000

var benchPrefixes = []string{
	"a",
	"min",
	"mosc",
	"zzzz",
}

func benchCityNames(n int) []string {
	names := make([]string, n)
	prefixes := []string{
		"New", "San", "Saint", "Port", "Fort", "Lake", "Mount", "North", "South", "East", "West",
		"Min", "Mos", "Par", "Ber", "Lon", "Mad", "Rom", "Ath", "Lis",
	}
	suffixes := []string{
		"ton", "ville", "burg", "field", "ford", "ham", "stead", "wood", "land", "view",
		"sk", "ow", "is", "ia", "um", "en", "er", "an", "in", "on",
	}
	for i := range names {
		p := prefixes[i%len(prefixes)]
		s := suffixes[(i/len(prefixes))%len(suffixes)]
		names[i] = fmt.Sprintf("%s%s%d", p, s, i%997)
	}
	return names
}

type radixNameIndex struct {
	tree *radixtree.Tree[[]int]
}

func buildRadix(names []string) *radixNameIndex {
	idx := &radixNameIndex{tree: radixtree.New[[]int]()}
	for i, name := range names {
		key := strings.ToLower(name)
		indexes, _ := idx.tree.Get(key)
		indexes = append(indexes, i)
		idx.tree.Put(key, indexes)
	}
	return idx
}

func (idx *radixNameIndex) indexesByPrefix(prefix string) []int {
	prefix = strings.ToLower(prefix)
	var res []int
	for _, indexes := range idx.tree.IterAt(prefix) {
		res = append(res, indexes...)
	}
	return res
}

type rwayNameIndex struct {
	tree *trie.Trie
}

func buildRWay(names []string) *rwayNameIndex {
	idx := &rwayNameIndex{tree: trie.New()}
	for i, name := range names {
		idx.tree.Add(strings.ToLower(name), i)
	}
	return idx
}

func (idx *rwayNameIndex) indexesByPrefix(prefix string) []int {
	prefix = strings.ToLower(prefix)
	keys := idx.tree.PrefixSearch(prefix)
	sort.Strings(keys)
	res := make([]int, 0, len(keys))
	for _, key := range keys {
		node, _ := idx.tree.Find(key)
		res = append(res, node.Meta().(int))
	}
	return res
}

func BenchmarkNameIndexBuild(b *testing.B) {
	names := benchCityNames(benchNameCount)
	b.ResetTimer()

	b.Run("Radix", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			idx := buildRadix(names)
			runtime.KeepAlive(idx)
		}
	})

	b.Run("RWay", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			idx := buildRWay(names)
			runtime.KeepAlive(idx)
		}
	})
}

func BenchmarkNameIndexPrefixSearch(b *testing.B) {
	names := benchCityNames(benchNameCount)
	radixIdx := buildRadix(names)
	rwayIdx := buildRWay(names)

	for _, prefix := range benchPrefixes {
		b.Run("Radix/"+prefix, func(b *testing.B) {
			b.ReportAllocs()
			var n int
			for i := 0; i < b.N; i++ {
				n += len(radixIdx.indexesByPrefix(prefix))
			}
			runtime.KeepAlive(n)
		})
		b.Run("RWay/"+prefix, func(b *testing.B) {
			b.ReportAllocs()
			var n int
			for i := 0; i < b.N; i++ {
				n += len(rwayIdx.indexesByPrefix(prefix))
			}
			runtime.KeepAlive(n)
		})
	}
}

func BenchmarkNameIndexRetainedHeap(b *testing.B) {
	names := benchCityNames(benchNameCount)

	measure := func(b *testing.B, build func([]string) any) {
		b.ReportAllocs()
		var lastRetained uint64
		for i := 0; i < b.N; i++ {
			runtime.GC()
			var before runtime.MemStats
			runtime.ReadMemStats(&before)

			idx := build(names)

			runtime.GC()
			var after runtime.MemStats
			runtime.ReadMemStats(&after)
			runtime.KeepAlive(idx)

			if after.HeapAlloc > before.HeapAlloc {
				lastRetained = after.HeapAlloc - before.HeapAlloc
			} else {
				lastRetained = 0
			}
		}
		b.ReportMetric(float64(lastRetained), "retained-B")
	}

	b.Run("Radix", func(b *testing.B) {
		measure(b, func(names []string) any { return buildRadix(names) })
	})
	b.Run("RWay", func(b *testing.B) {
		measure(b, func(names []string) any { return buildRWay(names) })
	})
}
