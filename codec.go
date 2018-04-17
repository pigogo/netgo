// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"fmt"
	"io"
	"encoding/binary"
	"unsafe"
)

var (
	localEndian binary.ByteOrder = getLocalEndian()
)

func getLocalEndian() binary.ByteOrder {
	var x uint16 = 0x1
	if (*[2]byte)(unsafe.Pointer(&x))[0] == 0 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

func swap_2(to []byte, from []byte) {
	to[0], to[1] = from[1], from[0]
}

func swap_4(to []byte, from []byte) {
	to[0], to[1], to[2], to[3] = from[3], from[2], from[1], from[0]
}

func swap_8(to []byte, from []byte) {
	to[0], to[1], to[2], to[3], to[4], to[5], to[6], to[7] = from[7], from[6], from[5], from[4], from[3], from[2], from[1], from[0]
}

func getUint32(reader io.Reader, order binary.ByteOrder) (res uint32, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	_, err = reader.Read((*[4]byte)(unsafe.Pointer(&res))[:])
	if err != nil {
		return
	}

	if order != localEndian {
		swap_4((*[4]byte)(unsafe.Pointer(&res))[:], (*[4]byte)(unsafe.Pointer(&res))[:])
	}
	return
}

func putUint32(val uint32, writer io.Writer, order binary.ByteOrder) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if order != localEndian {
		swap_4((*[4]byte)(unsafe.Pointer(&val))[:], (*[4]byte)(unsafe.Pointer(&val))[:])
	}

	_, err = writer.Write((*[4]byte)(unsafe.Pointer(&val))[:])
	return
}


func getUint64(reader io.Reader, order binary.ByteOrder) (res uint64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	_, err = reader.Read((*[8]byte)(unsafe.Pointer(&res))[:])
	if err != nil {
		return
	}

	if order != localEndian {
		swap_8((*[8]byte)(unsafe.Pointer(&res))[:], (*[8]byte)(unsafe.Pointer(&res))[:])
	}
	return
}

func putUint64(val uint64, writer io.Writer, order binary.ByteOrder) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if order != localEndian {
		swap_8((*[8]byte)(unsafe.Pointer(&val))[:], (*[8]byte)(unsafe.Pointer(&val))[:])
	}

	_, err = writer.Write((*[8]byte)(unsafe.Pointer(&val))[:])
	return
}