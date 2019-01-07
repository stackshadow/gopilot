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
func inetOrgPersonInit(basedn, uid, cn, sn string) ldapObject {

	newObject := ldapObjectCreate(
		[]string{"inetOrgPerson"},
		basedn,
		"uid",
		uid,
	)

	newObject.SetMustAttr("cn", cn)
	newObject.SetMustAttr("sn", sn)

	newObject.SetMayAttr("mail", "")
	newObject.SetMayAttr("displayName", "")
	newObject.SetMayAttr("userPassword", "")

	return newObject
}
