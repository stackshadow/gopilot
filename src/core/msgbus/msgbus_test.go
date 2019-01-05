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

import "testing"
import "fmt"
import "time"
import "core/clog"
import "os"

func TestInit(t *testing.T) {

	MsgBusInit()
	Init()
	clog.EnableDebug()
}

func TestRegisterDeregister(t *testing.T) {

	// this test, register an listener, which should ONLY be called once,
	// because all listeners will be deleted after
	ListenForGroup("PLUGIN A", "", onlyOnSingleMessage)
	ListenForGroup("PLUGIN B", "", onlyOnSingleMessage)
	ListenForGroup("PLUGIN C", "", onlyOnSingleMessage)
	ListenForGroup("PLUGIN D", "", onlyOnSingleMessage)
	ListenForGroup("PLUGIN D", "", onlyOnSingleMessage)

	if ListenersCount() != 5 {
		t.Error("There should be 5 listeners...")
		t.FailNow()
		return
	}

	// remove listener
	ListenNoMorePlugin("PLUGIN A")
	ListenNoMorePlugin("PLUGIN C")
	ListenNoMorePlugin("PLUGIN D")

	if ListenersCount() != 1 {
		t.Error("There should be 1 listeners...")
		t.FailNow()
		return
	}

	// send a message ( this now should be fired only once )
	Publish("DUMMY", "sourceNode", "targetNode", "group", "command", "payload")

	time.Sleep(time.Second * 2)
}

var onlyOnSingleMessageCount int = 0

func onlyOnSingleMessage(message *Msg, group, command, payload string) {
	if onlyOnSingleMessageCount > 0 {
		os.Exit(-1)
	}
	onlyOnSingleMessageCount++
}

func aTestJsonMessage(t *testing.T) {

	newMessage := Msg{
		id:        1,
		pluginSrc: nil,

		NodeSource: "source",
		NodeTarget: "target",
		Group:      "group",
		Command:    "command",
		Payload:    "payload",
	}

	// convert it to json
	jsonString, err := newMessage.ToJsonString()
	if err != nil {
		t.Error("Can not convert to string")
		t.FailNow()
		return
	}

	// check it
	if jsonString != "{\"s\":\"source\",\"t\":\"target\",\"g\":\"group\",\"c\":\"command\",\"v\":\"payload\"}" {
		t.Error("Json-String is wrong")
		t.FailNow()
		return
	}

	// convert to msg
	newMessage, _ = FromJsonString("{\"s\":\"src\",\"t\":\"trg\",\"g\":\"grp\",\"c\":\"cmd\",\"v\":\"pld\"}")
	if newMessage.NodeSource != "src" ||
		newMessage.NodeTarget != "trg" ||
		newMessage.Group != "grp" ||
		newMessage.Command != "cmd" ||
		newMessage.Payload != "pld" {
		t.Error("Json can not converted to message")
		t.FailNow()
		return
	}

}

func aTestPluginListener(t *testing.T) {

	var firstPluginID, secondPluginID, thirdPluginID Plugin

	firstPluginID = NewPlugin("First Plugin")
	secondPluginID = NewPlugin("Second Plugin")
	thirdPluginID = NewPlugin("Third Plugin")

	firstPluginID.Register()
	secondPluginID.Register()
	thirdPluginID.Register()

	firstPluginID.ListenForGroup("groupa", testOnMessage)
	secondPluginID.ListenForGroup("groupa", testOnMessage)
	thirdPluginID.ListenForGroup("groupa", testOnMessage)

	firstPluginID.Publish("me", "other", "groupa", "ping", "nopayload")
	firstPluginID.Publish("me", "other", "groupa", "ping", "nopayload")
	time.Sleep(time.Second * 2)
}

func testOnMessage(message *Msg, group, command, payload string) {
	fmt.Println("GROUP: ", group, " CMD: ", command, " PAYLOAD: ", payload)
}

func onMultipleMessage(message *Msg, group, command, payload string) {
	if group != "group" && group != "tst" {
		os.Exit(-1)
	}
}
func onSingleMessage(message *Msg, group, command, payload string) {
	if group != "tst" {
		os.Exit(-1)
	}
}
func onNeverMessage(message *Msg, group, command, payload string) {
	fmt.Println("GROUP: ", group, " CMD: ", command, " PAYLOAD: ", payload)
}
