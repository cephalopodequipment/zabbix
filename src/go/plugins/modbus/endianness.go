/*
** Zabbix
** Copyright (C) 2001-2020 Zabbix SIA
**
** This program is free software; you can redistribute it and/or modify
** it under the terms of the GNU General Public License as published by
** the Free Software Foundation; either version 2 of the License, or
** (at your option) any later version.
**
** This program is distributed in the hope that it will be useful,
** but WITHOUT ANY WARRANTY; without even the implied warranty of
** MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
** GNU General Public License for more details.
**
** You should have received a copy of the GNU General Public License
** along with this program; if not, write to the Free Software
** Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
**/

package modbus

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

func pack2Json(val []byte, p *MBParams) (jdata interface{}, err error) {

	if len(val) < int(p.Offset*2) {
		return nil, fmt.Errorf("Wrong length of received data: %d", len(val))
	}

	if p.Offset > 0 {
		val = val[p.Offset*2:]
	}

	if p.RetType == Bit {
		ar := getArr16(p.RetType, p.RetCount, val)
		if p.RetCount == 1 {
			return getFirst(ar), nil
		}
		jd, jerr := json.Marshal(ar)
		if jerr != nil {
			return nil, fmt.Errorf("Unable to create json: %s", jerr)
		}
		return string(jd), nil
	}

	var typeSize int
	switch p.RetType {
	case Uint64, Double:
		typeSize = 8
	case Int32, Uint32, Float:
		typeSize = 4
	case Int16, Uint16:
		typeSize = 2
	default:
		typeSize = 1
	}

	var arr interface{}
	if typeSize == 1 && p.Endianness.order == binary.LittleEndian {
		arr = swapPairByte(val, p.RetType, p.RetCount)
	} else {
		arr = makeRetArray(p.RetType, p.RetCount)
		r := bytes.NewReader(val)
		binary.Read(r, p.Endianness.order, arr)
	}

	if typeSize > 2 && 0 != p.Endianness.middle {
		arr = middlePack(arr, p.RetType)
	}

	if p.RetCount == 1 {
		return getFirst(arr), nil
	}

	if p.RetType == Uint8 {
		arr = getArr16(p.RetType, p.RetCount, arr.([]byte))
	}

	jd, jerr := json.Marshal(arr)
	if jerr != nil {
		return nil, fmt.Errorf("Unable to create json: %s", jerr)
	}
	return string(jd), nil
}

func swapPairByte(v []byte, retType Bits16, retCount uint) (ret interface{}) {
	switch retType {
	case Int8:
		ret = make([]int8, len(v))
		for i := 0; i < len(v)-1; i += 2 {
			ret.([]int8)[i] = int8(v[i+1])
			ret.([]int8)[i+1] = int8(v[i])
		}
		ret = ret.([]int8)[:retCount]
	case Uint8:
		ret = make([]byte, len(v))
		for i := 0; i < len(v)-1; i += 2 {
			ret.([]byte)[i] = v[i+1]
			ret.([]byte)[i+1] = v[i]
		}
		ret = ret.([]byte)[:retCount]
	}
	return ret
}

func getArr16(retType Bits16, retCount uint, val []byte) interface{} {
	ar := make([]uint16, retCount)
	for i := range val {
		if retType == Bit {
			for j := 0; j < 8; j++ {
				ar[i*8+j] = uint16(val[i] & (1 << j) >> j)
				if retCount--; retCount == 0 {
					return ar
				}
			}
		} else {
			ar[i] = uint16(val[i])
		}
	}
	return ar
}

func middlePack(v interface{}, rt Bits16) interface{} {
	buf := new(bytes.Buffer)
	switch rt {
	case Uint64:
		for i, num := range v.([]uint64) {
			binary.Write(buf, binary.BigEndian, &num)
			bs := buf.Bytes()
			bs = []byte{bs[2], bs[3], bs[0], bs[1], bs[6], bs[7], bs[4], bs[5]}
			rd := bytes.NewReader(bs)
			binary.Read(rd, binary.BigEndian, &num)
			v.([]uint64)[i] = num
			buf.Reset()
		}
	case Double:
		for i, num := range v.([]float64) {
			binary.Write(buf, binary.BigEndian, &num)
			bs := buf.Bytes()
			bs = []byte{bs[2], bs[3], bs[0], bs[1], bs[6], bs[7], bs[4], bs[5]}
			rd := bytes.NewReader(bs)
			binary.Read(rd, binary.BigEndian, &num)
			v.([]float64)[i] = num
			buf.Reset()
		}
	case Int32:
		for i, num := range v.([]int32) {
			binary.Write(buf, binary.BigEndian, &num)
			bs := buf.Bytes()
			bs = []byte{bs[2], bs[3], bs[0], bs[1]}
			rd := bytes.NewReader(bs)
			binary.Read(rd, binary.BigEndian, &num)
			v.([]int32)[i] = num
			buf.Reset()
		}
	case Uint32:
		for i, num := range v.([]uint32) {
			binary.Write(buf, binary.BigEndian, &num)
			bs := buf.Bytes()
			bs = []byte{bs[2], bs[3], bs[0], bs[1]}
			rd := bytes.NewReader(bs)
			binary.Read(rd, binary.BigEndian, &num)
			v.([]uint32)[i] = num
			buf.Reset()
		}
	case Float:
		for i, num := range v.([]float32) {
			binary.Write(buf, binary.BigEndian, &num)
			bs := buf.Bytes()
			bs = []byte{bs[2], bs[3], bs[0], bs[1]}
			rd := bytes.NewReader(bs)
			binary.Read(rd, binary.BigEndian, &num)
			v.([]float32)[i] = num
			buf.Reset()
		}
	}
	return v
}

func makeRetArray(retType Bits16, arraySize uint) (v interface{}) {
	switch retType {
	case Uint64:
		v = make([]uint64, arraySize)
	case Double:
		v = make([]float64, arraySize)
	case Int32:
		v = make([]int32, arraySize)
	case Uint32:
		v = make([]uint32, arraySize)
	case Float:
		v = make([]float32, arraySize)
	case Int16:
		v = make([]int16, arraySize)
	case Uint16:
		v = make([]uint16, arraySize)
	case Int8:
		v = make([]int8, arraySize)
	case Uint8:
		v = make([]byte, arraySize)
	}
	return v
}

func getFirst(v interface{}) interface{} {
	switch v.(type) {
	case []uint64:
		return v.([]uint64)[0]
	case []float64:
		return v.([]float64)[0]
	case []uint32:
		return v.([]uint32)[0]
	case []int32:
		return v.([]int32)[0]
	case []float32:
		return v.([]float32)[0]
	case []uint16:
		return v.([]uint16)[0]
	case []int16:
		return v.([]int16)[0]
	case []int8:
		return v.([]int8)[0]
	case []byte:
		return v.([]byte)[0]
	}
	return nil
}
