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

package nodes

import (
	"core/config"
	"github.com/mitchellh/mapstructure"
)

// IterateFct is a callback-function for IterateNodes()
type IterateFct func(JSONNodeType, string, int, string, int) // name, type, host, port

// IterateNodes call for every node in the config the NodesIterateFct
func IterateNodes(nodesIterateFctPt IterateFct) {

	// first, get the nodes from config
	nodes, err := config.GetJSONObject("nodes")
	if err != nil {
		return
	}

	for nodeName, jsonNodeInterface := range nodes {

		// convert it to struct
		var jsonNode JSONNodeType
		err = mapstructure.Decode(jsonNodeInterface, jsonNode)
		if err != nil {
			continue
		}

		var nodeType = int(jsonNode.Type)

		nodeHost := "127.0.0.1"
		if jsonNode.Host != "" {
			nodeHost = jsonNode.Host
		}

		var nodePort = 4444
		if jsonNode.Port != 0 {
			nodePort = int(jsonNode.Port)
		}

		nodesIterateFctPt(jsonNode, nodeName, nodeType, nodeHost, nodePort)
	}

}
