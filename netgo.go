// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"time"
	"syscall"
	"fmt"
	"container/list"
)

const (
	connectingTimeout = time.Second * 3
)

type netio struct {
	streams []*netStream
	toutStream *toutQueue
	sindex map[uint32]uint32
	sused int32
	lsindex list.List
	option SocketOption
	pollEvent epoll
	closeChan chan struct{}
	onPacket func(body []byte, naddr int32)
}

func setRLimit() error {
	var rlimit, zero syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return fmt.Errorf("Getrlimit: save failed: %v", err)
	}
	if zero == rlimit {
		return fmt.Errorf("Getrlimit: save failed: got zero value %#v", rlimit)
	}

	if rlimit.Cur < maxSocketNumber && rlimit.Cur < rlimit.Max  {
		rlimit.Cur = maxSocketNumber + 1024
		if rlimit.Cur > rlimit.Max {
			rlimit.Cur = rlimit.Max
		}
		return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	}
	return nil
}

func (nio *netio) Init(onPacket func(body []byte, naddr int32), option SocketOption) error {
	if err := setRLimit(); err != nil {
		return err
	}

	nio.onPacket = onPacket
	nio.option = option
	nio.streams = make([]*netStream, maxSocketNumber)
	nio.sindex = make(map[uint32]uint32)
	nio.closeChan = make(chan struct{})
	nio.toutStream = newToutQue()
	if pollEvent == nil {
		newEpoll()
	}
	return nil
}

func (nio *netio) watchTimeoutConnecting() {
	for {
		totNode := nio.toutStream.Peek()
		if totNode == nil {
			break
		}

		if time.Since(totNode.tm) < connectingTimeout {
			break
		}

		nio.toutStream.Pop()
		stream := nio.selectStream(totNode.naddr)
		if stream == nil {
			continue
		}

		//try reconnect check
		if stream.PreConnect() {
			stream.rawClose()
		} else { //direct close
			nio.Close(stream.NetAddr())
		}
	}
}

func (nio *netio) handleEvent() bool {
	for i := 0; i < pollEvent.activityEventCnt; i++ {
		naddr := pollEvent.epollEvents[i].Pad
		stream := nio.selectStream(naddr)
		if stream == nil {
			pollEvent.Del(int(pollEvent.epollEvents[i].Fd))
			continue
		}

		if syscall.EPOLLOUT & pollEvent.epollEvents[i].Events != 0 {
			pollEvent.Mod(stream.Sysfd(), pollEvent.epollEvents[i].Fd, pollEvent.epollEvents[i].Pad, syscall.EPOLLIN | syscall.EPOLLERR)
		}

		//remove timeout watch
		if stream.Opening() {
			nio.toutStream.Remove(naddr)
		}

		stream.OnEvent(pollEvent.epollEvents[i].Events)
		if stream.Closed() && stream.PreConnect() == false {
			nio.Close(stream.NetAddr())
		}
	}
	return true
}

func (nio *netio) poll() error {
	var(
		tickTime time.Duration
		nextTickTime time.Duration
		tnow time.Time
	)
	tickTime = time.Millisecond * 200
	nextTickTime = tickTime
ploop:
	for {
		tnow = time.Now()
		pollEvent.Wait(int(nextTickTime))
		nio.watchTimeoutConnecting()
		nio.handleEvent()
		//get next tickTime
		nextTickTime = tickTime - time.Now().Sub(tnow) / time.Millisecond
		if nextTickTime < 0 {
			nextTickTime = 0
		}

		select {
		case <-nio.closeChan:
			break ploop
		default:
			continue
		}
	}
	return nil
}

func (nio *netio) Serve() error {
	go nio.poll()

sloop:
	for {
		select {
		case <-nio.closeChan:
			break sloop
		}
	}

	syscall.Close(pollEvent.fd)
	pollEvent = nil
	return nil
}

//todo: cosafe
func (nio *netio) Close(addr int32) error {
	stream := nio.selectStream(addr)
	if stream == nil {
		return ErrInvalidSocketAddr
	}

	nio.freeStream(stream)
	return nil
}
