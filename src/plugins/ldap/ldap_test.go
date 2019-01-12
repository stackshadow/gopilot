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

var LdapTestHostName = "localhost"
var LdapTestBindDN = "cn=admin,dc=integration,dc=test"
var LdapTestPassword = "secret"

func TestConnectDisconnect(t *testing.T) {
	clog.Init()
	clog.EnableDebug()
	msgbus.MsgBusInit()
	msgbus.PluginsInit()

	// we init the global class storage
	ldapClassInit()

	// we create for every type a class to create the objectClass and attributes search strings
	ldapClassOrganizationRegister()
	ldapClassOrganizationalUnitRegister()
	ldapClassInetOrgPersonRegister()
	ldapClassgroupOfNamesRegister()

	err := BindConnect("localhost", 389, "cn=admin,dc=integration,dc=test", "secret")
	if err != nil {
		t.Error(err.Error())
		return
	}
	disconnect()

}

func TestCreateOUAndCheckIfExist(t *testing.T) {

	// connect
	err := BindConnect(LdapTestHostName, 389, LdapTestBindDN, LdapTestPassword)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// add an orga
	err, orga := organizationCreate("dc=test", "integration")
	if err != nil {
		t.Error(err)
		return
	}
	err = orga.Add(ldapCon)
	if err != nil {
		t.Error(err)
		return
	}

	// add an orga object
	err, orgaUnit := organizationalUnitCreate("dc=integration,dc=test", "groups", "The folder for all groups")
	if err != nil {
		t.Error(err)
		return
	}
	err = orgaUnit.Add(ldapCon)
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

	// read and compare
	/* err, curUnit := */
	// err, existingObject := GetLdapObject(ldapCon, orgaUnit.Dn)

	disconnect()
}

func TestCreateUserAndModifyIt(t *testing.T) {

	// connect
	err := BindConnect(LdapTestHostName, 389, LdapTestBindDN, LdapTestPassword)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// add an user
	_, user := inetOrgPersonCreate("dc=integration,dc=test", "testuser", "Test User", "Mr")
	user.Add(ldapCon)

	// we change an attribute
	user.SetAttrValue("mail", []string{"usermail"})
	user.Change(ldapCon)

	disconnect()

}
