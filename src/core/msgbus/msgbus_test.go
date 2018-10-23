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

func TestPluginList(t *testing.T) {

	Init()
	clog.EnableDebug()
	/*
		var firstPlugin Plugin
		var secondPlugin Plugin
		thirdPluginID := 1

		firstPlugin = NewPlugin("First Plugin")
		firstPlugin = NewPlugin("Second Plugin")
	*/
	time.Sleep(time.Second * 2)
}

func TestJsonMessage(t *testing.T) {

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

func TestPluginListener(t *testing.T) {

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
