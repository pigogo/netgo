// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import(
	"syscall"
	"net"
)

func (nio *netio) Listen(addr string) (err error) {
	stream := nio.aollocStream() 
	if stream == nil {
		err = ErrNoSocketAvailable
		return
	}

	var lfd int
	for {
		lfd, err = nio.rawListen(stream, addr)
		if err != nil {
			break		
		}
		if err = pollEvent.Add(lfd, int32(lfd), stream.NetAddr(), uint32(syscall.EPOLLIN|syscall.EPOLLERR)); err != nil {
			panic(err)
		}
		return
	}

	nio.freeStream(stream)

	return
}

func (nio *netio) accept(saddr int32) (err error) {
	stream := nio.selectStream(saddr)
	if stream == nil || !stream.PreAccept() {
		return ErrInvalidSocketAddr
	}

	var (
		nfd int
		na syscall.Sockaddr
	)

	accStream := nio.aollocStream()
	for {
		if accStream == nil {
			return ErrNoSocketAvailable
		}

		nfd, na, err = syscall.Accept(stream.Sysfd())
		if err != nil {
			break
		}

		if err = syscall.SetNonblock(nfd, true); err != nil {
			break
		}

		if err = pollEvent.Add(nfd, int32(nfd), accStream.NetAddr(), uint32(syscall.EPOLLIN|syscall.EPOLLERR)); err != nil {
			break
		}

		accStream.SetState(streamStateOpened)
		accStream.SetType(streamTypeAccept)
		accStream.SetFD(nfd, nil, na)
		return
	}

	if nfd > 0 {
		syscall.Close(nfd)
	}

	nio.freeStream(accStream)
	return
}

func (nio * netio) resolveSocketAddr4(addr string) syscall.Sockaddr {
	raddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil
	}

	ip := raddr.IP
	if len(ip) == 0 {
		ip = net.IPv4zero
	}
	ret := &syscall.SockaddrInet4{Port:raddr.Port}
	copy(ret.Addr[:], ip.To4())
	return ret
}

func (nio *netio) rawListen(stream *netStream, addr string) (lfd int, err error) {
	defer func() {
		if err != nil && lfd != 0 {
			syscall.Close(lfd)
		}
	}()

	lfd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return
	}

	if err = syscall.SetNonblock(lfd, true); err != nil {
		return
	}

	if nio.option.Keepalive {
		if err = SetKeepAlive(lfd, nio.option.KeepIdle); err != nil {
			return
		}
	}

	if nio.option.Nodelay {
		if err = SetNodelay(lfd); err != nil {
			return
		}
	}

	if err = SetLinger(lfd, nio.option.Linger); err != nil {
		return
	}

	if nio.option.ReusedAddr {
		if err = SetReusedAddr(lfd); err != nil {
			return
		}
		if err = SetReusedPort(lfd); err != nil {
			return
		}
	}

	sysaddr := nio.resolveSocketAddr4(addr)
	if sysaddr == nil {
		err = ErrInvalidListenAddr
		return
	}
	
	err = syscall.Bind(lfd, sysaddr)
	if err != nil {
		return
	}

	backlog := nio.option.MaxListenBacklog
	if backlog <= 0 {
		backlog = maxListenBacklog
	}
	err = syscall.Listen(lfd, backlog)
	if err != nil {
		return
	}

	stream.SetState(streamStateOpened)
	stream.SetFD(lfd, sysaddr, nil)
	stream.SetType(streamTypeListen)
	return
}