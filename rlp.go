package ethutil

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"math/big"
	"reflect"
)

///////////////////////////////////////
type EthEncoder interface {
	EncodeData(rlpData interface{}) []byte
}
type EthDecoder interface {
	Get(idx int) *RlpValue
}

//////////////////////////////////////

type RlpEncoder struct {
	rlpData []byte
}

func NewRlpEncoder() *RlpEncoder {
	encoder := &RlpEncoder{}

	return encoder
}
func (coder *RlpEncoder) EncodeData(rlpData interface{}) []byte {
	return Encode(rlpData)
}

// Data rlpValueibutes are returned by the rlp decoder. The data rlpValueibutes represents
// one item within the rlp data structure. It's responsible for all the casting
// It always returns something rlpValueid
type RlpValue struct {
	Value interface{}
	kind  reflect.Value
}

func (rlpValue *RlpValue) String() string {
	return fmt.Sprintf("%q", rlpValue.Value)
}

func Conv(rlpValueib interface{}) *RlpValue {
	return &RlpValue{Value: rlpValueib, kind: reflect.ValueOf(rlpValueib)}
}

func NewRlpValue(rlpValueib interface{}) *RlpValue {
	return &RlpValue{Value: rlpValueib}
}

func (rlpValue *RlpValue) Type() reflect.Kind {
	return reflect.TypeOf(rlpValue.Value).Kind()
}

func (rlpValue *RlpValue) IsNil() bool {
	return rlpValue.Value == nil
}

func (rlpValue *RlpValue) Length() int {
	//return rlpValue.kind.Len()
	if data, ok := rlpValue.Value.([]interface{}); ok {
		return len(data)
	}

	return 0
}

func (rlpValue *RlpValue) AsRaw() interface{} {
	return rlpValue.Value
}

func (rlpValue *RlpValue) AsUint() uint64 {
	if Value, ok := rlpValue.Value.(uint8); ok {
		return uint64(Value)
	} else if Value, ok := rlpValue.Value.(uint16); ok {
		return uint64(Value)
	} else if Value, ok := rlpValue.Value.(uint32); ok {
		return uint64(Value)
	} else if Value, ok := rlpValue.Value.(uint64); ok {
		return Value
	}

	return 0
}

func (rlpValue *RlpValue) AsByte() byte {
	if Value, ok := rlpValue.Value.(byte); ok {
		return Value
	}

	return 0x0
}

func (rlpValue *RlpValue) AsBigInt() *big.Int {
	if a, ok := rlpValue.Value.([]byte); ok {
		b := new(big.Int)
		b.SetBytes(a)
		return b
	}

	return big.NewInt(0)
}

func (rlpValue *RlpValue) AsString() string {
	if a, ok := rlpValue.Value.([]byte); ok {
		return string(a)
	} else if a, ok := rlpValue.Value.(string); ok {
		return a
	} else {
		//panic(fmt.Sprintf("not string %T: %v", rlpValue.Value, rlpValue.Value))
	}

	return ""
}

func (rlpValue *RlpValue) AsBytes() []byte {
	if a, ok := rlpValue.Value.([]byte); ok {
		return a
	}

	return make([]byte, 0)
}

func (rlpValue *RlpValue) AsSlice() []interface{} {
	if d, ok := rlpValue.Value.([]interface{}); ok {
		return d
	}

	return []interface{}{}
}

func (rlpValue *RlpValue) AsSliceFrom(from int) *RlpValue {
	slice := rlpValue.AsSlice()

	return NewRlpValue(slice[from:])
}

func (rlpValue *RlpValue) AsSliceTo(to int) *RlpValue {
	slice := rlpValue.AsSlice()

	return NewRlpValue(slice[:to])
}

func (rlpValue *RlpValue) AsSliceFromTo(from, to int) *RlpValue {
	slice := rlpValue.AsSlice()

	return NewRlpValue(slice[from:to])
}

// Threat the rlpValueibute as a slice
func (rlpValue *RlpValue) Get(idx int) *RlpValue {
	if d, ok := rlpValue.Value.([]interface{}); ok {
		// Guard for oob
		if len(d) <= idx {
			return NewRlpValue(nil)
		}

		if idx < 0 {
			panic("negative idx for Rlp Get")
		}

		return NewRlpValue(d[idx])
	}

	// If this wasn't a slice you probably shouldn't be using this function
	return NewRlpValue(nil)
}

func (rlpValue *RlpValue) Cmp(o *RlpValue) bool {
	return reflect.DeepEqual(rlpValue.Value, o.Value)
}

func (rlpValue *RlpValue) Encode() []byte {
	return Encode(rlpValue.Value)
}

func NewRlpValueFromBytes(rlpData []byte) *RlpValue {
	if len(rlpData) != 0 {
		data, _ := Decode(rlpData, 0)
		return NewRlpValue(data)
	}

	return NewRlpValue(nil)
}

/// Raw methods
func BinaryLength(n uint64) uint64 {
	if n == 0 {
		return 0
	}

	return 1 + BinaryLength(n/256)
}

func ToBinarySlice(n uint64, length uint64) []uint64 {
	if length == 0 {
		length = BinaryLength(n)
	}

	if n == 0 {
		return []uint64{}
	}

	slice := ToBinarySlice(n/256, 0)
	slice = append(slice, n%256)

	return slice
}

// RLP Encoding/Decoding methods

func ToBin(n uint64, length uint64) string {
	var buf bytes.Buffer
	for _, val := range ToBinarySlice(n, length) {
		buf.WriteByte(byte(val))
	}

	return buf.String()
}

func FromBin(data []byte) uint64 {
	if len(data) == 0 {
		return 0
	}

	return FromBin(data[:len(data)-1])*256 + uint64(data[len(data)-1])
}

const (
	RlpEmptyList = 0x80
	RlpEmptyStr  = 0x40
)

func Decode(data []byte, pos uint64) (interface{}, uint64) {
	if pos > uint64(len(data)-1) {
		log.Println(data)
		log.Panicf("index out of range %d for data %q, l = %d", pos, data, len(data))
	}

	char := int(data[pos])
	slice := make([]interface{}, 0)
	switch {
	case char < 24:
		return data[pos], pos + 1

	case char < 56:
		b := uint64(data[pos]) - 23
		return FromBin(data[pos+1 : pos+1+b]), pos + 1 + b

	case char < 64:
		b := uint64(data[pos]) - 55
		b2 := uint64(FromBin(data[pos+1 : pos+1+b]))
		return FromBin(data[pos+1+b : pos+1+b+b2]), pos + 1 + b + b2

	case char < 120:
		b := uint64(data[pos]) - 64
		return data[pos+1 : pos+1+b], pos + 1 + b

	case char < 128:
		b := uint64(data[pos]) - 119
		b2 := uint64(FromBin(data[pos+1 : pos+1+b]))
		return data[pos+1+b : pos+1+b+b2], pos + 1 + b + b2

	case char < 184:
		b := uint64(data[pos]) - 128
		pos++
		for i := uint64(0); i < b; i++ {
			var obj interface{}

			obj, pos = Decode(data, pos)
			slice = append(slice, obj)
		}
		return slice, pos

	case char < 192:
		b := uint64(data[pos]) - 183
		//b2 := int(FromBin(data[pos+1 : pos+1+b])) (ref implementation has an unused variable)
		pos = pos + 1 + b
		for i := uint64(0); i < b; i++ {
			var obj interface{}

			obj, pos = Decode(data, pos)
			slice = append(slice, obj)
		}
		return slice, pos

	default:
		panic(fmt.Sprintf("byte not supported: %q", char))
	}

	return slice, 0
}

func Encode(object interface{}) []byte {
	var buff bytes.Buffer

	if object != nil {
		switch t := object.(type) {
		case int:
			buff.Write(Encode(uint32(t)))
		case uint16, uint32, uint64, int64, int32, int16, int8:
			var num uint64
			if _num, ok := t.(uint64); ok {
				num = _num
			} else if _num, ok := t.(uint32); ok {
				num = uint64(_num)
			} else if _num, ok := t.(uint16); ok {
				num = uint64(_num)
			}

			if num >= 0 && num < 24 {
				buff.WriteString(string(num))
			} else if num <= uint64(math.Pow(2, 256)) {
				b := ToBin(num, 0)
				buff.WriteString(string(len(b)+23) + b)
			} else {
				b := ToBin(num, 0)
				b2 := ToBin(uint64(len(b)), 0)
				buff.WriteString(string(len(b2)+55) + b2 + b)
			}

		case *big.Int:
			buff.Write(Encode(string(t.Bytes())))

		case string:
			if len(t) < 56 {
				buff.WriteString(string(len(t)+64) + t)
			} else {
				b2 := ToBin(uint64(len(t)), 0)
				buff.WriteString(string(len(b2)+119) + b2 + t)
			}

		case byte:
			buff.Write(Encode(uint32(t)))
		case []byte:
			// Cast the byte slice to a string
			buff.Write(Encode(string(t)))

		case []interface{}, []string:
			// Inline function for writing the slice header
			WriteSliceHeader := func(length int) {
				if length < 56 {
					buff.WriteByte(byte(length + 128))
				} else {
					b2 := ToBin(uint64(length), 0)
					buff.WriteByte(byte(len(b2) + 183))
					buff.WriteString(b2)
				}
			}

			// FIXME How can I do this "better"?
			if interSlice, ok := t.([]interface{}); ok {
				WriteSliceHeader(len(interSlice))
				for _, val := range interSlice {
					buff.Write(Encode(val))
				}
			} else if stringSlice, ok := t.([]string); ok {
				WriteSliceHeader(len(stringSlice))
				for _, val := range stringSlice {
					buff.Write(Encode(val))
				}
			}
		}
	} else {
		// Empty list for nil
		buff.WriteByte(0x80)
	}

	return buff.Bytes()
}
