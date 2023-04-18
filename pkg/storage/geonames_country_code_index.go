package storage

type indexRange struct {
	begin int
	end   int
}

func (r *indexRange) contains(i int) bool {
	return r.begin <= i && i < r.end
}

type geoNameIndex struct {
	// country code to indexes
	currentIndex       int
	countryCodeToRange map[string]*indexRange
	lastCountryCode    string
}

func newGeoNameIndex() geoNameIndex {
	return geoNameIndex{currentIndex: -1, countryCodeToRange: make(map[string]*indexRange)}
}

func (idx *geoNameIndex) put(countryCode string) {
	idx.currentIndex++
	currentRange := idx.countryCodeToRange[idx.lastCountryCode]
	if currentRange == nil {
		currentRange = &indexRange{}
		idx.countryCodeToRange[idx.lastCountryCode] = currentRange
	}

	if countryCode == idx.lastCountryCode {
		currentRange.end = idx.currentIndex + 1
		return
	}

	idx.lastCountryCode = countryCode
	idx.countryCodeToRange[countryCode] = &indexRange{begin: idx.currentIndex, end: idx.currentIndex + 1}
}

func (idx *geoNameIndex) indexRange(countryCode string) *indexRange {
	if res := idx.countryCodeToRange[countryCode]; res != nil {
		return res
	}
	return &indexRange{0, 0}
}
