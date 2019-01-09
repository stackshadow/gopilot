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

/*
You need to run
docker run \
--rm \
-ti \
-e AdminDN=cn=admin,cn=config \
-e AdminPW=secret \
-e ROOTDN=cn=admin,dc=integration,dc=test \
-e ROOTPW=secret \
-e DBSuffix=dc=integration,dc=test \
-p 389:389 \
archlinux-ldap:latest

*/

package ldapclient

import "testing"
import "core/clog"
import "core/msgbus"

var TestHostName = "localhost"
var TestBindDN = ""
var TestPassword = ""

func TestConnectDisconnect(t *testing.T) {
	clog.Init()
	clog.EnableDebug()
	msgbus.MsgBusInit()
	msgbus.PluginsInit()

	err := BindConnect("localhost", 389, "cn=admin,dc=integration,dc=test", "secret")
	if err != nil {
		t.Error(err.Error())
		return
	}
	disconnect()

}

func TestCreateOUAndCheckIfExist(t *testing.T) {

	err := BindConnect("localhost", 389, "cn=admin,dc=integration,dc=test", "secret")
	if err != nil {
		t.Error(err.Error())
		return
	}

	// add an orga
	orga := organizationInit("dc=test", "integration")
	orga.Add(ldapCon)
	err = orga.Add(ldapCon)
	if err != nil {
		t.Error(err)
		return
	}

	// add an orga object
	orgaUnit := organizationalUnitInit("dc=integration,dc=test", "groups", "The folder for all groups")
	orgaUnit.Add(ldapCon)
	err = orga.Add(ldapCon)
	if err != nil {
		t.Error(err)
		return
	}

	// check if it exist
	err, result := orgaUnit.Exists(ldapCon)
	if result == false {
		t.Error("orgaUnit should exist")
		return
	}

	disconnect()
}
