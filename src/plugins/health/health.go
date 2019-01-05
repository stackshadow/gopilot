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

package health

import (
	"core/clog"
	"core/msgbus"
	"encoding/json"
	"plugins/core"
)

type pluginHealth struct {
	logging clog.Logger
	plugin  msgbus.Plugin

	healthSourceName string  // The last one who set the health
	health           float32 // The health in percent 0=broken 100=gooooooood

}

func Init() pluginHealth {

	var newPluginHealth pluginHealth

	newPluginHealth.logging = clog.New("HEALTH")
	newPluginHealth.plugin = msgbus.NewPlugin("HEALTH")
	newPluginHealth.plugin.Register()
	newPluginHealth.plugin.ListenForGroup("hlt", newPluginHealth.onMessage)

	// web-server
	/*
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Welcome to my website!")
		})
	*/

	return newPluginHealth
}

func (curHealth *pluginHealth) onMessage(message *msgbus.Msg, group, command, payload string) {

	// from here: only commands for THIS node
	if message.NodeTarget != core.NodeName {
		return
	}

	if command == "set" {

		// get new node
		type msgHealthSet struct {
			Source string  `json:"source"`
			Value  float32 `json:"value"`
		}

		// parse json
		var jsonHealth msgHealthSet
		err := json.Unmarshal([]byte(payload), &jsonHealth)
		if err != nil {
			message.Answer(&curHealth.plugin, "error", err.Error())
			return
		}

		if jsonHealth.Value < curHealth.health {

			// get health source description
			curHealth.healthSourceName = jsonHealth.Source

			// get health value
			curHealth.health = jsonHealth.Value

			return
		}
	}

}
