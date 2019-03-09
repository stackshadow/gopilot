package pluginLdap

/*
class: organizationalUnit
MUST: ou
MAX: userPassword $ searchGuide $ seeAlso $ businessCategory $ x121Address $
registeredAddress $ destinationIndicator $ preferredDeliveryMethod $ telexNumber
$ teletexTerminalIdentifier $ telephoneNumber $ internationaliSDNNumber $
facsimileTelephoneNumber $ street $ postOfficeBox $ postalCode $ postalAddress
$ physicalDeliveryOfficeName $ st $ l $ description
*/
func ldapClassOrganizationRegister() {

	ldapClassRegister(
		[]string{"top", "dcObject", "organization"},
		"dc",
		[]string{"dc", "o"},
		[]string{},
	)

}

func organizationCreate(basedn, name string) (error, *ldapObject) {

	err, newObject := ldapClassCreateLdapObject([]string{"top", "dcObject", "organization"})
	if err != nil {
		return err, nil
	}

	// set the basedn
	newObject.DnBase = basedn

	// set the main attr
	newObject.SetAttrValue("dc", []string{name})
	newObject.SetAttrValue("o", []string{name})

	return nil, newObject
}
