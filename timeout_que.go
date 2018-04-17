// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"container/list"
	"time"
)

type listNode struct {
	tm time.Time
	naddr int32
}

type toutQueue struct {
	timeoutList	list.List
	listEle map[int32]*list.Element
}

func newToutQue() *toutQueue {
	return &toutQueue {
		listEle: make(map[int32]*list.Element),
	}
}

func (tq *toutQueue) Peek() *listNode {
	if tq.timeoutList.Len() == 0 {
		return nil
	}

	return tq.timeoutList.Front().Value.(*listNode)
}

func (tq *toutQueue) Pop() {
	if tq.timeoutList.Len() == 0 {
		return
	}

	delete(tq.listEle, tq.timeoutList.Front().Value.(*listNode).naddr)
	tq.timeoutList.Remove(tq.timeoutList.Front())
}

func (tq *toutQueue) Remove(naddr int32) {
	if ele, ok := tq.listEle[naddr]; ok {
		delete(tq.listEle, naddr)
		tq.timeoutList.Remove(ele)
	}
}

func (tq *toutQueue) Push(naddr int32) {
	ele, ok := tq.listEle[naddr]
	if ok {
		ele.Value.(*listNode).tm = time.Now()
		return
	}

	ele = tq.timeoutList.PushBack(&listNode{
		tm: time.Now(),
		naddr: naddr,
	})
	tq.listEle[naddr] = ele
}