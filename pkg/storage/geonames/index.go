package geonames

import (
	"strings"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/gammazero/radixtree"
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

	geoNameIDToCollectionIndex map[uint32]int
	names                      *radixtree.Tree[[]int]
	countryCodeToRange         map[string]*indexRange
}

func (idx *index[T]) Init(collection []T) {
	idx.collection = collection

	idx.names = radixtree.New[[]int]()

	idx.countryCodeToRange = make(map[string]*indexRange)
	idx.geoNameIDToCollectionIndex = make(map[uint32]int)

	for i, item := range collection {
		// search by name prefix
		key := strings.ToLower(item.GetName())
		indexes, _ := idx.names.Get(key)
		indexes = append(indexes, i)
		idx.names.Put(key, indexes)

		// search by geoNameID
		idx.geoNameIDToCollectionIndex[uint32(item.GetGeoNameID())] = i

		// filter by country code
		// expected collection sorted by country code
		currentRange := idx.countryCodeToRange[item.GetCountryCode()]
		if currentRange == nil {
			currentRange = &indexRange{begin: i, end: i + 1}
			idx.countryCodeToRange[item.GetCountryCode()] = currentRange
		} else {
			currentRange.end = i + 1
		}
	}
}

func (idx *index[T]) GetFiltered(filter entity.GeoNameFilter) (res []T) {
	switch {
	case len(filter.GeoNameIDs) > 0:
		collection := make([]T, 0, len(filter.GeoNameIDs))
		for _, geoNameID := range filter.GeoNameIDs {
			if i, ok := idx.geoNameIDToCollectionIndex[geoNameID]; ok {
				collection = append(collection, idx.collection[i])
			}
		}
		return collection
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
	namePrefix = strings.ToLower(namePrefix)
	var res []int
	for _, indexes := range idx.names.IterAt(namePrefix) {
		res = append(res, indexes...)
	}
	return res
}

func (idx *index[T]) rangesByCountryCodes(codes ...string) []*indexRange {
	var ranges []*indexRange
	for _, code := range codes {
		code = strings.ToUpper(code)
		idxRange := idx.countryCodeToRange[code]
		if idxRange == nil {
			continue
		}
		ranges = append(ranges, idxRange)
	}
	return ranges
}
