package ldapclient

/*
RFC: https://tools.ietf.org/html/rfc2798 page 6
class: inetOrgPerson
MUST: 	cn $ objectClass $ sn
MAY: 	description $ destinationIndicator $ facsimileTelephoneNumber $
        internationaliSDNNumber $ l $ ou $ physicalDeliveryOfficeName $
        postalAddress $ postalCode $ postOfficeBox $
        preferredDeliveryMethod $ registeredAddress $ seeAlso $
        st $ street $ telephoneNumber $ teletexTerminalIdentifier $
        telexNumber $ title $ userPassword $ x121Address
*/
func ldapClassInetOrgPersonRegister() {

	ldapClassRegister(
		[]string{"inetOrgPerson"},
		"uid",
		[]string{"uid", "cn", "sn"},
		[]string{"mail", "displayName", "userPassword", "memberOf"},
	)

}

func inetOrgPersonCreate(basedn, uid, cn, sn string) (error, *ldapObject) {

	err, newObject := ldapClassCreateLdapObject([]string{"inetOrgPerson"})
	if err != nil {
		return err, nil
	}

	// set the basedn
	newObject.DnBase = basedn

	// set the main attr
	newObject.SetAttrValue("uid", []string{uid})
	newObject.SetAttrValue("cn", []string{cn})
	newObject.SetAttrValue("sn", []string{sn})

	return nil, newObject
}
