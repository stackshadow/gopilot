package ldapclient

/*
class: inetOrgPerson
MUST: sn cn
MAX: audio $ businessCategory $ carLicense $ departmentNumber $ displayName
$ employeeNumber $ employeeType $ givenName $ homePhone $ homePostalAddress
$ initials $ jpegPhoto $ labeledURI $ mail $ manager $ mobile $ o $ pager
$ photo $ roomNumber $ secretary $ uid $ userCertificate $ x500uniqueIdentifier
$ preferredLanguage $ userSMIMECertificate $ userPKCS12
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

	return newObject
}
