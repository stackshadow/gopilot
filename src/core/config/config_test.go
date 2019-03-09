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
package config

import "testing"

func TestInit(t *testing.T) {
	ParseCmdLine()
	Init()

	ConfigPath = "/tmp"
	Read()

	t.Run("Test non existing config", GetNonExistingConfig)
	t.Run("Test creation of new config", CreateNewConfigEntry)
}

func GetNonExistingConfig(t *testing.T) {

	_, err := GetJSONObject("dontexist")
	if err == nil {
		t.FailNow()
	}
}

func CreateNewConfigEntry(t *testing.T) {

	jsonConfig := make(map[string]interface{})

	jsonConfig["test"] = "okay"

	SetJSONObject("newone", jsonConfig)
	Save()

	// try to get it
	Read()
	_, err := GetJSONObject("newone")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
