// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"syscall"
)

func SetKeepAlive(fd, secs int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs); err != nil {
		return err
	}
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs)
}

func SetReusedPort(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, 0xf, 1)
}

func SetReusedAddr(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}

func SetLinger(fd, sec int) error {
	var l syscall.Linger
	if sec >= 0 {
		l.Onoff = 1
		l.Linger = int32(sec)
	} else {
		l.Onoff = 0
		l.Linger = 0
	}
	return syscall.SetsockoptLinger(fd, syscall.SOL_SOCKET, syscall.SO_LINGER, &l)
}

func SetNodelay(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
}

// SetsockoptInt wraps the setsockopt network call with an int argument.
func  SetsockoptInt(fd, level, name, arg int) error {
	return syscall.SetsockoptInt(fd, level, name, arg)
}

// SetsockoptInet4Addr wraps the setsockopt network call with an IPv4 address.
func  SetsockoptInet4Addr(fd, level, name int, arg [4]byte) error {
	return syscall.SetsockoptInet4Addr(fd, level, name, arg)
}

// SetsockoptLinger wraps the setsockopt network call with a Linger argument.
func  SetsockoptLinger(fd, level, name int, l *syscall.Linger) error {
	return syscall.SetsockoptLinger(fd, level, name, l)
}
