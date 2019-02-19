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

func TestLdapConnectDisconnect(t *testing.T) {
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

	// we need our global client
	ldapClient = ldapClientNew()

	// we mock the connection
	var mockedConnection gopilotLdapConnectionMocked
	ldapClient.conn = &mockedConnection
	ldapClient.isConnected = false

	ldapClient.BindConnect("", 0, "", "")
}

func TestCreateOUAndCheckIfExist(t *testing.T) {

	// add an orga
	err, orga := organizationCreate("dc=test", "integration")
	if err != nil {
		t.Error(err)
		return
	}
	err = orga.Add(ldapClient)
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
	err = orgaUnit.Add(ldapClient)
	if err != nil {
		t.Error(err)
		return
	}

	// check if it exist
	err, result := orgaUnit.Exists(ldapClient)
	if result == false {
		t.Error(err)
		return
	}

	// read and compare
	// err, curUnit :=
	// err, existingObject := GetLdapObject(ldapCon, orgaUnit.Dn)

}

func TestCreateUserAndModifyIt(t *testing.T) {

	// add an user
	err, user := inetOrgPersonCreate("dc=integration,dc=test", "testuser", "Test User", "Georg")
	if err != nil {
		t.Error(err.Error())
		return
	}
	user.Add(ldapClient)

	// get it back
	err, curUserObject := GetLdapObject(ldapClient, "uid=testuser,dc=integration,dc=test")
	if err != nil {
		t.Error(err.Error())
		return
	}

	// get sn
	err, curUseValues := curUserObject.GetAttrValue("sn")
	if err != nil {
		t.Error(err.Error())
		return
	}

	// check if sn is correct
	if curUseValues[0] != "Georg" {
		t.Error("Error on user")
		return
	}

	// we change an attribute
	user.SetAttrValue("sn", []string{"John"})
	err = user.Change(ldapClient)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// get it back
	err, curUserObject = GetLdapObject(ldapClient, "uid=testuser,dc=integration,dc=test")
	if err != nil {
		t.Error(err.Error())
		return
	}

	// get sn
	err, curUseValues = curUserObject.GetAttrValue("sn")
	if err != nil {
		t.Error(err.Error())
		return
	}

	// check if sn is correct
	if curUseValues[0] != "John" {
		t.Error("Error on user")
		return
	}

	// change element which not exists before
	user.SetAttrValue("mail", []string{"john@gmail.com"})
	err = user.Change(ldapClient)
	if err != nil {
		t.Error(err.Error())
		return
	}

	// get it back
	err, curUserObject = GetLdapObject(ldapClient, "uid=testuser,dc=integration,dc=test")
	if err != nil {
		t.Error(err.Error())
		return
	}

	// get sn
	err, curUseValues = curUserObject.GetAttrValue("mail")
	if err != nil {
		t.Error(err.Error())
		return
	}

	// check if sn is correct
	if curUseValues[0] != "john@gmail.com" {
		t.Error("Error on user")
		return
	}

	// remove it
	err = curUserObject.Remove(ldapClient)
	if err != nil {
		t.Error(err.Error())
		return
	}
}
