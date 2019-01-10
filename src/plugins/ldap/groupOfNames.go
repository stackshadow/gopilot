package ldapclient

/*
class: inetOrgPerson
MUST: cn
MAX: member businessCategory description o ou owner seeAlso
*/
func groupOfNamesInit(basedn, cn, firstmember string) ldapObject {

	newObject := ldapObjectCreate(
		[]string{"groupOfNames"},
		basedn,
		"cn",
		cn,
	)

	newObject.SetMustAttr("member", []string{firstmember})

	return newObject
}
