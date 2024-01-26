package repository

import (
	"reflect"

	"github.com/maxmind/mmdbwriter/mmdbtype"
)

func toMMDBType(val interface{}) mmdbtype.DataType {
	switch v := val.(type) {
	case string:
		return mmdbtype.String(v)
	case bool:
		return mmdbtype.Bool(v)
	case int32:
		return mmdbtype.Int32(v)
	case uint16:
		return mmdbtype.Uint16(v)
	case uint32:
		return mmdbtype.Uint32(v)
	case uint64:
		return mmdbtype.Uint64(v)
	case float32:
		return mmdbtype.Float32(v)
	case float64:
		return mmdbtype.Float64(v)
	}

	vof := reflect.ValueOf(val)
	switch vof.Kind() {
	case reflect.Slice:
		slice := make([]mmdbtype.DataType, 0, vof.Len())
		for i := 0; i < vof.Len(); i++ {
			slice = append(slice, toMMDBType(vof.Index(i).Interface()))
		}
		return mmdbtype.Slice(slice)
	case reflect.Map:
		m := make(map[mmdbtype.String]mmdbtype.DataType)
		for _, key := range vof.MapKeys() {
			m[mmdbtype.String(key.String())] = toMMDBType(vof.MapIndex(key).Interface())
		}
		return mmdbtype.Map(m)
	}
	return nil
}
