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

// SetData will set the nodeType, hostname and port of an node ( will be created if not exist
func SetData(nodeName string, nodeType int, host string, port int) error {

	// get the single node
	var node JSONNodeType
	nodeI, err := GetNodeObject(nodeName)
	if err == nil {
		err = mapstructure.Decode(nodeI, node)
		if err != nil {
			return err
		}
	}

	node.Type = float64(nodeType)
	node.Host = host
	node.Port = float64(port)

	// save it back
	// first, get the nodes from config
	nodes, err := config.GetJSONObject("nodes")
	if err != nil {
		return err
	}

	// overwrite node
	nodes[nodeName] = node

	// save it back
	config.SetJSONObject("nodes", nodes)

	return nil
}
