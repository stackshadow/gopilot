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

package msgbus

import (
	"core/clog"

	"fmt"
	"strconv"
)

// structs
type Plugin struct {
	id        int
	name      string
	callbacks []pluginCallback
}

var logging clog.Logger

var pluginList []*Plugin

var messageList chan Msg
var messageListLastID int

// vars
var lastPluginID int

func Init() {

	logging = clog.New("BUS")

	messageList = make(chan Msg, 10)
	pluginList = make([]*Plugin, 0)

	for w := 1; w <= 1; w++ {
		logging.Debug("WORKER "+strconv.Itoa(w), "Start")
		go worker(w, messageList)
	}
}

func NewPlugin(pluginName string) Plugin {

	newPlugin := Plugin{
		id:   lastPluginID,
		name: pluginName,
	}

	lastPluginID++

	return newPlugin
}

func (curPlugin *Plugin) Register() {
	pluginList = append(pluginList, curPlugin)

	logging.Debug(fmt.Sprintf("PLUGIN %p", curPlugin),
		fmt.Sprintf(
			"New Plugin '%s' registered. We now have %d plugins.",
			curPlugin.name, len(pluginList),
		),
	)
}

func (curPlugin *Plugin) DeRegister() {

	for pluginIndex := range pluginList {
		actPlugin := pluginList[pluginIndex]

		if actPlugin == curPlugin {
			logging.Info(fmt.Sprintf("PLUGIN '%s'", curPlugin.name),
				fmt.Sprintf(
					"Deregister Plugin with name '%s' on index '%d'",
					actPlugin.name, pluginIndex,
				),
			)

			pluginList = append(pluginList[:pluginIndex], pluginList[pluginIndex+1:]...)
		}

	}

	for pluginIndex := range pluginList {
		actPlugin := pluginList[pluginIndex]
		logging.Debug(fmt.Sprintf("PLUGIN '%s'", actPlugin.name),
			"",
		)
	}
}
