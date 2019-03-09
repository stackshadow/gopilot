package pluginLdap

/*
class: inetOrgPerson
MUST: cn
MAX: member businessCategory description o ou owner seeAlso
*/
func ldapClassgroupOfNamesRegister() {

	ldapClassRegister(
		[]string{"groupOfNames"},
		"cn",
		[]string{"cn", "member"},
		[]string{},
	)

}

func groupOfNamesCreate(basedn, cn, firstmember string) (error, *ldapObject) {

	err, newObject := ldapClassCreateLdapObject([]string{"groupOfNames"})
	if err != nil {
		return err, nil
	}

	// set the basedn
	newObject.DnBase = basedn

	// set the main attr
	newObject.SetAttrValue("cn", []string{cn})
	newObject.SetAttrValue("member", []string{firstmember})

	return nil, newObject
}
