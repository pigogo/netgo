
// +build linux

// Copyright (C) 2018 Kun Zhong All rights reserved.
// Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

package main

import (
	"os"
	"bufio"
	"runtime"
	"github.com/pigogo/netgo/netgo"
	"fmt"
	"flag"
	"net/http"
	"golibs/x/funclib"
	_ "net/http/pprof"
	"runtime/pprof"
)

var (
	netgoApi netgo.INetgo
	netAddr int32
)

func initPprof() {
	funclib.InitCpuProfile("")
	go http.ListenAndServe("0.0.0.0:9999", nil)
}

func init() {
	initPprof()
}

func onPacket(body []byte, naddr int32) {
	fmt.Printf("Resp:%v\n", string(body))
	go input("")
}

func input(str string) {
	if len(str) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Send:")
		instring, _ := reader.ReadString('\n')
		netgoApi.Send(netAddr, []byte(instring))
	} else {
		fmt.Printf("Send:%v\n", str)
		netgoApi.Send(netAddr, []byte(str))
	}
}

func main() {
	var connip string
	var port int
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&connip, "ip", "localhost", "connect ip")
	flag.IntVar(&port, "port", 8899, "connect port")
	flag.Parse()

	option := netgo.SocketOption{
		Keepalive: true,
		KeepIdle: 300,
		ReusedAddr: true,
		Nodelay: true,
		Linger: 1,
	}

	var (
		err error
	)

	netgoApi = netgo.NewNetgo()
	if err = netgoApi.Init(onPacket, option); err != nil {
		panic(err)
	}
	

	if netAddr, err = netgoApi.Connect(fmt.Sprintf("%v:%v", connip, port)); err != nil {
		panic(fmt.Sprintln("connect error:", err))
		return
	}

	go input("hello world!")
	netgoApi.Serve()
	pprof.StopCPUProfile()
}