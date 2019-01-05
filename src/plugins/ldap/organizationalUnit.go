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
func organizationalUnitInit(basedn, ou, description string) ldapObject {

	newObject := ldapObjectCreate(
		[]string{"organizationalUnit"},
		basedn,
		"ou",
		ou,
	)

	newObject.SetMayAttr("description", description)

	return newObject
}
