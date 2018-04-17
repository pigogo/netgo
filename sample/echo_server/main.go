
// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package main

import (
	"flag"
	"runtime"
	"github.com/pigogo/netgo"
	"golibs/x/funclib"
	"net/http"
	"fmt"
	_ "net/http/pprof"
)

var (
	netgoApi netgo.INetgo
)

type packet struct {
	body []byte
	addr int32
}

func onPacket(body []byte, naddr int32) {
	netgoApi.Send(naddr, body)
	return 
}

func initPprof() {
	funclib.InitCpuProfile("")
	go http.ListenAndServe("0.0.0.0:9988", nil)
}

func init() {
	initPprof()
}

func main() {
	var listenip string
	var port int
	flag.StringVar(&listenip, "ip", "localhost", "connect ip")
	flag.IntVar(&port, "port", 8899, "connect port")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())
	option := netgo.SocketOption{
		Keepalive: true,
		KeepIdle: 300,
		ReusedAddr: true,
		Nodelay: true,
		Linger: 1,
	}

	netgoApi = netgo.NewNetgo()
	netgoApi.Init(onPacket, option)
	if err := netgoApi.Listen(fmt.Sprintf("%v:%v", listenip,port)); err != nil {
		panic(fmt.Sprintln("listen error:", err))
	}
	netgoApi.Serve()
}