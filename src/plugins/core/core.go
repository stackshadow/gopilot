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
	"core/msgbus"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var logging clog.Logger
var corePlugin msgbus.Plugin
var NodeName string
var ConfigPath string
var jsonConfig map[string]interface{}

func ParseCmdLine() {

	// nodename
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	flag.StringVar(&NodeName, "nodeName", hostname, "Set the name of this node ( normaly the hostname is used )")

	// config path
	flag.StringVar(&ConfigPath, "configPath", ".", "The base path")
}

func Init() {
	logging = clog.New("CORE")
	logging.Info("HOST", "MyNode: "+NodeName)

	corePlugin = msgbus.NewPlugin("Core")
	corePlugin.Register()
	corePlugin.ListenForGroup("co", onMessage)

	jsonConfig = make(map[string]interface{})
}

const NodeTypeUndefined int = 0 // do nothing with it
const NodeTypeServer int = 1    // serve an connection
const NodeTypeClient int = 2    // connect to an server as client
const NodeTypeIncoming int = 3  // incoming connection from another node

type nodesIterateFct func(map[string]interface{}, string, int, string, int) // name, type, host, port

func ConfigRead() {

	// Open our jsonFile
	jsonFile, err := os.Open(ConfigPath + "/core.json")
	if err != nil {
		logging.Info("CONFIG", err.Error())
		return
	}
	logging.Debug("CONFIG", "Successfully Opened '"+ConfigPath+"/core.json'")
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &jsonConfig)

	/*
		if jsonNodes, ok := jsonConfig["nodes"].(map[string]interface{}); ok {
			fmt.Println(jsonNodes)
		}
	*/
	/*
		for k, v := range jsonConfig {
			fmt.Println("k:", k, "v:", v)

			if v2, ok := v.(map[string]interface{}); ok {
				fmt.Println(v2)
			}
		}
		fmt.Println(f)
	*/
}

func IterateNodes(nodesIterateFctPt nodesIterateFct) {
	if jsonNodes, ok := jsonConfig["nodes"].(map[string]interface{}); ok {
		for nodeName, jsonNodeInterface := range jsonNodes {

			if jsonNode, ok := jsonNodeInterface.(map[string]interface{}); ok {

				nodeType := NodeTypeUndefined
				if jsonNode["type"] != nil {
					nodeType = int(jsonNode["type"].(float64))
				}

				nodeHost := "127.0.0.1"
				if jsonNode["host"] != nil {
					nodeHost = jsonNode["host"].(string)
				}

				nodePort := 4444
				if jsonNode["port"] != nil {
					nodePort = int(jsonNode["port"].(float64))
				}

				nodesIterateFctPt(jsonNode, nodeName, nodeType, nodeHost, nodePort)
			}
		}
	}
}

// This function return an map from the node with nodeName
// This function DONT create a new Node inside the json if it dont exist
func GetNodeObject(nodeName string) (*map[string]interface{}, error) {

	var jsonNode map[string]interface{}

	if jsonNodes, ok := jsonConfig["nodes"].(map[string]interface{}); ok {
		if jsonNodes[nodeName] != nil {
			jsonNode = jsonNodes[nodeName].(map[string]interface{})
			return &jsonNode, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Node '%s' not found", nodeName))
}

func GetNode(nodeName string) (int, string, int, error) {

	if jsonNodes, ok := jsonConfig["nodes"].(map[string]interface{}); ok {
		if jsonNodes[nodeName] != nil {

			jsonNode := jsonNodes[nodeName].(map[string]interface{})

			nodeType := NodeTypeUndefined
			if jsonNode["type"] != nil {
				nodeType = int(jsonNode["type"].(float64))
			}

			nodeHost := "127.0.0.1"
			if jsonNode["host"] != nil {
				nodeHost = jsonNode["host"].(string)
			}

			nodePort := 4444
			if jsonNode["port"] != nil {
				nodePort = int(jsonNode["port"].(float64))
			}

			return nodeType, nodeHost, nodePort, nil
		}
	}

	return NodeTypeUndefined, "127.0.0.1", 4444, errors.New("Node not found")
}

func SetNode(nodeName string, nodeType int, host string, port int) {

	var jsonNodes map[string]interface{}
	var jsonNode map[string]interface{}

	if jsonConfig["nodes"] != nil {
		jsonNodes = jsonConfig["nodes"].(map[string]interface{})

		if jsonNodes[nodeName] != nil {
			jsonNode = jsonNodes[nodeName].(map[string]interface{})
		} else {
			jsonNode = make(map[string]interface{})
			jsonNodes[nodeName] = jsonNode
		}
	} else {
		jsonNodes = make(map[string]interface{})
		jsonConfig["nodes"] = jsonNodes
		jsonNode = make(map[string]interface{})
		jsonNodes[nodeName] = jsonNode
	}

	jsonNode["type"] = nodeType
	jsonNode["host"] = host
	jsonNode["port"] = port

	fmt.Println("jsonConfig:", jsonConfig)
}

func DeleteNode(nodeName string) {

	if jsonConfig["nodes"] == nil {
		return
	}

	jsonNodes := jsonConfig["nodes"].(map[string]interface{})
	delete(jsonNodes, nodeName)

}

func GetJsonObject(name string) (*map[string]interface{}, error) {

	if jsonObject, ok := jsonConfig[name].(map[string]interface{}); ok {
		return &jsonObject, nil
	}

	return nil, errors.New(fmt.Sprintf("No Object found with name '%s'", name))
}

func NewJsonObject(name string) (*map[string]interface{}, error) {
	var jsonNode = make(map[string]interface{})

	jsonConfig[name] = jsonNode

	return &jsonNode, nil
}

func SetJsonObject(name string, jsonNode map[string]interface{}) {
	jsonConfig[name] = jsonNode
}

func ConfigSave() {
	byteValue, _ := json.MarshalIndent(jsonConfig, "", "    ")
	err := ioutil.WriteFile(ConfigPath+"/core.json", byteValue, 0644)
	if err != nil {
		logging.Error("CONFIG", err.Error())
		os.Exit(-1)
	}
}

func onMessage(message *msgbus.Msg, group, command, payload string) {

	// from here: all nodes can request these
	if command == "nodeNameGet" {
		message.Answer(&corePlugin, "nodeName", NodeName)
		return
	}

	// from here: only commands for THIS node
	if message.NodeTarget != NodeName {
		return
	}

	if command == "getNodes" {

		IterateNodes(func(jsonNode map[string]interface{}, nodeName string, nodeType int, host string, port int) {

			var requested bool
			var accepted bool

			if jsonNode["peerCertSignatureReq"] != nil {
				requested = true
				accepted = false
			}
			if jsonNode["peerCertSignature"] != nil {
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

		if jsonNodes, ok := jsonConfig["nodes"].(map[string]interface{}); ok {

			for nodeName, jsonNodeInterface := range jsonNewNodes {
				if jsonNode, ok := jsonNodeInterface.(map[string]interface{}); ok {
					jsonNodes[nodeName] = jsonNode
					message.Answer(&corePlugin, "nodeSaveOk", nodeName)
				}
			}

		}

		ConfigSave()

		IterateNodes(func(jsonNode map[string]interface{}, nodeName string, nodeType int, host string, port int) {

			var requested bool
			var accepted bool

			if jsonNode["peerCertSignatureReq"] == nil {
				requested = true
				accepted = false
			}
			if jsonNode["peerCertSignature"] == nil {
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

	}

	if command == "nodeDelete" {
		DeleteNode(payload)
		message.Answer(&corePlugin, "nodeDeleteOk", payload)
		return
	}

	if command == "ping" {
		message.Answer(&corePlugin, "pong", "")
		return
	}
}
