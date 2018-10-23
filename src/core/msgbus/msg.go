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
	"encoding/json"
	"fmt"
	"strconv"
)

type Msg struct {
	id        int     `json:"-"`
	pluginSrc *Plugin `json:"-"`

	NodeSource string `json:"s"`
	NodeTarget string `json:"t"`
	Group      string `json:"g"`
	Command    string `json:"c"`
	Payload    string `json:"v"`
}

type pluginCallback struct {
	group     string
	command   string
	onMessage onMessageFct
}

// callbacks
type onMessageFct func(*Msg, string, string, string)

func (curPlugin *Plugin) ListenForGroup(group string, onMessageFP onMessageFct) {

	// create new plugin and append it
	newCallback := pluginCallback{
		group:     group,
		command:   "",
		onMessage: onMessageFP,
	}

	curPlugin.callbacks = append(curPlugin.callbacks, newCallback)

	if group != "" {
		logging.Debug(fmt.Sprintf("PLUGIN %s", curPlugin.name), "Listen for group: "+group)
	} else {
		logging.Debug(fmt.Sprintf("PLUGIN %s", curPlugin.name), "Listen for all groups")
	}

	logging.Debug(fmt.Sprintf(
		"PLUGIN %s", curPlugin.name),
		fmt.Sprintf("Plugin %s - Callbacks %d", curPlugin.name, len(curPlugin.callbacks)),
	)
}

// publish an message to the BUS
// @param pluginIDSrc An pointer to an int where the plugin id is saved ( which was create before with Register() )
func (curPlugin *Plugin) Publish(nodeSource, nodeTarget, group, command, payload string) {

	newMessage := Msg{
		id:         messageListLastID,
		pluginSrc:  curPlugin,
		NodeSource: nodeSource,
		NodeTarget: nodeTarget,
		Group:      group,
		Command:    command,
		Payload:    payload,
	}

	logging.Debug("MSG "+strconv.Itoa(newMessage.id),
		fmt.Sprintf(
			"FROM %s TO %s/%s/%s",
			curPlugin.name, newMessage.NodeTarget, newMessage.Group, newMessage.Command,
		),
	)

	messageListLastID++

	messageList <- newMessage

	//close(messageList)

}

func (curPlugin *Plugin) PublishMsg(newMessage Msg) {

	newMessage.id = messageListLastID
	newMessage.pluginSrc = curPlugin

	logging.Debug("MSG "+strconv.Itoa(newMessage.id),
		fmt.Sprintf(
			"FROM %s TO %s/%s/%s",
			curPlugin.name, newMessage.NodeTarget, newMessage.Group, newMessage.Command,
		),
	)
	messageListLastID++

	messageList <- newMessage

	//close(messageList)

}

func worker(no int, messages <-chan Msg) {
	workerName := fmt.Sprintf("WORKER %d", no)
	logging.Debug(workerName, "Run")

	for curMessage := range messages {

		logging.Debug(workerName, fmt.Sprintf("MSG %d", curMessage.id))

		//logging.Debug(workerName, fmt.Sprintf("pluginList contains %d Plugins", len(pluginList)))
		for pluginIndex := range pluginList {
			curPlugin := pluginList[pluginIndex]

			logging.Debug(workerName, fmt.Sprintf("Plugin %s - Callbacks %d", curPlugin.name, len(curPlugin.callbacks)))

			// skip if the sender is also the reciever
			if curPlugin == curMessage.pluginSrc {

				logging.Debug(workerName,
					fmt.Sprintf(
						"[MSG %d] [PLUGIN '%s'] -> [PLUGIN '%s'] WE DONT SEND TO US",
						curMessage.id, curMessage.pluginSrc.name, curPlugin.name,
					),
				)

				continue
			}

			/*
				logging.Debug(workerName, fmt.Sprintf(
					"[MSG %d] [PLUGIN '%s'] contains %d callbacks",
					curMessage.id,
					curMessage.pluginSrc.name,
					len(curPlugin.callbacks)),
				)
			*/

			//for callbackIndex := 0; callbackIndex < len(curPlugin.callbacks); callbackIndex++ {
			for callbackIndex := range curPlugin.callbacks {
				curCallback := curPlugin.callbacks[callbackIndex]

				// check group
				if curCallback.group == "" || curCallback.group == curMessage.Group {

					logging.Debug(workerName,
						fmt.Sprintf(
							"[MSG %d] [PLUGIN '%s'] -> [PLUGIN '%s'] CALL",
							curMessage.id, curMessage.pluginSrc.name, curPlugin.name,
						),
					)

					curCallback.onMessage(&curMessage, curMessage.Group, curMessage.Command, curMessage.Payload)

				} else {
					logging.Debug(workerName,
						fmt.Sprintf(
							"[MSG %d] [PLUGIN '%s'] -> [PLUGIN '%s']  GROUP DONT MATCH",
							curMessage.id, curMessage.pluginSrc.name, curPlugin.name,
						),
					)
				}

			}

		}

	}

}

func (curMessage *Msg) ToJsonByteArray() ([]uint8, error) {

	b, err := json.Marshal(curMessage)
	if err != nil {
		fmt.Println("error:", err)
	}

	return b, nil
}

func (curMessage *Msg) ToJsonString() (string, error) {

	b, err := json.Marshal(curMessage)
	if err != nil {
		fmt.Println("error:", err)
	}

	return string(b), nil
}

func FromJsonString(jsonString string) (Msg, error) {

	var newMessage Msg

	err := json.Unmarshal([]byte(jsonString), &newMessage)
	if err != nil {
		fmt.Println("error: ", err)
		return newMessage, err
	}

	return newMessage, nil
}

func (curMessage *Msg) Answer(curPlugin *Plugin, command, payload string) error {

	curPlugin.Publish(
		curMessage.NodeTarget,
		curMessage.NodeSource,
		curMessage.Group,
		command,
		payload,
	)

	return nil
}
