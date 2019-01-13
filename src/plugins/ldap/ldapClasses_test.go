package ldapclient

import "testing"

func TestCreateAndFindClasses(t *testing.T) {

	ldapClassInit()

	ldapClassRegister(
		[]string{"top", "dcObject", "organization"},
		"dc",
		[]string{"dc", "o"},
		[]string{},
	)

	ldapClassRegister(
		[]string{"organizationalUnit"},
		"ou",
		[]string{"ou", "description"},
		[]string{},
	)

	ldapClassRegister(
		[]string{"inetOrgPerson"},
		"uid",
		[]string{"uid", "cn", "sn"},
		[]string{"mail", "displayName", "userPassword"},
	)

	// there should be 9 elements
	if len(ldapAllAttrs) != 12 {
		t.FailNow()
		return
	}

	// this should not exist
	err, _ := ldapClassGet([]string{"organizationalUni"})
	if err == nil {
		t.FailNow()
		return
	}

	// this also not
	err, _ = ldapClassGet([]string{"dcObject"})
	if err == nil {
		t.FailNow()
	}

	// this should exist
	err, _ = ldapClassGet([]string{"organizationalUnit"})
	if err != nil {
		t.FailNow()
	}

	// this should exist
	err, _ = ldapClassGet([]string{"organization", "top", "dcObject"})
	if err != nil {
		t.FailNow()
	}

	/*
		// get an ldapObject from class
		err, curLdapObject := ldapClassesGetLdapObject([]string{"organization", "top", "dcObject"})
		if err != nil {
			t.FailNow()
		}
	*/
}
