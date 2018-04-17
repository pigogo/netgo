// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo


import (
	"syscall"
	"bytes"
)

func (stream * netStream) recv() (err error){
	var (
		header *netPackHead
		n int
	)
	
	stream.recvGroup.Add(1)
	recvd := stream.recvd
exit:
	for {
		n, err = stream.netDriver.rawRecv(stream.fd, stream.inBuffer[recvd:])		
		if n <= 0 {
			if err != nil && err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
				stream.recvGroup.Done()
				stream.rawClose()
				return
			}
			break
		} else {
			recvd += uint32(n)
		}
		
		bodyOffset := uint32(0)
		for {
			//parse netpack header
			if recvd - bodyOffset >= netPackHeadLen {
				buf := bytes.NewBuffer(stream.inBuffer[bodyOffset:])
				header, err = decodeHeader(buf)
				if err != nil {
					stream.recvGroup.Done()
					stream.rawClose()
					return
				}

				if header.Valid() == false {
					err = ErrInvalidPacketRecv
					stream.recvGroup.Done()
					stream.rawClose()
					return
				}
			} else {
				break
			}

			//get body
			if recvd - bodyOffset < header.bodyLen + netPackHeadLen {
				if header.bodyLen + netPackHeadLen > 0xFFFF { //too long to be cached
					stream.packflag = true
					//init in packet
					if stream.inpack == nil {
						stream.inpack = &netPack {
							header: *header,
							body: make([]byte, header.bodyLen),
						}
					} else {
						if stream.inpack.header.bodyLen < header.bodyLen {
							stream.inpack.body = make([]byte, header.bodyLen)
						}
						stream.inpack.header.bodyLen = header.bodyLen
					}
					copy(stream.inpack.body[:], stream.inBuffer[bodyOffset + netPackHeadLen:recvd])
					recvd -= bodyOffset
					break exit
				}
				break
			}

			bodyOffset += uint32(netPackHeadLen)
			stream.onPacket(stream.inBuffer[bodyOffset:bodyOffset + header.bodyLen], stream.netaddr)
			bodyOffset += header.bodyLen
		}

		//move unused data
		if bodyOffset > 0 {
			copy(stream.inBuffer[:], stream.inBuffer[bodyOffset:recvd])
			recvd -= bodyOffset
		}
	}
	
	stream.recvd = recvd
	stream.recvGroup.Done()
	return 
}

func (stream *netStream) recvPack() (err error) {
	var (
		n int
	)
	
	stream.recvGroup.Add(1)
	recvd := stream.recvd
	for {
		n, err = stream.netDriver.rawRecv(stream.fd, stream.inpack.body[recvd - netPackHeadLen:stream.inpack.header.bodyLen])		
		if n <= 0 {
			if err != nil && err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
				stream.recvGroup.Done()
				stream.rawClose()
				return
			}
			break
		} else {
			recvd += uint32(n)
		}
		
		//get a packet
		if recvd == netPackHeadLen + stream.inpack.header.bodyLen {
			stream.onPacket(stream.inpack.body[:stream.inpack.header.bodyLen], stream.netaddr)
			recvd = 0
			stream.packflag = false
			break
		}
	}

	stream.recvd = recvd
	stream.recvGroup.Done()
	return
}