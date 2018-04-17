// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

type INetgo interface{
	Init(func(body []byte, naddr int32), SocketOption) error
	Send(naddr int32, buf[]byte) error
	Connect(addr string) (naddr int32, err error)
	Close(addr int32) error
	Listen(laddr string) (err error)
	Serve() error
}

func NewNetgo() INetgo {
	return &netio {}
}