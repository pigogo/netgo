// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"errors"
)

var (
	ErrOutcacheFull		= errors.New("out cache full")
	ErrInputcacheFull	= errors.New("input cache full")
	ErrUnknowSysError	= errors.New("unknow system error")
	ErrListenError		= errors.New("listen socket error")
	ErrNoSocketAvailable = errors.New("no socket available")
	ErrInvalidListenAddr = errors.New("invalid listen addr")
	ErrInvalidConnectAddr = errors.New("invalid connect addr")
	ErrInvalidSocketAddr = errors.New("invalid socket address")
	ErrInvalidPacketRecv = errors.New("invalid packet recv")
	ErrSocketBeClosed = errors.New("socket be closed")
)