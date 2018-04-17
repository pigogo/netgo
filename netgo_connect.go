// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"net"
	"syscall"
)

//todo: proto support tcp4, tcp6
func parseAddr(addr string) (saddr syscall.Sockaddr, err error) {
	taddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}

	switch len(taddr.IP) {
	case 0:
		break
	case 4:
		sa4 := &syscall.SockaddrInet4 {
			Port: taddr.Port,
		}
		copy(sa4.Addr[:], taddr.IP[:])
		saddr = sa4
		return
	case 16:
		sa6 := &syscall.SockaddrInet6{
			Port: taddr.Port,
		}
		copy(sa6.Addr[:], taddr.IP[:])
		saddr = sa6
		return
	}
	return nil, ErrInvalidConnectAddr
}

func (nio *netio) Connect(addr string) (naddr int32, err error) {
	stream := nio.aollocStream()
	if stream == nil {
		err = ErrNoSocketAvailable
		return
	}

	var (
		remoteAddr syscall.Sockaddr
	)

	for {
		remoteAddr, err = parseAddr(addr) 
		if err != nil {
			break
		}

		err = nio.rawConnect(stream, remoteAddr)
		if err != nil {
			break
		}

		naddr = stream.NetAddr()
		return
	}

	nio.freeStream(stream)
	naddr = INVALIDADDR
	return
}

func (nio * netio) rawConnect(stream *netStream, raddr syscall.Sockaddr) (err error) {
	var fd int
	defer func() {
		if err != nil && fd >  0 {
			syscall.Close(fd)
		}
	}()

	switch raddr.(type) {
	case *syscall.SockaddrInet4:
		fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
		if err != nil {
			return 
		}
		break
	case *syscall.SockaddrInet6:
		fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
		if err != nil {
			return
		}
		break
	}
	
	err = syscall.SetNonblock(fd, true)
	if err != nil {
		return
	}

	if nio.option.Keepalive {
		if err = SetKeepAlive(fd, nio.option.KeepIdle); err != nil {
			return
		}
	}

	if nio.option.Nodelay {
		if err = SetNodelay(fd); err != nil {
			return
		}
	}

	if err = SetLinger(fd, nio.option.Linger); err != nil {
		return
	}

	err = syscall.Connect(fd, raddr)
	if err != nil && err != syscall.EINPROGRESS{
		return
	}

	sstate := streamStateOpened
	event := syscall.EPOLLERR
	if err == nil {
		event |= syscall.EPOLLIN
	} else {
		event |= syscall.EPOLLOUT
		sstate = streamStateOpening
		nio.toutStream.Push(stream.NetAddr())
	}

	stream.SetState(sstate)
	stream.SetType(streamTypeConnect)
	stream.SetFD(fd, nil, raddr)
	if err = pollEvent.Add(fd, int32(fd), stream.NetAddr(), uint32(event)); err != nil {
		syscall.Close(fd)
		return
	}

	return
}