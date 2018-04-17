// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"bytes"
	"syscall"
	"sync/atomic"
)

func (stream * netStream) SendPacket(naddr int32, packet *netPack) (err error) {
	err = stream.CacheSendPacket(packet)
	if err != nil {
		return
	}

	if stream.State() != streamStateOpened {
		return
	}

	//trigger routine 
	select {
	case stream.eventOut <- 1:
		break
	default:
		break
	}
	return
}

func (stream *netStream) CacheSendPacket(pack *netPack) error {
	if atomic.LoadUint32(&stream.state) == streamStateUnknow {
		select {
		case stream.sendCh<-pack:
			return nil
		default:
			return ErrOutcacheFull
		}
	} else {
		select {
		case stream.sendCh<-pack:
			return nil
		}
	}
}

func (stream * netStream) send() {
	for {
		state := atomic.LoadUint32(&stream.state)
		if state == streamStateOpened {
			break
		}
		
		return
	}
	
	var (
		err error
	)

	stream.sendGroup.Add(1)
	for {
		if stream.cachedDataLen() == 0 && stream.hasBackPacket() == false {
			break
		}
		
		if stream.outpack == nil {
			stream.aggregatePacket()
		}
		
		err = stream.rawSend()
		if err != nil {
			if err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
				stream.sendGroup.Done()
				stream.rawClose()
				return
			}
			break
		}
		continue
	}
	stream.sendGroup.Done()
}

func (stream *netStream) rawSend() (err error) {
	breakLoop := false
	for breakLoop == false {
		lastSended := 0
		offset := stream.sendOffset
		if stream.cachedDataLen () > 0 {
			for offset < stream.cached {
				lastSended, err = stream.netDriver.rawSend(stream.netaddr, stream.fd, stream.outBuffer[offset:stream.cached])
				if lastSended < 0 && err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
					return
				}

				if lastSended > 0 {
					offset += lastSended
				}

				if lastSended < 0 || err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
					breakLoop = true
					break
				}
			}

			stream.sendOffset = offset
			if stream.sendOffset == stream.cached {
				stream.sendOffset = 0
				stream.cached = 0
			} else if stream.cached - stream.sendOffset < 64 || stream.sendOffset > 0xFFFF - 0xF00 {
				copy(stream.outBuffer[:], stream.outBuffer[stream.sendOffset:stream.cached])
				stream.cached = stream.cached - stream.sendOffset
				stream.sendOffset = 0
			}
		} else { //send packet too long to be aggregated
			//no more to send
			if stream.outpack == nil {
				break
			}
	
			for {
				//send header
				if offset < netPackHeadLen {
					lastSended, err = stream.netDriver.rawSend(stream.netaddr, stream.fd, stream.outpack.binHeader[offset:])
					if lastSended < 0 && err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
						return
					}
				} else {
					lastSended, err = stream.netDriver.rawSend(stream.netaddr, stream.fd, stream.outpack.body[offset - netPackHeadLen:])
					if lastSended < 0 && err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
						return
					}
				}
				
				if lastSended > 0 {
					offset += lastSended
					if offset == netPackHeadLen + int(stream.outpack.header.bodyLen) {
						stream.outpack = nil
						offset = 0
						break
					}
				}

				if lastSended < 0 || err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
					breakLoop = true
					break
				}			
			}
			stream.sendOffset = offset
			//finish send the big pack
			if stream.outpack == nil {
				stream.aggregatePacket()
			}
		}
	}

	return
}

func (stream *netStream) hasBackPacket() bool {
	return stream.outpack != nil || len(stream.sendCh) > 0
}

func (stream * netStream) cachedDataLen() int {
	return stream.cached - stream.sendOffset
}

func (stream * netStream) aggregatePacket() bool {
	var (
		n int
		err error
		newCached int
		maxCanCache int
		writer *bytes.Buffer
	)

	maxCanCache = len(stream.outBuffer) - stream.cached
	writer = bytes.NewBuffer(stream.outBuffer[stream.cached:stream.cached])
	//aggregrate netpack
	for {
		if stream.outpack == nil {
			select {
			case stream.outpack = <- stream.sendCh:
				break
			default:
				break
			}
		}

		if stream.outpack == nil {
			break
		}

		if newCached + netPackHeadLen + int(stream.outpack.header.bodyLen) > maxCanCache {
			break
		}

		//cache header
		if n, err = writer.Write(stream.outpack.binHeader[:]); err != nil || n != netPackHeadLen {
			panic(ErrUnknowSysError)
		}
		newCached += n
		
		//cache body
		if n, err = writer.Write(stream.outpack.body); err != nil || n != int(stream.outpack.header.bodyLen) {
			panic(ErrUnknowSysError)
		}
		newCached += n
		stream.outpack = nil
	}
	stream.cached += newCached
	return (newCached > 0)
}
