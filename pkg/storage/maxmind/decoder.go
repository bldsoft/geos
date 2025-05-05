package maxmind

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sync"
)

type Decoder struct {
	Buffer []byte
}

type dataType int

const (
	_Extended dataType = iota
	_Pointer
	_String
	_Float64
	_Bytes
	_Uint16
	_Uint32
	_Map
	_Int32
	_Uint64
	_Uint128
	_Slice
	// We don't use the next two. They are placeholders. See the spec
	// for more details.
	_Container //nolint: deadcode, varcheck // above
	_Marker    //nolint: deadcode, varcheck // above
	_Bool
	_Float32
)

const (
	// This is the value used in libmaxminddb.
	maximumDataStructureDepth = 512
)

type deserializer interface {
	ShouldSkip(offset uintptr) (bool, error)
	StartSlice(size uint) error
	StartMap(size uint) error
	End() error
	String(string) error
	Float64(float64) error
	Bytes([]byte) error
	Uint16(uint16) error
	Uint32(uint32) error
	Int32(int32) error
	Uint64(uint64) error
	Uint128(*big.Int) error
	Bool(bool) error
	Float32(float32) error
}

func (d *Decoder) Decode(offset uint, result reflect.Value, depth int) (uint, error) {
	if depth > maximumDataStructureDepth {
		return 0, fmt.Errorf(
			"exceeded maximum data structure depth; database is likely corrupt",
		)
	}
	typeNum, size, newOffset, err := d.decodeCtrlData(offset)
	if err != nil {
		return 0, err
	}

	if typeNum != _Pointer && result.Kind() == reflect.Uintptr {
		result.Set(reflect.ValueOf(uintptr(offset)))
		return d.nextValueOffset(offset, 1)
	}
	return d.decodeFromType(typeNum, size, newOffset, result, depth+1)
}

func (d *Decoder) decodeToDeserializer(
	offset uint,
	dser deserializer,
	depth int,
	getNext bool,
) (uint, error) {
	if depth > maximumDataStructureDepth {
		return 0, fmt.Errorf(
			"exceeded maximum data structure depth; database is likely corrupt",
		)
	}
	skip, err := dser.ShouldSkip(uintptr(offset))
	if err != nil {
		return 0, err
	}
	if skip {
		if getNext {
			return d.nextValueOffset(offset, 1)
		}
		return 0, nil
	}

	typeNum, size, newOffset, err := d.decodeCtrlData(offset)
	if err != nil {
		return 0, err
	}

	return d.decodeFromTypeToDeserializer(typeNum, size, newOffset, dser, depth+1)
}

func (d *Decoder) decodeCtrlData(offset uint) (dataType, uint, uint, error) {
	newOffset := offset + 1
	if offset >= uint(len(d.Buffer)) {
		return 0, 0, 0, fmt.Errorf("invalid range")
	}
	ctrlByte := d.Buffer[offset]

	typeNum := dataType(ctrlByte >> 5)
	if typeNum == _Extended {
		if newOffset >= uint(len(d.Buffer)) {
			return 0, 0, 0, fmt.Errorf("invalid range")
		}
		typeNum = dataType(d.Buffer[newOffset] + 7)
		newOffset++
	}

	var size uint
	size, newOffset, err := d.sizeFromCtrlByte(ctrlByte, newOffset, typeNum)
	return typeNum, size, newOffset, err
}

func (d *Decoder) sizeFromCtrlByte(
	ctrlByte byte,
	offset uint,
	typeNum dataType,
) (uint, uint, error) {
	size := uint(ctrlByte & 0x1f)
	if typeNum == _Extended {
		return size, offset, nil
	}

	var bytesToRead uint
	if size < 29 {
		return size, offset, nil
	}

	bytesToRead = size - 28
	newOffset := offset + bytesToRead
	if newOffset > uint(len(d.Buffer)) {
		return 0, 0, fmt.Errorf("invalid offset")
	}
	if size == 29 {
		return 29 + uint(d.Buffer[offset]), offset + 1, nil
	}

	sizeBytes := d.Buffer[offset:newOffset]

	switch {
	case size == 30:
		size = 285 + uintFromBytes(0, sizeBytes)
	case size > 30:
		size = uintFromBytes(0, sizeBytes) + 65821
	}
	return size, newOffset, nil
}

func (d *Decoder) decodeFromType(
	dtype dataType,
	size uint,
	offset uint,
	result reflect.Value,
	depth int,
) (uint, error) {
	result = indirect(result)

	// For these types, size has a special meaning
	switch dtype {
	case _Bool:
		return unmarshalBool(size, offset, result)
	case _Map:
		return d.unmarshalMap(size, offset, result, depth)
	case _Pointer:
		return d.unmarshalPointer(size, offset, result, depth)
	case _Slice:
		return d.unmarshalSlice(size, offset, result, depth)
	}

	// For the remaining types, size is the byte size
	if offset+size > uint(len(d.Buffer)) {
		return 0, fmt.Errorf("invalid offset")
	}
	switch dtype {
	case _Bytes:
		return d.unmarshalBytes(size, offset, result)
	case _Float32:
		return d.unmarshalFloat32(size, offset, result)
	case _Float64:
		return d.unmarshalFloat64(size, offset, result)
	case _Int32:
		return d.unmarshalInt32(size, offset, result)
	case _String:
		return d.unmarshalString(size, offset, result)
	case _Uint16:
		return d.unmarshalUint(size, offset, result, 16)
	case _Uint32:
		return d.unmarshalUint(size, offset, result, 32)
	case _Uint64:
		return d.unmarshalUint(size, offset, result, 64)
	case _Uint128:
		return d.unmarshalUint128(size, offset, result)
	default:
		return 0, fmt.Errorf("unknown type: %d", dtype)
	}
}

func (d *Decoder) decodeFromTypeToDeserializer(
	dtype dataType,
	size uint,
	offset uint,
	dser deserializer,
	depth int,
) (uint, error) {
	// For these types, size has a special meaning
	switch dtype {
	case _Bool:
		v, offset := decodeBool(size, offset)
		return offset, dser.Bool(v)
	case _Map:
		return d.decodeMapToDeserializer(size, offset, dser, depth)
	case _Pointer:
		pointer, newOffset, err := d.decodePointer(size, offset)
		if err != nil {
			return 0, err
		}
		_, err = d.decodeToDeserializer(pointer, dser, depth, false)
		return newOffset, err
	case _Slice:
		return d.decodeSliceToDeserializer(size, offset, dser, depth)
	}

	// For the remaining types, size is the byte size
	if offset+size > uint(len(d.Buffer)) {
		return 0, fmt.Errorf("invalid offset")
	}
	switch dtype {
	case _Bytes:
		v, offset := d.decodeBytes(size, offset)
		return offset, dser.Bytes(v)
	case _Float32:
		v, offset := d.decodeFloat32(size, offset)
		return offset, dser.Float32(v)
	case _Float64:
		v, offset := d.decodeFloat64(size, offset)
		return offset, dser.Float64(v)
	case _Int32:
		v, offset := d.decodeInt(size, offset)
		return offset, dser.Int32(int32(v))
	case _String:
		v, offset := d.decodeString(size, offset)
		return offset, dser.String(v)
	case _Uint16:
		v, offset := d.decodeUint(size, offset)
		return offset, dser.Uint16(uint16(v))
	case _Uint32:
		v, offset := d.decodeUint(size, offset)
		return offset, dser.Uint32(uint32(v))
	case _Uint64:
		v, offset := d.decodeUint(size, offset)
		return offset, dser.Uint64(v)
	case _Uint128:
		v, offset := d.decodeUint128(size, offset)
		return offset, dser.Uint128(v)
	default:
		return 0, fmt.Errorf("unknown type: %d", dtype)
	}
}

func unmarshalBool(size, offset uint, result reflect.Value) (uint, error) {
	if size > 1 {
		return 0, fmt.Errorf(
			"the MaxMind DB file's data section contains bad data (bool size of %v)",
			size,
		)
	}
	value, newOffset := decodeBool(size, offset)

	switch result.Kind() {
	case reflect.Bool:
		result.SetBool(value)
		return newOffset, nil
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
}

// indirect follows pointers and create values as necessary. This is
// heavily based on encoding/json as my original version had a subtle
// bug. This method should be considered to be licensed under
// https://golang.org/LICENSE
func indirect(result reflect.Value) reflect.Value {
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if result.Kind() == reflect.Interface && !result.IsNil() {
			e := result.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				result = e
				continue
			}
		}

		if result.Kind() != reflect.Ptr {
			break
		}

		if result.IsNil() {
			result.Set(reflect.New(result.Type().Elem()))
		}

		result = result.Elem()
	}
	return result
}

var sliceType = reflect.TypeOf([]byte{})

func (d *Decoder) unmarshalBytes(size, offset uint, result reflect.Value) (uint, error) {
	value, newOffset := d.decodeBytes(size, offset)

	switch result.Kind() {
	case reflect.Slice:
		if result.Type() == sliceType {
			result.SetBytes(value)
			return newOffset, nil
		}
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf(
		"cannot unmarshal %v into %v", value, result.Type())
}

func (d *Decoder) unmarshalFloat32(size, offset uint, result reflect.Value) (uint, error) {
	if size != 4 {
		return 0, fmt.Errorf(
			"the MaxMind DB file's data section contains bad data (float32 size of %v)",
			size,
		)
	}
	value, newOffset := d.decodeFloat32(size, offset)

	switch result.Kind() {
	case reflect.Float32, reflect.Float64:
		result.SetFloat(float64(value))
		return newOffset, nil
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
}

func (d *Decoder) unmarshalFloat64(size, offset uint, result reflect.Value) (uint, error) {
	if size != 8 {
		return 0, fmt.Errorf(
			"the MaxMind DB file's data section contains bad data (float 64 size of %v)",
			size,
		)
	}
	value, newOffset := d.decodeFloat64(size, offset)

	switch result.Kind() {
	case reflect.Float32, reflect.Float64:
		if result.OverflowFloat(value) {
			return 0, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
		}
		result.SetFloat(value)
		return newOffset, nil
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
}

func (d *Decoder) unmarshalInt32(size, offset uint, result reflect.Value) (uint, error) {
	if size > 4 {
		return 0, fmt.Errorf(
			"the MaxMind DB file's data section contains bad data (int32 size of %v)",
			size,
		)
	}
	value, newOffset := d.decodeInt(size, offset)

	switch result.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := int64(value)
		if !result.OverflowInt(n) {
			result.SetInt(n)
			return newOffset, nil
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		n := uint64(value)
		if !result.OverflowUint(n) {
			result.SetUint(n)
			return newOffset, nil
		}
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
}

func (d *Decoder) unmarshalMap(
	size uint,
	offset uint,
	result reflect.Value,
	depth int,
) (uint, error) {
	result = indirect(result)
	switch result.Kind() {
	default:
		return 0, fmt.Errorf("map")
	case reflect.Struct:
		return d.decodeStruct(size, offset, result, depth)
	case reflect.Map:
		return d.decodeMap(size, offset, result, depth)
	case reflect.Interface:
		if result.NumMethod() == 0 {
			rv := reflect.ValueOf(make(map[string]any, size))
			newOffset, err := d.decodeMap(size, offset, rv, depth)
			result.Set(rv)
			return newOffset, err
		}
		return 0, fmt.Errorf("map")
	}
}

func (d *Decoder) unmarshalPointer(
	size, offset uint,
	result reflect.Value,
	depth int,
) (uint, error) {
	pointer, newOffset, err := d.decodePointer(size, offset)
	if err != nil {
		return 0, err
	}
	_, err = d.Decode(pointer, result, depth)
	return newOffset, err
}

func (d *Decoder) unmarshalSlice(
	size uint,
	offset uint,
	result reflect.Value,
	depth int,
) (uint, error) {
	switch result.Kind() {
	case reflect.Slice:
		return d.decodeSlice(size, offset, result, depth)
	case reflect.Interface:
		if result.NumMethod() == 0 {
			a := []any{}
			rv := reflect.ValueOf(&a).Elem()
			newOffset, err := d.decodeSlice(size, offset, rv, depth)
			result.Set(rv)
			return newOffset, err
		}
	}
	return 0, fmt.Errorf("array")
}

func (d *Decoder) unmarshalString(size, offset uint, result reflect.Value) (uint, error) {
	value, newOffset := d.decodeString(size, offset)

	switch result.Kind() {
	case reflect.String:
		result.SetString(value)
		return newOffset, nil
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
}

func (d *Decoder) unmarshalUint(
	size, offset uint,
	result reflect.Value,
	uintType uint,
) (uint, error) {
	if size > uintType/8 {
		return 0, fmt.Errorf(
			"the MaxMind DB file's data section contains bad data (uint%v size of %v)",
			uintType,
			size,
		)
	}

	value, newOffset := d.decodeUint(size, offset)

	switch result.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := int64(value)
		if !result.OverflowInt(n) {
			result.SetInt(n)
			return newOffset, nil
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		if !result.OverflowUint(value) {
			result.SetUint(value)
			return newOffset, nil
		}
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
}

var bigIntType = reflect.TypeOf(big.Int{})

func (d *Decoder) unmarshalUint128(size, offset uint, result reflect.Value) (uint, error) {
	if size > 16 {
		return 0, fmt.Errorf(
			"the MaxMind DB file's data section contains bad data (uint128 size of %v)",
			size,
		)
	}
	value, newOffset := d.decodeUint128(size, offset)

	switch result.Kind() {
	case reflect.Struct:
		if result.Type() == bigIntType {
			result.Set(reflect.ValueOf(*value))
			return newOffset, nil
		}
	case reflect.Interface:
		if result.NumMethod() == 0 {
			result.Set(reflect.ValueOf(value))
			return newOffset, nil
		}
	}
	return newOffset, fmt.Errorf("cannot unmarshal %v into %v", value, result.Type())
}

func decodeBool(size, offset uint) (bool, uint) {
	return size != 0, offset
}

func (d *Decoder) decodeBytes(size, offset uint) ([]byte, uint) {
	newOffset := offset + size
	bytes := make([]byte, size)
	copy(bytes, d.Buffer[offset:newOffset])
	return bytes, newOffset
}

func (d *Decoder) decodeFloat64(size, offset uint) (float64, uint) {
	newOffset := offset + size
	bits := binary.BigEndian.Uint64(d.Buffer[offset:newOffset])
	return math.Float64frombits(bits), newOffset
}

func (d *Decoder) decodeFloat32(size, offset uint) (float32, uint) {
	newOffset := offset + size
	bits := binary.BigEndian.Uint32(d.Buffer[offset:newOffset])
	return math.Float32frombits(bits), newOffset
}

func (d *Decoder) decodeInt(size, offset uint) (int, uint) {
	newOffset := offset + size
	var val int32
	for _, b := range d.Buffer[offset:newOffset] {
		val = (val << 8) | int32(b)
	}
	return int(val), newOffset
}

func (d *Decoder) decodeMap(
	size uint,
	offset uint,
	result reflect.Value,
	depth int,
) (uint, error) {
	if result.IsNil() {
		result.Set(reflect.MakeMapWithSize(result.Type(), int(size)))
	}

	mapType := result.Type()
	keyValue := reflect.New(mapType.Key()).Elem()
	elemType := mapType.Elem()
	var elemValue reflect.Value
	for i := uint(0); i < size; i++ {
		var key []byte
		var err error
		key, offset, err = d.decodeKey(offset)
		if err != nil {
			return 0, err
		}

		if elemValue.IsValid() {
			// After 1.20 is the minimum supported version, this can just be
			// elemValue.SetZero()
			reflectSetZero(elemValue)
		} else {
			elemValue = reflect.New(elemType).Elem()
		}

		offset, err = d.Decode(offset, elemValue, depth)
		if err != nil {
			return 0, fmt.Errorf("decoding value for %s: %w", key, err)
		}

		keyValue.SetString(string(key))
		result.SetMapIndex(keyValue, elemValue)
	}
	return offset, nil
}

func (d *Decoder) decodeMapToDeserializer(
	size uint,
	offset uint,
	dser deserializer,
	depth int,
) (uint, error) {
	err := dser.StartMap(size)
	if err != nil {
		return 0, err
	}
	for i := uint(0); i < size; i++ {
		// TODO - implement key/value skipping?
		offset, err = d.decodeToDeserializer(offset, dser, depth, true)
		if err != nil {
			return 0, err
		}

		offset, err = d.decodeToDeserializer(offset, dser, depth, true)
		if err != nil {
			return 0, err
		}
	}
	err = dser.End()
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (d *Decoder) decodePointer(
	size uint,
	offset uint,
) (uint, uint, error) {
	pointerSize := ((size >> 3) & 0x3) + 1
	newOffset := offset + pointerSize
	if newOffset > uint(len(d.Buffer)) {
		return 0, 0, fmt.Errorf("invalid offset")
	}
	pointerBytes := d.Buffer[offset:newOffset]
	var prefix uint
	if pointerSize == 4 {
		prefix = 0
	} else {
		prefix = size & 0x7
	}
	unpacked := uintFromBytes(prefix, pointerBytes)

	var pointerValueOffset uint
	switch pointerSize {
	case 1:
		pointerValueOffset = 0
	case 2:
		pointerValueOffset = 2048
	case 3:
		pointerValueOffset = 526336
	case 4:
		pointerValueOffset = 0
	}

	pointer := unpacked + pointerValueOffset

	return pointer, newOffset, nil
}

func (d *Decoder) decodeSlice(
	size uint,
	offset uint,
	result reflect.Value,
	depth int,
) (uint, error) {
	result.Set(reflect.MakeSlice(result.Type(), int(size), int(size)))
	for i := 0; i < int(size); i++ {
		var err error
		offset, err = d.Decode(offset, result.Index(i), depth)
		if err != nil {
			return 0, err
		}
	}
	return offset, nil
}

func (d *Decoder) decodeSliceToDeserializer(
	size uint,
	offset uint,
	dser deserializer,
	depth int,
) (uint, error) {
	err := dser.StartSlice(size)
	if err != nil {
		return 0, err
	}
	for i := uint(0); i < size; i++ {
		offset, err = d.decodeToDeserializer(offset, dser, depth, true)
		if err != nil {
			return 0, err
		}
	}
	err = dser.End()
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (d *Decoder) decodeString(size, offset uint) (string, uint) {
	newOffset := offset + size
	return string(d.Buffer[offset:newOffset]), newOffset
}

func (d *Decoder) decodeStruct(
	size uint,
	offset uint,
	result reflect.Value,
	depth int,
) (uint, error) {
	fields := cachedFields(result)

	// This fills in embedded structs
	for _, i := range fields.anonymousFields {
		_, err := d.unmarshalMap(size, offset, result.Field(i), depth)
		if err != nil {
			return 0, err
		}
	}

	// This handles named fields
	for i := uint(0); i < size; i++ {
		var (
			err error
			key []byte
		)
		key, offset, err = d.decodeKey(offset)
		if err != nil {
			return 0, err
		}
		// The string() does not create a copy due to this compiler
		// optimization: https://github.com/golang/go/issues/3512
		j, ok := fields.namedFields[string(key)]
		if !ok {
			offset, err = d.nextValueOffset(offset, 1)
			if err != nil {
				return 0, err
			}
			continue
		}

		offset, err = d.Decode(offset, result.Field(j), depth)
		if err != nil {
			return 0, fmt.Errorf("decoding value for %s: %w", key, err)
		}
	}
	return offset, nil
}

type fieldsType struct {
	namedFields     map[string]int
	anonymousFields []int
}

var fieldsMap sync.Map

func cachedFields(result reflect.Value) *fieldsType {
	resultType := result.Type()

	if fields, ok := fieldsMap.Load(resultType); ok {
		return fields.(*fieldsType)
	}
	numFields := resultType.NumField()
	namedFields := make(map[string]int, numFields)
	var anonymous []int
	for i := 0; i < numFields; i++ {
		field := resultType.Field(i)

		fieldName := field.Name
		if tag := field.Tag.Get("maxminddb"); tag != "" {
			if tag == "-" {
				continue
			}
			fieldName = tag
		}
		if field.Anonymous {
			anonymous = append(anonymous, i)
			continue
		}
		namedFields[fieldName] = i
	}
	fields := &fieldsType{namedFields, anonymous}
	fieldsMap.Store(resultType, fields)

	return fields
}

func (d *Decoder) decodeUint(size, offset uint) (uint64, uint) {
	newOffset := offset + size
	bytes := d.Buffer[offset:newOffset]

	var val uint64
	for _, b := range bytes {
		val = (val << 8) | uint64(b)
	}
	return val, newOffset
}

func (d *Decoder) decodeUint128(size, offset uint) (*big.Int, uint) {
	newOffset := offset + size
	val := new(big.Int)
	val.SetBytes(d.Buffer[offset:newOffset])

	return val, newOffset
}

func uintFromBytes(prefix uint, uintBytes []byte) uint {
	val := prefix
	for _, b := range uintBytes {
		val = (val << 8) | uint(b)
	}
	return val
}

// decodeKey decodes a map key into []byte slice. We use a []byte so that we
// can take advantage of https://github.com/golang/go/issues/3512 to avoid
// copying the bytes when decoding a struct. Previously, we achieved this by
// using unsafe.
func (d *Decoder) decodeKey(offset uint) ([]byte, uint, error) {
	typeNum, size, dataOffset, err := d.decodeCtrlData(offset)
	if err != nil {
		return nil, 0, err
	}
	if typeNum == _Pointer {
		pointer, ptrOffset, err := d.decodePointer(size, dataOffset)
		if err != nil {
			return nil, 0, err
		}
		key, _, err := d.decodeKey(pointer)
		return key, ptrOffset, err
	}
	if typeNum != _String {
		return nil, 0, fmt.Errorf("unexpected type when decoding string: %v", typeNum)
	}
	newOffset := dataOffset + size
	if newOffset > uint(len(d.Buffer)) {
		return nil, 0, fmt.Errorf("invalid offset")
	}
	return d.Buffer[dataOffset:newOffset], newOffset, nil
}

// This function is used to skip ahead to the next value without decoding
// the one at the offset passed in. The size bits have different meanings for
// different data types.
func (d *Decoder) nextValueOffset(offset, numberToSkip uint) (uint, error) {
	if numberToSkip == 0 {
		return offset, nil
	}
	typeNum, size, offset, err := d.decodeCtrlData(offset)
	if err != nil {
		return 0, err
	}
	switch typeNum {
	case _Pointer:
		_, offset, err = d.decodePointer(size, offset)
		if err != nil {
			return 0, err
		}
	case _Map:
		numberToSkip += 2 * size
	case _Slice:
		numberToSkip += size
	case _Bool:
	default:
		offset += size
	}
	return d.nextValueOffset(offset, numberToSkip-1)
}

func reflectSetZero(v reflect.Value) {
	v.SetZero()
}
