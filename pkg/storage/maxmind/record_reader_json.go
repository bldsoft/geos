package maxmind

import (
	"encoding/json"
	"io"
	"net"

	"github.com/bldsoft/geos/pkg/entity"
)

type JSONRecordReader struct {
	records    []MMDBRecord
	currentIdx int
}

func NewJSONRecordReader(r io.Reader) (recordReader *JSONRecordReader, err error) {
	m := make(map[string]entity.City)
	// m := make(map[string]interface{})
	dec := json.NewDecoder(r)
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	res := JSONRecordReader{}

	records := make([]MMDBRecord, 0, len(m))
	for key, value := range m {
		var record MMDBRecord
		_, record.Network, err = net.ParseCIDR(key)
		if err != nil {
			return nil, err
		}
		record.Data = value.ToMMDBType()
		// record.Data = toMMDBType(value).(mmdbtype.Map)
		records = append(records, record)
	}

	res.records = records

	return &res, nil
}

func (r *JSONRecordReader) ReadMMDBRecord() (record MMDBRecord, err error) {
	if r.currentIdx >= len(r.records) {
		return record, io.EOF
	}
	rec := r.records[r.currentIdx]
	r.currentIdx++
	return rec, nil
}
