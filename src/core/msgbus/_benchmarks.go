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

func BenchmarkMessaging(b *testing.B) {
	Init()

	var firstPluginID, secondPluginID int

	Register(&firstPluginID, "First Plugin")
	Register(&secondPluginID, "Second Plugin")
	ListenForGroup(&secondPluginID, "groupa", benchMessage)

	for n := 0; n < b.N; n++ {
		Publish(&firstPluginID, "me", "other", "groupa", "ping", "nopayload")
	}
}

func benchMessage(group, command, payload string) {

}
