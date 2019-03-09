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
	"github.com/mitchellh/mapstructure"
)

// GetData return the node-type, the node-hostname and the node-port
func GetData(nodeName string) (int, string, int, error) {

	// get the single node
	nodeI, err := GetNodeObject(nodeName)
	if err != nil {
		return 0, "", 0, err
	}

	// convert it to struct
	var node JSONNodeType
	err = mapstructure.Decode(nodeI, node)
	if err != nil {
		return 0, "", 0, err
	}

	if node.Type == 0 {
		node.Type = float64(NodeTypeUndefined)
	}

	if node.Host == "" {
		node.Host = "127.0.0.1"
	}

	if node.Port == 0 {
		node.Port = float64(4444)
	}

	return int(node.Type), node.Host, int(node.Port), nil

}
