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
func organizationInit(basedn, name string) ldapObject {

	newObject := ldapObjectCreate(
		[]string{"top", "dcObject", "organization"},
		basedn,
		"dc",
		name,
	)

	newObject.SetMustAttr("o", name)

	return newObject
}
