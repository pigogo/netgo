// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import(
	"io"
	"encoding/binary"
)

const (
	netpackMagic = 0xabcdabcd
)

type netPackHead struct {
	magic uint32
	bodyLen uint32
	target uint64
}

func (pack *netPackHead) Valid() bool {
	return pack.magic == netpackMagic
}

func decodeHeader(reader io.Reader) (header*netPackHead, err error) {
	header = &netPackHead{}
	header.magic, err = getUint32(reader, binary.BigEndian)
	if err != nil {
		return
	}

	header.bodyLen, err = getUint32(reader, binary.BigEndian)
	if err != nil {
		return
	}

	header.target, err = getUint64(reader, binary.BigEndian)
	return
}

func encoderHeader(writer io.Writer, header *netPackHead) (err error) {
	err = putUint32(header.magic, writer, binary.BigEndian)
	if err != nil {
		return
	}

	err = putUint32(header.bodyLen, writer, binary.BigEndian)
	if err != nil {
		return
	}
	
	return putUint64(header.target, writer, binary.BigEndian)
} 