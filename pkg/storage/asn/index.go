package asn

import (
	"sort"
	"strconv"
	"strings"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/utils"
	"github.com/gammazero/radixtree"
)

type index struct {
	collection []*entity.ASN
	byNumber   map[uint32]*entity.ASN
	names      *radixtree.Tree[[]int]
}

func (idx *index) Init(collection []*entity.ASN) {
	sort.Slice(collection, func(i, j int) bool {
		return collection[i].ASN < collection[j].ASN
	})
	idx.collection = collection
	idx.byNumber = make(map[uint32]*entity.ASN, len(collection))
	idx.names = radixtree.New[[]int]()

	for i, item := range collection {
		idx.byNumber[item.ASN] = item
		idx.putName(strings.ToLower(item.ASO), i)
		num := strconv.FormatUint(uint64(item.ASN), 10)
		idx.putName(num, i)
		idx.putName("as"+num, i)
	}
}

func (idx *index) putName(key string, i int) {
	if key == "" {
		return
	}
	indexes, _ := idx.names.Get(key)
	indexes = append(indexes, i)
	idx.names.Put(key, indexes)
}

func (idx *index) GetFiltered(filter entity.ASNFilter) (res []*entity.ASN) {
	switch {
	case len(filter.Numbers) > 0:
		res = make([]*entity.ASN, 0, len(filter.Numbers))
		for _, n := range filter.Numbers {
			if item, ok := idx.byNumber[n]; ok {
				res = append(res, item)
			}
		}
		return res
	case filter.NamePrefix == "":
		return idx.collection
	default:
		seen := utils.NewSet[int]()
		for _, i := range idx.indexesByNamePrefix(filter.NamePrefix) {
			if seen.Has(i) {
				continue
			}
			seen.Put(i)
			res = append(res, idx.collection[i])
		}
		return res
	}
}

func (idx *index) indexesByNamePrefix(namePrefix string) []int {
	namePrefix = strings.ToLower(namePrefix)
	var res []int
	for _, indexes := range idx.names.IterAt(namePrefix) {
		res = append(res, indexes...)
	}
	return res
}
