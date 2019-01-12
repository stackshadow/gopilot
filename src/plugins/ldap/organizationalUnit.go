package ldapclient

/*
class: organizationalUnit
MUST: ou
MAX: userPassword $ searchGuide $ seeAlso $ businessCategory $ x121Address $
registeredAddress $ destinationIndicator $ preferredDeliveryMethod $ telexNumber
$ teletexTerminalIdentifier $ telephoneNumber $ internationaliSDNNumber $
facsimileTelephoneNumber $ street $ postOfficeBox $ postalCode $ postalAddress
$ physicalDeliveryOfficeName $ st $ l $ description
*/

func ldapClassOrganizationalUnitRegister() {

	ldapClassRegister(
		[]string{"organizationalUnit"},
		"ou",
		[]string{"ou"},
		[]string{"description"},
	)

}

func organizationalUnitCreate(basedn, ou, description string) (error, *ldapObject) {

	err, newObject := ldapClassCreateLdapObject([]string{"organizationalUnit"})
	if err != nil {
		return err, nil
	}

	// set the basedn
	newObject.DnBase = basedn

	// set the main attr
	newObject.SetAttrValue("ou", []string{ou})
	newObject.SetAttrValue("description", []string{description})

	return nil, newObject
}
