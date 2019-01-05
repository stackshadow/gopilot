/*
Copyright (C) 2018 by Martin Langlotz aka stackshadow

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
	"core/msgbus"
	"flag"
	"fmt"
	"plugins/core"
	"plugins/ctls"
	"plugins/health"
	"plugins/ldap"
	"plugins/webclient"
	"time"
)

type Job struct {
	id       int
	randomno int
}

func main() {

	clog.ParseCmdLine()
	core.ParseCmdLine()
	webclient.ParseCmdLine()
	ctls.ParseCmdLine()
	ldapclient.ParseCmdLine()
	flag.Parse()

	clog.Init()
	msgbus.MsgBusInit()
	msgbus.PluginsInit()

	core.Init()
	core.ConfigRead()

	// get my node
	var host string
	var nodeType, port int
	var err error
	nodeType, host, port, err = core.GetNode(core.NodeName)
	if err != nil {
		core.SetNode(core.NodeName, nodeType, host, port)
		core.ConfigSave()
	}

	ctls.Init()
	webclient.Init()
	health.Init()
	ldapclient.Init()

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
