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

package ctls

import "testing"
import "fmt"

func TestHMAC(t *testing.T) {

	hash := ComputeHmac256("message", "secret")
	secondHash := ComputeHmac256("message", "secret")

	if hash != secondHash {
		t.Error("Hash function incorrect")
		t.FailNow()
		return
	}
	fmt.Println("Test OK: ", hash)

}
