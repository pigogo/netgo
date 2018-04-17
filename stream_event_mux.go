
// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"syscall"
	"sync/atomic"
)

func (stream *netStream) OnEvent(events uint32) {
	for {
		if events & syscall.EPOLLERR != 0 {
			stream.rawClose()
			break
		}

		if events & syscall.EPOLLOUT != 0 {
			select {
			case stream.eventOut <- events:
				break
			default:
				break
			}
		}
		
		if events & syscall.EPOLLIN != 0 {
			if stream.PreAccept() {
				stream.netDriver.accept(stream.netaddr)
				break
			}

			select {
			case stream.eventIn<-events:
				break
			default:
				break
			}
		}

		break
	}
}

func (stream *netStream) handleEventIn() {
loop:
	for {
		select {
		case <-stream.eventIn:
			if stream.State() == streamStateUnknow {
				continue
			}

			if stream.packflag {
				stream.recvPack()
			} else {
				stream.recv()
			}
			break
		case <-stream.closeCh:
			break loop
		}
	}
}

func (stream *netStream) handleEventOut() {
loop:
	for {
next:
		select {
		case <-stream.eventOut:
			state := stream.State()
			if state == streamStateOpened {
				stream.send()
			} else if state == streamStateOpening {
				for {
					if atomic.CompareAndSwapUint32(&stream.state, state, streamStateOpened)	{
						break
					}
					state = stream.State()
					if state == streamStateUnknow {
						break next
					}
				}
				stream.send()
			}
			
			break
		case <-stream.closeCh:
			break loop
		}
	}
}

func (stream *netStream) mux() {
loop:
	for {
		select {
		case <-stream.closeCh:
			break loop
		case <-stream.reconnectTimer.C:
			stream.onReconnectTimer()
			break
		}
	}
	stream.reconnectTimer.Stop()
}

func (stream *netStream) onReconnectTimer() {
	if stream.fd.fd >= 0 {
		stream.reconnectTimer.Reset(infinitTime)
		return
	}

	err := stream.netDriver.rawConnect(stream, stream.fd.remoteAddr)
	if err != nil{
		return
	}

	stream.timerPeriod = infinitTime
	stream.reconnectTimer.Reset(stream.timerPeriod)
}