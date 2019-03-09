package pluginldap

import (
	"errors"
	"sort"
	"strings"
)

var globalObjClasses []string

type ldapClass struct {
	ObjectClass []string `json:"objectClass"`

	attrMain string
	attrMust []string
	attrMay  []string
}

var ldapAllClasses map[string]ldapClass = nil

var ldapAllAttrs map[string]bool = nil

func globalAttrsString() []string {

	var stringArray []string

	for attrName := range ldapAllAttrs {
		stringArray = append(stringArray, attrName)
	}

	return stringArray
}

func ldapClassInit() {

	// create a new global class mapping
	if ldapAllClasses == nil {
		ldapAllClasses = make(map[string]ldapClass)
	}
	// create global attributes if needed
	if ldapAllAttrs == nil {
		ldapAllAttrs = make(map[string]bool)
		ldapAllAttrs["objectClass"] = false
		ldapAllAttrs["dn"] = false
	}

}

func ldapClassRegister(objectClass []string, attrMain string, attrMust []string, attrMay []string) ldapClass {

	// create an new class-object
	newClass := ldapClass{
		ObjectClass: objectClass,
		attrMain:    attrMain,
		attrMust:    attrMust,
		attrMay:     attrMay,
	}

	// add main attribute to attributes
	ldapAllAttrs[attrMain] = false

	// read names of must attributes and add it to global attribute mapping
	for _, attrName := range attrMust {
		ldapAllAttrs[attrName] = false
	}
	// read names of may attributes and add it to global attribute mapping
	for _, attrName := range attrMay {
		ldapAllAttrs[attrName] = false
	}

	// sort classes for mapping kay
	var sortedClassArray = make([]string, len(objectClass))
	copy(sortedClassArray, objectClass)
	sort.Strings(sortedClassArray)
	sortedClassString := strings.Join(sortedClassArray, ",")

	// register the class name
	globalObjClasses = append(globalObjClasses, objectClass...)

	// register the class
	ldapAllClasses[sortedClassString] = newClass
	logging.Debug("ldapClassRegister", sortedClassString)

	return newClass
}

func ldapClassGet(objectClassesToFind []string) (error, *ldapClass) {

	sort.Strings(objectClassesToFind)
	sortedClassString := strings.Join(objectClassesToFind, ",")

	if ldapClass, ok := ldapAllClasses[sortedClassString]; ok {
		return nil, &ldapClass
	}

	return errors.New("Could not find class '" + sortedClassString + "'"), nil
}

func ldapClassCreateLdapObject(objectClassesToFind []string) (error, *ldapObject) {

	// get the class definition
	err, ldapClass := ldapClassGet(objectClassesToFind)
	if err != nil {
		return err, nil
	}

	// create a new object with class definition
	newLdapObject := ldapObjectCreate(ldapClass.ObjectClass, "", ldapClass.attrMain)
	newLdapObject.SetAttrDefinition(ldapClass.attrMust, ldapClass.attrMay)

	return nil, &newLdapObject
}
