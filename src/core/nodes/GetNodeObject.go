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
	"fmt"
)

// GetNodeObject return an map from the node with nodeName
// This function DONT create a new Node inside the json if it dont exist
func GetNodeObject(nodeName string) (*map[string]interface{}, error) {

	// first, get the nodes from config
	nodes, err := config.GetJSONObject("nodes")
	if err != nil {
		return nil, err
	}

	// then get the node itselfe
	if node, ok := nodes[nodeName].(map[string]interface{}); ok {
		return &node, nil
	}

	return nil, fmt.Errorf("Node '%s' not found", nodeName)
}
