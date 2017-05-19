package gossdb

import (
	"encoding/json"
	"github.com/seefan/to"
	"time"
)

//扩展值，原始类型为 string
type Value string

//返回 string 的值
func (this Value) String() string {
	return string(this)
}

//返回 int64 的值
func (this Value) Int64() int64 {
	return to.Int64(this)
}

//返回 int32 的值
func (this Value) Int32() int32 {
	return int32(this.Int64())
}

//返回 int16 的值
func (this Value) Int16() int16 {
	return int16(this.Int64())
}

//返回 int8 的值
func (this Value) Int8() int8 {
	return int8(this.Int64())
}

//返回 int 的值
func (this Value) Int() int {
	return int(this.Int64())
}

//返回 uint64 的值
func (this Value) UInt64() uint64 {
	return to.Uint64(this)
}

//返回 uint32 类型的值
func (this Value) UInt32() uint32 {
	return uint32(this.UInt64())
}

//返回 uint16 类型的值
func (this Value) UInt16() uint16 {
	return uint16(this.UInt64())
}

//返回 uint8 类型的值
func (this Value) UInt8() uint8 {
	return uint8(this.UInt64())
}

//返回 byte 类型的值
func (this Value) Byte() byte {
	return this.UInt8()
}

//返回 uint 类型的值
func (this Value) UInt() uint {
	return uint(this.UInt64())
}

//返回 float64 类型的值
func (this Value) Float64() float64 {
	return to.Float64(this)
}

//返回 float32 类型的值
func (this Value) Float32() float32 {
	return float32(this.Float64())
}

//返回 bool 类型的值
func (this Value) Bool() bool {
	return to.Bool(this)
}

//返回 time.Time 类型的值
func (this Value) Time() time.Time {
	return to.Time(this)
}

//返回 time.Duration 类型的值
func (this Value) Duration() time.Duration {
	return to.Duration(this)
}

//返回 []byte 类型的值
func (this Value) Bytes() []byte {
	return []byte(this)
}

//判断是否为空
func (this Value) IsEmpty() bool {
	return this == ""
}

//按json 转换指定类型
//
//  value 传入的指针
//
//示例
//  var abc time.Time
//  v.As(&abc)
func (this Value) As(value interface{}) (err error) {
	err = json.Unmarshal(this.Bytes(), value)
	return
}
