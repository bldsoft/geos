package storage

import (
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/derekparker/trie"
)

type indexRange struct {
	begin int
	end   int
}

func (r *indexRange) contains(i int) bool {
	return r.begin <= i && i < r.end
}

type index[T entity.GeoNameEntity] struct {
	collection []T

	trie               *trie.Trie
	countryCodeToRange map[string]*indexRange
}

func (idx *index[T]) Init(collection []T) {
	idx.collection = collection

	idx.trie = trie.New()

	idx.countryCodeToRange = make(map[string]*indexRange)

	for i, item := range collection {
		// search by name prefix
		idx.trie.Add(item.Name(), i)

		// filter by country code
		// expected collection sorted by country code
		currentRange := idx.countryCodeToRange[item.CountryCode()]
		if currentRange == nil {
			currentRange = &indexRange{begin: i, end: i + 1}
			idx.countryCodeToRange[item.CountryCode()] = currentRange
		} else {
			currentRange.end = i + 1
		}
	}
}

func (idx *index[T]) GetFiltered(filter entity.GeoNameFilter) (res []T) {
	switch {
	case len(filter.CountryCodes) == 0 && len(filter.NamePrefix) == 0:
		return idx.collection
	case len(filter.CountryCodes) > 0 && len(filter.NamePrefix) > 0:
		for _, index := range idx.indexesByNamePrefix(filter.NamePrefix) {
			for _, rng := range idx.rangesByCountryCodes(filter.CountryCodes...) {
				if rng.contains(index) {
					res = append(res, idx.collection[index])
				}
			}
		}
	case len(filter.CountryCodes) > 0:
		collectionRanges := idx.rangesByCountryCodes(filter.CountryCodes...)
		for _, rng := range collectionRanges {
			res = append(res, idx.collection[rng.begin:rng.end]...)
		}
	default:
		for _, index := range idx.indexesByNamePrefix(filter.NamePrefix) {
			res = append(res, idx.collection[index])
		}
	}
	return
}

func (idx *index[T]) indexesByNamePrefix(namePrefix string) []int {
	keys := idx.trie.PrefixSearch(namePrefix)
	res := make([]int, 0, len(keys))
	for _, key := range keys {
		node, _ := idx.trie.Find(key)
		res = append(res, node.Meta().(int))
	}
	return res
}

func (idx *index[T]) rangesByCountryCodes(codes ...string) []*indexRange {
	var ranges []*indexRange
	for _, code := range codes {
		idxRange := idx.countryCodeToRange[code]
		if idxRange == nil {
			continue
		}
		ranges = append(ranges, idxRange)
	}
	return ranges
}
