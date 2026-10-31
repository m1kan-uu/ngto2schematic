package main

// 最小限の NBT 読み書き (型付き・キー順保持)

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

const (
	TagEnd byte = iota
	TagByte
	TagShort
	TagInt
	TagLong
	TagFloat
	TagDouble
	TagByteArray
	TagString
	TagList
	TagCompound
	TagIntArray
	TagLongArray
)

type Tag struct {
	T byte
	V interface{} // int8,int16,int32,int64,float32,float64,[]byte,string,*List,*Compound,[]int32,[]int64
}

type List struct {
	Elem  byte
	Items []*Tag
}

type Compound struct {
	Keys []string
	M    map[string]*Tag
}

func NewCompound() *Compound { return &Compound{M: map[string]*Tag{}} }

func (c *Compound) Get(k string) *Tag { return c.M[k] }
func (c *Compound) Set(k string, t *Tag) {
	if _, ok := c.M[k]; !ok {
		c.Keys = append(c.Keys, k)
	}
	c.M[k] = t
}
func (c *Compound) Del(k string) {
	if _, ok := c.M[k]; !ok {
		return
	}
	delete(c.M, k)
	for i, kk := range c.Keys {
		if kk == k {
			c.Keys = append(c.Keys[:i], c.Keys[i+1:]...)
			break
		}
	}
}
func (c *Compound) Clone() *Compound {
	n := NewCompound()
	for _, k := range c.Keys {
		n.Set(k, c.M[k])
	}
	return n
}

// 数値として取り出す (型は問わない)
func (c *Compound) Int(k string, def int) int {
	t := c.M[k]
	if t == nil {
		return def
	}
	switch v := t.V.(type) {
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	}
	return def
}
func (c *Compound) Str(k string) string {
	if t := c.M[k]; t != nil {
		if s, ok := t.V.(string); ok {
			return s
		}
	}
	return ""
}
func (c *Compound) Comp(k string) *Compound {
	if t := c.M[k]; t != nil {
		if s, ok := t.V.(*Compound); ok {
			return s
		}
	}
	return nil
}
func (c *Compound) List(k string) *List {
	if t := c.M[k]; t != nil {
		if s, ok := t.V.(*List); ok {
			return s
		}
	}
	return nil
}

type reader struct {
	r   *bufio.Reader
	err error
}

func (r *reader) read(n int) []byte {
	if r.err != nil {
		return make([]byte, n)
	}
	b := make([]byte, n)
	_, r.err = io.ReadFull(r.r, b)
	return b
}
func (r *reader) u8() byte    { return r.read(1)[0] }
func (r *reader) u16() uint16 { return binary.BigEndian.Uint16(r.read(2)) }
func (r *reader) u32() uint32 { return binary.BigEndian.Uint32(r.read(4)) }
func (r *reader) u64() uint64 { return binary.BigEndian.Uint64(r.read(8)) }
func (r *reader) str() string { return string(r.read(int(r.u16()))) }

func (r *reader) payload(t byte, depth int) *Tag {
	if depth > 512 {
		r.err = errors.New("NBT が深すぎます")
	}
	if r.err != nil {
		return &Tag{T: t}
	}
	switch t {
	case TagByte:
		return &Tag{t, int8(r.u8())}
	case TagShort:
		return &Tag{t, int16(r.u16())}
	case TagInt:
		return &Tag{t, int32(r.u32())}
	case TagLong:
		return &Tag{t, int64(r.u64())}
	case TagFloat:
		return &Tag{t, math.Float32frombits(r.u32())}
	case TagDouble:
		return &Tag{t, math.Float64frombits(r.u64())}
	case TagByteArray:
		n := int32(r.u32())
		if n < 0 || n > 1<<28 {
			r.err = errors.New("配列長が不正")
			return &Tag{t, []byte{}}
		}
		return &Tag{t, r.read(int(n))}
	case TagString:
		return &Tag{t, r.str()}
	case TagList:
		et := r.u8()
		n := int32(r.u32())
		if n < 0 || n > 1<<24 {
			r.err = errors.New("リスト長が不正")
			return &Tag{t, &List{Elem: et}}
		}
		l := &List{Elem: et}
		for i := int32(0); i < n && r.err == nil; i++ {
			l.Items = append(l.Items, r.payload(et, depth+1))
		}
		return &Tag{t, l}
	case TagCompound:
		c := NewCompound()
		for r.err == nil {
			ct := r.u8()
			if ct == TagEnd {
				break
			}
			name := r.str()
			c.Set(name, r.payload(ct, depth+1))
		}
		return &Tag{t, c}
	case TagIntArray:
		n := int32(r.u32())
		if n < 0 || n > 1<<26 {
			r.err = errors.New("配列長が不正")
			return &Tag{t, []int32{}}
		}
		a := make([]int32, n)
		for i := range a {
			a[i] = int32(r.u32())
		}
		return &Tag{t, a}
	case TagLongArray:
		n := int32(r.u32())
		if n < 0 || n > 1<<25 {
			r.err = errors.New("配列長が不正")
			return &Tag{t, []int64{}}
		}
		a := make([]int64, n)
		for i := range a {
			a[i] = int64(r.u64())
		}
		return &Tag{t, a}
	}
	r.err = fmt.Errorf("不明なタグ %d", t)
	return &Tag{T: t}
}

// gzip でも生でも読める
func ReadNBT(data []byte) (*Compound, error) {
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		d, err := io.ReadAll(zr)
		if err != nil {
			return nil, err
		}
		data = d
	}
	r := &reader{r: bufio.NewReader(bytes.NewReader(data))}
	if r.u8() != TagCompound {
		return nil, errors.New("NBT のルートが Compound ではありません")
	}
	r.str()
	t := r.payload(TagCompound, 0)
	if r.err != nil {
		return nil, r.err
	}
	return t.V.(*Compound), nil
}

type writer struct{ b bytes.Buffer }

func (w *writer) u8(v byte) { w.b.WriteByte(v) }
func (w *writer) u16(v uint16) {
	var x [2]byte
	binary.BigEndian.PutUint16(x[:], v)
	w.b.Write(x[:])
}
func (w *writer) u32(v uint32) {
	var x [4]byte
	binary.BigEndian.PutUint32(x[:], v)
	w.b.Write(x[:])
}
func (w *writer) u64(v uint64) {
	var x [8]byte
	binary.BigEndian.PutUint64(x[:], v)
	w.b.Write(x[:])
}
func (w *writer) str(s string) { w.u16(uint16(len(s))); w.b.WriteString(s) }

func (w *writer) payload(t *Tag) {
	switch v := t.V.(type) {
	case int8:
		w.u8(byte(v))
	case int16:
		w.u16(uint16(v))
	case int32:
		w.u32(uint32(v))
	case int64:
		w.u64(uint64(v))
	case float32:
		w.u32(math.Float32bits(v))
	case float64:
		w.u64(math.Float64bits(v))
	case []byte:
		w.u32(uint32(len(v)))
		w.b.Write(v)
	case string:
		w.str(v)
	case *List:
		et := v.Elem
		if len(v.Items) == 0 {
			et = TagEnd
		}
		w.u8(et)
		w.u32(uint32(len(v.Items)))
		for _, it := range v.Items {
			w.payload(it)
		}
	case *Compound:
		for _, k := range v.Keys {
			ct := v.M[k]
			w.u8(ct.T)
			w.str(k)
			w.payload(ct)
		}
		w.u8(TagEnd)
	case []int32:
		w.u32(uint32(len(v)))
		for _, x := range v {
			w.u32(uint32(x))
		}
	case []int64:
		w.u32(uint32(len(v)))
		for _, x := range v {
			w.u64(uint64(x))
		}
	}
}

// gzip 圧縮して書き出す
func WriteNBT(rootName string, c *Compound) ([]byte, error) {
	w := &writer{}
	w.u8(TagCompound)
	w.str(rootName)
	w.payload(&Tag{TagCompound, c})
	var out bytes.Buffer
	zw := gzip.NewWriter(&out)
	if _, err := zw.Write(w.b.Bytes()); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// 小さなコンストラクタ
func TByte(v int) *Tag       { return &Tag{TagByte, int8(v)} }
func TShort(v int) *Tag      { return &Tag{TagShort, int16(v)} }
func TInt(v int) *Tag        { return &Tag{TagInt, int32(v)} }
func TDouble(v float64) *Tag { return &Tag{TagDouble, v} }
func TStr(v string) *Tag     { return &Tag{TagString, v} }
func TBytes(v []byte) *Tag   { return &Tag{TagByteArray, v} }
func TComp(c *Compound) *Tag { return &Tag{TagCompound, c} }
func TList(et byte, items []*Tag) *Tag {
	return &Tag{TagList, &List{Elem: et, Items: items}}
}
