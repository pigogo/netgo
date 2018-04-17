// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"sync"
	"syscall"
	"time"
	"sync/atomic"
	"golang.org/x/net/context"
)

type streamType byte
const (
	streamTypeUnknow streamType = 0
	streamTypeListen streamType = 1 
	streamTypeAccept streamType = 2
	streamTypeConnect streamType = 3 
)

const (
	streamStateUnknow uint32 = 0
	streamStateOpening uint32 = 1
	streamStateOpened uint32 = 2
)

const (
	infinitTime =  time.Hour * 24 * 30 * 12 
	maxReconnectPeriod = time.Second * 30
	netPackHeadLen = 16
	maxNetPackLen = 8 * 1024 * 1024
	maxRecvCachePacket = 10000
	maxSendCachePacket = 10000
)

type netPack struct {
	header netPackHead
	binHeader [netPackHeadLen]byte	//for send 
	body []byte
	naddr int32
}

type FD struct {
	fd int
	localAddr syscall.Sockaddr
	remoteAddr syscall.Sockaddr
}

func (fd *FD) Init(f int, laddr, raddr syscall.Sockaddr) {
	fd.fd = f
	fd.localAddr = laddr
	fd.remoteAddr = raddr
}

func (fd *FD) Reset() {
	fd.fd = -1
	fd.localAddr = nil
	fd.remoteAddr = nil
}

type netStream struct {
	fd FD
	sindex int32
	seqid int32
	netaddr int32
	state uint32
	stype streamType

	//for recv
	inpack *netPack
	recvd uint32
	packflag bool
	recvGroup sync.WaitGroup

	//for send
	cached int
	sendOffset int
	outpack *netPack
	sendGroup sync.WaitGroup

	netDriver *netio
	closeCh chan struct{}
	netpackCh chan *netPack
	sendCh chan *netPack
	eventIn chan uint32
	eventOut chan uint32
	reconnectTimer *time.Timer
	timerPeriod time.Duration
	inBuffer [0xFFFF]byte
	outBuffer[0xFFFF]byte
	onPacket func(body []byte, naddr int32)
}

func (stream *netStream) Init() {
	if stream.closeCh != nil {
		close(stream.closeCh)
	}

	stream.timerPeriod = infinitTime
	stream.closeCh = make(chan struct{}, 1)
	stream.reconnectTimer = time.NewTimer(stream.timerPeriod)
	stream.sendCh = make(chan *netPack, maxSendCachePacket)
	stream.netpackCh = make(chan *netPack, maxRecvCachePacket)
	stream.state = streamStateUnknow
	stream.eventIn = make(chan uint32)
	stream.eventOut = make(chan uint32, 1)

	stream.fd.Reset()
	stream.Reset()
	go stream.mux()
	go stream.handleEventIn()
	go stream.handleEventOut()
}

func (stream *netStream) SetFD(f int, laddr, raddr syscall.Sockaddr) {
	stream.fd.Init(f, laddr, raddr)
}

func (stream *netStream) SetType(stype streamType) {
	stream.stype = stype
}

func (stream *netStream) SetState(iniState uint32) {
	if atomic.CompareAndSwapUint32(&stream.state, streamStateUnknow, iniState) == false {
		panic("unknow system error")
	}
}

func (stream *netStream) Reset() {
	stream.seqid++
	stream.netaddr = ((SeqidMASK & stream.seqid) << 20 | (OffsetMASK & stream.sindex))
}

func (stream * netStream) Close() {
	defer func() {
		recover()
	} ()

	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond * 300)
	//send the cache packet
	if len(stream.sendCh) > 0 {
		sendloop:
		for {
			state := atomic.LoadUint32(&stream.state)
			if state == streamStateUnknow {
				break
			}

			if state == streamStateOpened {
				err := stream.rawSend()
				if err != nil && err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
					break
				}
			}

			if len(stream.sendCh) == 0 {
				break
			}

			select {
			case <-ctx.Done():
				break sendloop
			default:
				break
			}
		}
	}

	stream.rawClose()
	stream.fd.Reset()
	close(stream.closeCh)
	close(stream.sendCh)
	close(stream.netpackCh)
	stream.closeCh = nil
	stream.sendCh = nil
	stream.netpackCh = nil
	stream.stype = streamTypeUnknow
}

func (stream *netStream) rawClose() {
	for {
		state := atomic.LoadUint32(&stream.state)
		if state == streamStateUnknow {
			return
		}
		
		if atomic.CompareAndSwapUint32(&stream.state, state, streamStateUnknow) {
			break
		}
	}
	
	stream.sendGroup.Wait()
	stream.recvGroup.Wait()
	pollEvent.Del(stream.fd.fd)
	stream.fd.fd = -1
	stream.sendOffset = 0
	stream.cached = 0
	stream.recvd = 0
	stream.packflag = false
	stream.resetReconnectTimer()
}

func (stream *netStream) resetReconnectTimer() {
	if stream.PreConnect() {
		stream.timerPeriod = time.Second
		stream.reconnectTimer.Reset(stream.timerPeriod)	
	}
}

func (stream *netStream) State() uint32 {
	return atomic.LoadUint32(&stream.state)
}

func (stream *netStream) Sysfd() int {
	return stream.fd.fd
}

func (stream *netStream) NetAddr() int32 {
	return stream.netaddr
}

func (stream *netStream) Opening() bool {
	return stream.state == streamStateOpening
}

func (stream *netStream) PreSend() bool {
	return stream.stype != streamTypeListen
}

func (stream *netStream) PreConnect() bool {
	return stream.stype == streamTypeConnect
}

func (stream *netStream) PreAccept() bool {
	return stream.stype == streamTypeListen
}

func (stream *netStream) PreClose() bool {
	return stream.fd.fd > 0
}

func (stream *netStream) Closed() bool {
	return stream.fd.fd < 0
}
