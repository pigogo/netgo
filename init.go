// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"syscall"
)

func init() {
	efd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	pollEvent = &epoll {
		fd: efd,
	}
}