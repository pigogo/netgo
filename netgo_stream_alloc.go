// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"sync"
)

const (
	SeqidMASK = 0x7FF
	OffsetMASK = 0xFFFFF
)

var (
	addrMutex sync.RWMutex
)

func (nio *netio) makeStream(offset int32) *netStream {
	if offset >= maxSocketNumber {
		return nil
	}

	if nio.streams[offset] == nil {
		nio.streams[offset] = &netStream{
			netDriver: nio,
			sindex: offset,
			onPacket: nio.onPacket,
		}
		nio.streams[offset].Init()
	}

	return nio.streams[offset]
}

//check seqid
func (nio *netio) selectStream(naddr int32) *netStream{
	offset := nio.getOffset(naddr)
	if offset >= maxSocketNumber {
		return nil
	}
	stream := nio.streams[offset]
	if stream.NetAddr() != naddr {
		return nil
	}
	return stream
}

func (nio *netio) aollocStream() (stream *netStream) {
	addrMutex.Lock()
	defer addrMutex.Unlock()

	if nio.lsindex.Len() > 0 {
		litem := nio.lsindex.Front()
		nio.lsindex.Remove(litem)
		naddr := litem.Value.(int32)
		offset := nio.getOffset(naddr)
		if nio.streams[offset] == nil {
			nio.streams[offset] = nio.makeStream(offset)
		} else {
			nio.streams[offset].Init()
		}
		return nio.streams[offset]
	}

	if int(nio.sused) < maxSocketNumber {
		if nio.streams[nio.sused] == nil {
			nio.streams[nio.sused] = nio.makeStream(nio.sused)
		}
		
		stream = nio.streams[nio.sused]
		nio.sused++
		return
	}

	return nil
}

func (nio *netio) freeStream(stream *netStream) {
	addrMutex.Lock()
	defer addrMutex.Unlock()
	
	if stream == nil {
		return
	}

	stream.Close()
	stream.Reset()
	nio.lsindex.PushBack(stream.NetAddr())
}

func (nio *netio) getOffset(addr int32) int32 {
	return addr & OffsetMASK
}

func (nio *netio) getSeqid(addr int32) int32 {
	return (addr >> 20) & SeqidMASK
}