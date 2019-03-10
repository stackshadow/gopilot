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

package msgbus

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

// Msg represent a single message inside the bus
type Msg struct {
	id            int
	pluginNameSrc string

	NodeSource string `json:"s"`
	NodeTarget string `json:"t"`
	Group      string `json:"g"`
	Command    string `json:"c"`
	Payload    string `json:"v"`
}

type msgListener struct {
	pluginName string
	target     string // can be ""
	group      string // can be ""
	command    string // can be ""
	onMessage  onMessageFct
}

var messageList chan Msg
var messageListLastID int
var messageListeners []msgListener
var messageListenersMutex sync.Mutex

// callbacks
type onMessageFct func(*Msg, string /* group */, string /*command*/, string /*payload*/) // For example: onMessage(message *msgbus.Msg, group, command, payload string)

func MsgBusInit() {

	messageList = make(chan Msg, 10)
	messageListLastID = 0

	for w := 1; w <= 1; w++ {
		logging.Debug("WORKER "+strconv.Itoa(w), "Start")
		go worker2(w, messageList)
	}
}

func ListenForGroup(pluginName string, group string, onMessageFP onMessageFct) {

	// create new plugin and append it
	newListener := msgListener{
		pluginName: pluginName,
		target:     "",
		group:      group,
		command:    "",
		onMessage:  onMessageFP,
	}

	messageListenersMutex.Lock()
	messageListeners = append(messageListeners, newListener)
	messageListenersMutex.Unlock()

	if group != "" {
		logging.Debug(fmt.Sprintf("PLUGIN %s", newListener.pluginName), "Listen for group: "+group)
	} else {
		logging.Debug(fmt.Sprintf("PLUGIN %s", newListener.pluginName), "Listen for all groups")
	}

}

func ListenNoMorePlugin(pluginName string) {

	messageListenersMutex.Lock()
	var NewMessageListeners []msgListener

	for listenerIndex, curListener := range messageListeners {
		//curListener := messageListeners[listenerIndex]

		if curListener.pluginName == pluginName {
			logging.Debug(
				fmt.Sprintf("PLUGIN %s", curListener.pluginName),
				fmt.Sprintf("Remove Listener in index '%d' for target: '%s' group: '%s'", listenerIndex, curListener.target, curListener.group),
			)
		} else {
			NewMessageListeners = append(NewMessageListeners, curListener)
		}

	}

	messageListeners = NewMessageListeners
	messageListenersMutex.Unlock()

}

func ListenersCount() int {
	return len(messageListeners)
}

func Publish(pluginName string, nodeSource, nodeTarget, group, command, payload string) {

	newMessage := Msg{
		id:            messageListLastID,
		pluginNameSrc: pluginName,
		NodeSource:    nodeSource,
		NodeTarget:    nodeTarget,
		Group:         group,
		Command:       command,
		Payload:       payload,
	}

	logging.Debug("MSG "+strconv.Itoa(newMessage.id),
		fmt.Sprintf(
			"FROM %s TO %s/%s/%s",
			newMessage.pluginNameSrc, newMessage.NodeTarget, newMessage.Group, newMessage.Command,
		),
	)

	messageListLastID++

	messageList <- newMessage

	//close(messageList)

}

func PublishMsg(pluginName string, newMessage Msg) {

	newMessage.id = messageListLastID
	newMessage.pluginNameSrc = pluginName

	logging.Debug("MSG "+strconv.Itoa(newMessage.id),
		fmt.Sprintf(
			"FROM %s TO %s/%s/%s",
			newMessage.pluginNameSrc, newMessage.NodeTarget, newMessage.Group, newMessage.Command,
		),
	)
	messageListLastID++

	messageList <- newMessage

	//close(messageList)

}

func worker2(no int, messages <-chan Msg) {
	workerName := fmt.Sprintf("WORKER %d", no)
	logging.Debug(workerName, "Run")

	for curMessage := range messages {

		logging.Debug(workerName, fmt.Sprintf("MSG %d", curMessage.id))

		//logging.Debug(workerName, fmt.Sprintf("pluginList contains %d Plugins", len(pluginList)))

		messageListenersMutex.Lock()
		for listenerIndex := range messageListeners {
			curListener := messageListeners[listenerIndex]

			// skip if the sender is also the reciever
			if curListener.pluginName == curMessage.pluginNameSrc {

				logging.Debug(workerName,
					fmt.Sprintf(
						"[MSG %d] [PLUGIN '%s'] -> [PLUGIN '%s'] WE DONT SEND TO US",
						curMessage.id, curMessage.pluginNameSrc, curListener.pluginName,
					),
				)

				continue
			}

			// check group
			if curListener.group == "" || curListener.group == curMessage.Group {

				logging.Debug(workerName,
					fmt.Sprintf(
						"[MSG %d] [PLUGIN '%s'] -> [PLUGIN '%s'] CALL",
						curMessage.id, curMessage.pluginNameSrc, curListener.pluginName,
					),
				)

				curListener.onMessage(&curMessage, curMessage.Group, curMessage.Command, curMessage.Payload)

			} else {
				logging.Debug(workerName,
					fmt.Sprintf(
						"[MSG %d] [PLUGIN '%s'] -> [PLUGIN '%s']  GROUP DONT MATCH",
						curMessage.id, curMessage.pluginNameSrc, curListener.pluginName,
					),
				)
			}

		}
		messageListenersMutex.Unlock()

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

	Publish(
		curPlugin.id,
		curMessage.NodeTarget,
		curMessage.NodeSource,
		curMessage.Group,
		command,
		payload,
	)

	return nil
}
