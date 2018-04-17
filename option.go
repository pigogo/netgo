// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

type SocketOption struct {
	Linger int
	Nodelay bool
	Keepalive bool
	KeepIdle int
	ReusedAddr bool
	MaxListenBacklog int
}