// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"syscall"
)

func (nio * netio) rawRecv(fd FD, buf[]byte) (totalRecv int, err error) {
	if fd.fd < 0 {
		return -1, ErrSocketBeClosed
	}

	for {
		totalRecv, err = syscall.Read(fd.fd, buf)
		if totalRecv < 0 && err == syscall.EINTR{
			continue
		}

		break
	} 

	if totalRecv == len(buf) {
		return
	}

	if totalRecv == 0 || (totalRecv < 0 && err !=  syscall.EAGAIN && err != syscall.EWOULDBLOCK) {
		totalRecv = -1
		if err == nil {
			err = ErrSocketBeClosed
		}
		return
	}

	return
}