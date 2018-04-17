// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import(
	"syscall"
	"bytes"
)

func (nio *netio) Send(naddr int32, buf[]byte) (err error) {
	stream := nio.selectStream(naddr)
	if stream == nil || !stream.PreSend() {
		return ErrInvalidSocketAddr
	}

	packet := &netPack{
		header: netPackHead {
			magic: netpackMagic,
			bodyLen: uint32(len(buf)),
			target: 0,//todo:target
		},
		body : buf,
		naddr: naddr,
	}
	
	headbuf := bytes.NewBuffer(packet.binHeader[:0])
	if err = encoderHeader(headbuf, &packet.header); err != nil {
		return err
	}

	return stream.SendPacket(naddr, packet)
}

func (nio *netio) rawSend(naddr int32, fd FD, buf[]byte) (totalSend int, err error) {
	if fd.fd < 0 {
		return -1, ErrInvalidSocketAddr
	}

	buflen := len(buf)
	var lastSend int 
	for totalSend < buflen {
		lastSend, err = syscall.Write(fd.fd, buf[totalSend:])
		if lastSend < 0 && err != syscall.EINTR || lastSend == 0{
			break
		}

		totalSend += lastSend
	}

	if totalSend == buflen {
		return
	}

	if lastSend < 0 && err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
		return -1, err
	}

	if pollEvent.Mod(fd.fd, int32(fd.fd), naddr, syscall.EPOLLIN|syscall.EPOLLOUT|syscall.EPOLLERR) != nil {
		return -1, err
	}
	return
}