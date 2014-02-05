package ethutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	_ "math"
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

// Data rlpValueutes are returned by the rlp decoder. The data rlpValueutes represents
// one item within the rlp data structure. It's responsible for all the casting
// It always returns something rlpValueid
type RlpValue struct {
	Value interface{}
	kind  reflect.Value
}

func (rlpValue *RlpValue) String() string {
	return fmt.Sprintf("%q", rlpValue.Value)
}

func Conv(rlpValue interface{}) *RlpValue {
	return &RlpValue{Value: rlpValue, kind: reflect.ValueOf(rlpValue)}
}

func NewRlpValue(rlpValue interface{}) *RlpValue {
	return &RlpValue{Value: rlpValue}
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

// Threat the rlpValueute as a slice
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

// RlpValue value setters
// An empty rlp value is always a list
func EmptyRlpValue() *RlpValue {
	return NewRlpValue([]interface{}{})
}

func (rlpValue *RlpValue) AppendList() *RlpValue {
	list := EmptyRlpValue()
	rlpValue.Value = append(rlpValue.AsSlice(), list)

	return list
}

func (rlpValue *RlpValue) Append(v interface{}) *RlpValue {
	rlpValue.Value = append(rlpValue.AsSlice(), v)

	return rlpValue
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

func ToBin(n uint64, length uint64) []byte {
	var buf bytes.Buffer
	for _, val := range ToBinarySlice(n, length) {
		buf.WriteByte(byte(val))
	}

	return buf.Bytes()
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

	var slice []interface{}
	char := int(data[pos])
	switch {
	case char <= 0x7c:
		return data[pos], pos + 1

	case char <= 0xb7:
		b := uint64(data[pos]) - 0x80

		return data[pos+1 : pos+1+b], pos + 1 + b

	case char <= 0xbf:
		b := uint64(data[pos]) - 0xb8
		b2 := uint64(FromBin(data[pos+1 : pos+1+b]))

		return data[pos+1+b : pos+1+b+b2], pos + 1 + b + b2

	case char <= 0xf7:
		b := uint64(data[pos]) - 0xc0
		prevPos := pos
		pos++
		for i := uint64(0); i < b; {
			var obj interface{}

			obj, prevPos = Decode(data, pos)
			slice = append(slice, obj)

			i += (prevPos - pos)
			pos = prevPos
		}
		return slice, pos

	case char <= 0xff:
		b := uint64(data[pos]) - 0xf8
		pos = pos + 1 + b
		prevPos := uint64(0)
		for i := uint64(0); i < b; {
			var obj interface{}

			obj, prevPos = Decode(data, pos)
			slice = append(slice, obj)

			i += (prevPos - pos)
			pos = prevPos
		}
		return slice, pos

	default:
		panic(fmt.Sprintf("byte not supported: %q", char))
	}

	return slice, 0
}

var (
	directRlp = big.NewInt(0x7f)
	numberRlp = big.NewInt(0xb7)
	zeroRlp   = big.NewInt(0x0)
)

func Encode(object interface{}) []byte {
	var buff bytes.Buffer

	if object != nil {
		switch t := object.(type) {
		case *RlpValue:
			buff.Write(Encode(t.AsRaw()))
		// Code dup :-/
		case int:
			buff.Write(Encode(big.NewInt(int64(t))))
		case uint:
			buff.Write(Encode(big.NewInt(int64(t))))
		case int8:
			buff.Write(Encode(big.NewInt(int64(t))))
		case int16:
			buff.Write(Encode(big.NewInt(int64(t))))
		case int32:
			buff.Write(Encode(big.NewInt(int64(t))))
		case int64:
			buff.Write(Encode(big.NewInt(t)))
		case uint16:
			buff.Write(Encode(big.NewInt(int64(t))))
		case uint32:
			buff.Write(Encode(big.NewInt(int64(t))))
		case uint64:
			buff.Write(Encode(big.NewInt(int64(t))))
		case byte:
			buff.Write(Encode(big.NewInt(int64(t))))
		case *big.Int:
			buff.Write(Encode(t.Bytes()))
		case []byte:
			if len(t) < 56 {
				buff.Write(append([]byte{byte(len(t) + 0x80)}, t...))
			} else {
				var b bytes.Buffer
				binary.Write(&b, binary.BigEndian, len(t))
				buff.Write(append(append([]byte{byte(len(b.Bytes()) + 0xb7)}, b.Bytes()...), t...))

				//b2 := ToBin(uint64(len(t)), 0)
				//buff.Write(append(append([]byte{byte(len(b2) + 0xb7)}, b2...), t...))
			}
		case string:
			buff.Write(Encode([]byte(t)))
			/*
				if len(t) < 56 {
					buff.Write(append([]byte{byte(len(t) + 0x80)}, t...))
				} else {
					b2 := ToBin(uint64(len(t)), 0)
					buff.Write(append(append([]byte{byte(len(b2) + 0xb7)}, b2...), t...))
				}
			*/

		case []interface{}:
			// Inline function for writing the slice header
			WriteSliceHeader := func(length int) {
				if length < 56 {
					buff.WriteByte(byte(length + 0xc0))
				} else {
					var b bytes.Buffer
					binary.Write(&b, binary.BigEndian, length)
					buff.WriteByte(byte(len(b.Bytes()) + 0xf7))
					buff.Write(b.Bytes())
				}
			}

			var b bytes.Buffer
			for _, val := range t {
				b.Write(Encode(val))
			}
			WriteSliceHeader(len(b.Bytes()))
			buff.Write(b.Bytes())
		}
	} else {
		// Empty list for nil
		buff.WriteByte(0x80)
	}

	return buff.Bytes()
}
