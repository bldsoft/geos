package storage

import "strings"

type nameIndex struct {
	names []string
}

func (idx *nameIndex) Index(name string) {
	idx.names = append(idx.names, name)
}

func (idx *nameIndex) GetColectionIndexes(namePrefix string) []int {
	var res []int
	for i, n := range idx.names {
		if strings.HasPrefix(n, namePrefix) {
			res = append(res, i)
		}
	}
	return res
}
