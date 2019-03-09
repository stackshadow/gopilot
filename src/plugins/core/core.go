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

package core

import (
	"core/clog"
	"core/config"
	"core/msgbus"
	"core/nodes"
	"encoding/json"

	"fmt"
)

var logging clog.Logger
var corePlugin msgbus.Plugin

func Init() {
	logging = clog.New("CORE")
	logging.Info("HOST", "MyNode: "+config.NodeName)

	corePlugin = msgbus.NewPlugin("Core")
	corePlugin.Register()
	corePlugin.ListenForGroup("co", onMessage)
}

func onMessage(message *msgbus.Msg, group, command, payload string) {

	// from here: all nodes can request these
	if command == "nodeNameGet" {
		message.Answer(&corePlugin, "nodeName", config.NodeName)
		return
	}

	// from here: only commands for THIS node
	if message.NodeTarget != config.NodeName {
		return
	}

	if command == "getNodes" {

		nodes.IterateNodes(func(jsonNode nodes.JSONNodeType, nodeName string, nodeType int, host string, port int) {

			var requested bool
			var accepted bool

			if jsonNode.PeerCertSignatureReq != "" {
				requested = true
				accepted = false
			}
			if jsonNode.PeerCertSignature != "" {
				requested = false
				accepted = true
			}

			message.Answer(&corePlugin, "node",
				fmt.Sprintf(
					"{\"%s\":{ \"host\":\"%s\", \"port\":%d, \"type\":%d, \"req\": %t, \"acc\": %t } }",
					nodeName, host, port, nodeType, requested, accepted,
				),
			)

		})

		message.Answer(&corePlugin, "nodeEnd", "")

		return
		/*
			if jsonNodes, ok := jsonConfig["nodes"].(map[string]interface{}); ok {
				b, _ := json.Marshal(jsonNodes)
				message.Answer(&corePlugin, "nodes", string(b))
				return
			}
		*/
	}

	if command == "nodeSave" {

		var jsonNewNodes map[string]interface{}
		err := json.Unmarshal([]byte(payload), &jsonNewNodes)
		if err != nil {
			message.Answer(&corePlugin, "error", err.Error())
			return
		}

	}

	if command == "nodeDelete" {
		nodes.Delete(payload)
		message.Answer(&corePlugin, "nodeDeleteOk", payload)
		return
	}

	if command == "ping" {
		message.Answer(&corePlugin, "pong", "")
		return
	}
}
