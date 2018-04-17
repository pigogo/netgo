// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"syscall"
)

const (
	maxListenEvent = 64
)

var (
	pollEvent *epoll
)

type epoll struct {
	fd int
	epollEvents [maxListenEvent]syscall.EpollEvent
	activityEventCnt int
}

func newEpoll() {
	if pollEvent != nil {
		return
	}

	efd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	pollEvent = &epoll {
		fd: efd,
	}
}

func (ep * epoll) Add(fd int, sfd, pad int32, event uint32) error {
	e := &syscall.EpollEvent {
		Events : event,
		Fd : sfd,
		Pad : pad,
	}
	return syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, fd, e)
}

func (ep * epoll) Mod(fd int, sfd, pad int32, event uint32) error {
	e := &syscall.EpollEvent {
		Events : event,
		Fd : sfd,
		Pad : pad,
	}
	return syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_MOD, fd, e)
}

func (ep * epoll) Del(fd int) error {
	return syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (ep * epoll) DelWrite(fd int, sfd, pad int32) error {
	e := &syscall.EpollEvent {
		Events : syscall.EPOLLOUT,
		Fd : sfd,
		Pad : pad,
	}

	return syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_DEL, fd, e)
}

func (ep * epoll) Wait(msec int) (err error) {
	ep.activityEventCnt, err = syscall.EpollWait(ep.fd, ep.epollEvents[:], msec)
	return
}