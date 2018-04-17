// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package netgo

import (
	"math"
)
const (
	//INVALIDADDR invalid socket addr offset
	INVALIDADDR = math.MaxInt32
	maxSocketNumber = 500000
	maxListenBacklog = 10000
)
