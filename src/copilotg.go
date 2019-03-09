/*
Copyright (C) 2019 by Martin Langlotz aka stackshadow

This file is part of gopilot, an rewrite of the copilot-project in go

gopilot is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, version 3 of this License

gopilot is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gopilot.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"core/clog"
	"core/config"
	"core/msgbus"
	"core/nodes"
	"flag"
	"fmt"
	"plugins/core"
	"plugins/ctls"
	"plugins/health"
	"plugins/ldap"
	"plugins/nft"
	"plugins/webclient"
	"runtime"
	"time"
)

type Job struct {
	id       int
	randomno int
}

func main() {

	fmt.Printf("Git Version %s from %s\n", plugincore.Gitversion, plugincore.Gitdate)

	go printMemUsage()

	// ########################## Command line parse ##########################
	// core stuff
	clog.ParseCmdLine()
	config.ParseCmdLine()

	// plugins
	pluginwebclient.ParseCmdLine()
	pluginctls.ParseCmdLine()
	pluginldap.ParseCmdLine()
	pluginnft.ParseCmdLine()
	flag.Parse()

	// ########################## Init ##########################
	// core
	clog.Init()

	config.Init()
	config.Read()

	nodes.Init()

	msgbus.MsgBusInit()
	msgbus.PluginsInit()

	// get my node
	var host string
	var nodeType, port int
	var err error
	nodeType, host, port, err = nodes.GetData(config.NodeName)
	if err != nil {
		nodes.SaveData(config.NodeName, nodeType, host, port)
	}

	plugincore.Init()
	pluginctls.Init()
	pluginwebclient.Init()
	pluginhealth.Init()
	pluginldap.Init()
	pluginnft.Init()

	for {
		time.Sleep(time.Second)
	}
}

func testOnMessage(group, command, payload string) {
	fmt.Println("GROUP: ", group, " CMD: ", command, " PAYLOAD: ", payload)
}

func testNodeIter(nodeName string, nodeType int, host string, port int) {
	fmt.Println("nodeName:", nodeName, "host:", host, "port:", port)
}

func bToKb(b uint64) uint64 {
	return b / 1024
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func printMemUsage() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		// For info on each, see: https://golang.org/pkg/runtime/#MemStats
		fmt.Printf("Alloc = %v KiB", bToKb(m.Alloc))
		fmt.Printf("\tTotalAlloc = %v KiB", bToKb(m.TotalAlloc))
		fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
		fmt.Printf("\tNumGC = %v\n", m.NumGC)
		time.Sleep(time.Second * 60)
	}
}
