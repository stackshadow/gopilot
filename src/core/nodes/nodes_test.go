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
	"testing"
)

func TestInit(t *testing.T) {

	// load the config
	config.ParseCmdLine()
	config.Init()

	config.ConfigPath = "/tmp"
	config.Read()

	// init the nodes
	Init()

	t.Run("Check if node dont exist", GetMissingNode)
	t.Run("Manipulate a Node", ManipulateANode)
	t.Run("Delete a Node", DeleteANode)

}

func GetMissingNode(t *testing.T) {

	testNode, err := GetNodeObject("testnode")
	if testNode != nil || err == nil {
		t.FailNow()
	}

}

func ManipulateANode(t *testing.T) {

	testNode, err := GetNodeObject("testnode")
	if testNode != nil || err == nil {
		t.FailNow()
	}

	// create a new node
	testNode = make(map[string]interface{})

	// set a new field inside the node
	testNode["foo"] = "bar"
	SaveNodeObject("testnode", testNode)

	// check if it was saved
	config.Read()
	testNode, err = GetNodeObject("testnode")
	if testNode["foo"] != "bar" {
		t.FailNow()
	}

}

func DeleteANode(t *testing.T) {
	Delete("testnode")

	config.Read()

	testNode, err := GetNodeObject("testnode")
	if testNode != nil || err == nil {
		t.FailNow()
	}
}
