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
	"core/tools"
	"fmt"
)

// structs
type Plugin struct {
	id   string
	name string
}

var logging clog.Logger

var pluginList []*Plugin

func PluginsInit() {

	logging = clog.New("BUS")

	pluginList = make([]*Plugin, 0)

}

func NewPlugin(pluginName string) Plugin {

	newPlugin := Plugin{
		id:   pluginName + "-" + tools.RandomString(4),
		name: pluginName,
	}

	return newPlugin
}

func (curPlugin *Plugin) Register() {
	pluginList = append(pluginList, curPlugin)

	logging.Debug(fmt.Sprintf("PLUGIN %s", curPlugin.id),
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

func (curPlugin *Plugin) ListenForGroup(group string, onMessageFP onMessageFct) {
	ListenForGroup(curPlugin.id, group, onMessageFP)
}

// publish an message to the BUS
// @param pluginIDSrc An pointer to an int where the plugin id is saved ( which was create before with Register() )
func (curPlugin *Plugin) Publish(nodeSource, nodeTarget, group, command, payload string) {
	Publish(curPlugin.id, nodeSource, nodeTarget, group, command, payload)
}

func (curPlugin *Plugin) PublishMsg(newMessage Msg) {
	PublishMsg(curPlugin.id, newMessage)
}
