package ldapclient

//import "errors"
import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/ldap.v2"
)

type ldapObject struct {
	DnBase      string   `json:"basedn"`
	ObjectClass []string `json:"objectClass"`
	Dn          string   `json:"dn"`

	attrMain string
	attrMust map[string][]string
	attrMay  map[string][]string
}

type ldapObjectJson struct {
	DnBase      string              `json:"basedn"`
	ObjectClass []string            `json:"objectClass"`
	Dn          string              `json:"dn"`
	AttrMain    string              `json:"attrMain"`
	AttrMust    []string            `json:"attrMust"`
	AttrMay     []string            `json:"attrMay"`
	AttrData    map[string][]string `json:"attrData"`
}

func ldapObjectCreate(objectClass []string, basedn, attrMain string) ldapObject {

	newObject := ldapObject{
		DnBase:      basedn,
		ObjectClass: objectClass,
		Dn:          "",
		attrMain:    attrMain,
	}

	newObject.attrMust = make(map[string][]string)
	newObject.attrMay = make(map[string][]string)

	logging.Debug("ldapObject::Create", "Create new object with DN '"+newObject.Dn+"'")

	return newObject
}

func (curObject *ldapObject) SetAttrDefinition(attrsMust []string, attrsMay []string) error {

	// can not overwrite already set
	if len(curObject.attrMust) > 0 {
		return errors.New("Can not overwrite existing Must-Attributes")
	}
	if len(curObject.attrMay) > 0 {
		return errors.New("Can not overwrite existing May-Attributes")
	}

	// add Must-Attributes
	for _, attrName := range attrsMust {
		curObject.attrMust[attrName] = []string{}
	}

	// add May-Attributes
	for _, attrName := range attrsMay {
		curObject.attrMay[attrName] = []string{}
	}

	return nil
}

func (curObject *ldapObject) SetAttrValue(attributeName string, values []string) error {

	// check if main attr is set and recreate DN
	if attributeName == curObject.attrMain {
		curObject.Dn = curObject.attrMain + "=" + values[0]

		// add the base dn ( if not empty )
		if curObject.DnBase != "" {
			curObject.Dn += "," + curObject.DnBase
		}

	}

	// add Must-Attributes
	for attrName := range curObject.attrMust {
		if attrName == attributeName {
			curObject.attrMust[attrName] = values

			logging.Debug("ldapObject::SetAttrValue", "Set must-attribute-value '"+attributeName+"' for object with dn '"+curObject.Dn+"'")
			return nil
		}
	}

	// add May-Attributes
	for attrName := range curObject.attrMay {
		if attrName == attributeName {
			curObject.attrMay[attrName] = values

			logging.Debug("ldapObject::SetAttrValue", "Set may-attribute-value '"+attributeName+"' for object with dn '"+curObject.Dn+"'")
			return nil
		}
	}

	logging.Error("ldapObject::SetAttrValue", "Attribute '"+attributeName+"' don't exist in definition")
	return errors.New("Attribute don't exist in Object")
}

func (curObject *ldapObject) GetAttrValue(attributeName string) (error, []string) {

	// add Must-Attributes
	for attrName := range curObject.attrMust {
		if attrName == attributeName {
			return nil, curObject.attrMust[attrName]
		}
	}

	// add May-Attributes
	for attrName := range curObject.attrMay {
		if attrName == attributeName {
			return nil, curObject.attrMay[attrName]
		}
	}

	err := errors.New("Object don't know the '" + attributeName + "' attribute.")
	return err, []string{}
}

/*
@detail
This function don't log
*/
func (curObject *ldapObject) Exists(ldapConnection *ldap.Conn) (error, bool) {

	// connected ?
	if ldapConnection == nil {
		return errors.New("Not connected"), false
	}

	// create serch request
	searchRequest := ldap.NewSearchRequest(
		curObject.Dn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass="+curObject.ObjectClass[0]+"))",
		[]string{"dn"},
		nil,
	)

	// send serch request
	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {

		// on resultcode 32 ( not found ) no error occured
		if err, ok := err.(*ldap.Error); ok {
			if err.ResultCode == 32 {
				return nil, false
			}
			return err, false
		}

		return err, false
	}

	// did we found something ?
	if len(sr.Entries) > 0 {
		return nil, true
	}

	return nil, false
}

func (curObject *ldapObject) Add(ldapConnection *ldap.Conn) error {

	// connected ?
	if ldapConnection == nil {
		return errors.New("Not connected")
	}

	// check if already exist
	err, exist := curObject.Exists(ldapConnection)
	if err != nil {
		logging.Error(curObject.Dn, err.Error())
		return err
	}
	if exist == true {
		logging.Info(curObject.Dn, "Already exsits, do nothing")
		return nil
	}

	// create add-request
	addReq := ldap.NewAddRequest(curObject.Dn)

	// add objectClass
	addReq.Attribute("objectClass", curObject.ObjectClass)

	// add Must-Attributes
	for key, value := range curObject.attrMust {

		// check for empty values
		if len(value) <= 0 {
			err := errors.New(fmt.Sprintf("Value for mandatory attribute '%s' missing.", key))
			logging.Error("Add", err.Error())
			return err
		}
		if len(value) == 1 {
			if value[0] == "" {
				err := errors.New(fmt.Sprintf("Value for mandatory attribute '%s' missing.", key))
				logging.Error("Add", err.Error())
				return err
			}
		}

		addReq.Attribute(key, value)
	}

	// add May-Attributes
	for key, value := range curObject.attrMay {

		// ommit empty value
		if len(value) <= 0 {
			continue
		}
		if len(value) == 1 {
			if value[0] == "" {
				continue
			}
		}

		addReq.Attribute(key, value)
	}

	// Send out the request
	err = ldapConnection.Add(addReq)
	if err != nil {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Added")
	return nil
}

func (curObject *ldapObject) Remove(ldapConnection *ldap.Conn) error {

	// connected ?
	if ldapConnection == nil {
		return errors.New("Not connected")
	}

	delRequest := ldap.NewDelRequest(curObject.Dn, nil)
	err := ldapConnection.Del(delRequest)
	if err != nil {
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Deleted")
	return nil
}

func (curObject *ldapObject) Change(ldapConnection *ldap.Conn) error {

	// connected ?
	if ldapConnection == nil {
		return errors.New("Not connected")
	}

	// create change-request
	changeReq := ldap.NewModifyRequest(curObject.Dn)

	// add Must-Attributes
	for key, value := range curObject.attrMust {

		// ommit empty value
		if len(value) <= 0 {
			continue
		}
		if len(value) == 1 {
			if value[0] == "" {
				continue
			}
		}

		changeReq.Replace(key, value)
	}

	// add May-Attributes
	for key, value := range curObject.attrMay {

		// ommit empty value
		if len(value) <= 0 {
			continue
		}
		if len(value) == 1 {
			if value[0] == "" {
				continue
			}
		}

		changeReq.Replace(key, value)
	}

	// Send out the request
	err := ldapConnection.Modify(changeReq)
	if err != nil {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Changed")
	return nil
}

func (curObject *ldapObject) IsClass(entry ldap.Entry, className string) bool {
	classes := entry.GetAttributeValues("objectClass")

	for index := range classes {
		fmt.Println("index:", index)
	}

	return false
}

func (curObject *ldapObject) ToJsonObject() ldapObjectJson {

	ldapObject := ldapObjectJson{
		DnBase:      curObject.DnBase,
		ObjectClass: curObject.ObjectClass,
		Dn:          curObject.Dn,
		AttrMain:    curObject.attrMain,
		AttrMust:    []string{},
		AttrMay:     []string{},
		AttrData:    make(map[string][]string),
	}

	// set AttrData
	for key, value := range curObject.attrMust {
		ldapObject.AttrData[key] = value
	}
	for key, value := range curObject.attrMay {
		ldapObject.AttrData[key] = value
	}

	return ldapObject
}

func (curObject *ldapObject) ToJsonString() string {

	// build mapping
	var newJson = curObject.ToJsonObject()

	// convert to bytearray
	groupObjectBytes, err := json.Marshal(&newJson)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}

	logging.Debug(fmt.Sprintf("%s", curObject.Dn), string(groupObjectBytes))

	return string(groupObjectBytes)
}

func (curObject *ldapObject) ToTemplateString() string {

	// we use the original object
	var newJson = curObject.ToJsonObject()

	// Create array with must attributes
	newJson.AttrMust = []string{}
	for key := range curObject.attrMust {
		newJson.AttrMust = append(newJson.AttrMust, key)
	}

	// Create array with may attributes
	newJson.AttrMay = []string{}
	for key := range curObject.attrMay {
		newJson.AttrMay = append(newJson.AttrMay, key)
	}

	// convert to bytearray
	groupObjectBytes, err := json.Marshal(&newJson)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}

	logging.Debug(fmt.Sprintf("%s", curObject.Dn), string(groupObjectBytes))

	return string(groupObjectBytes)
}

func SearchAllFull(ldapConnection *ldap.Conn, basedn string, callback func(*ldap.Entry)) {

	// connected ?
	if ldapConnection == nil {
		return
	}

	var classFilter string = "(|"
	for index := range globalObjClasses {
		classFilter += "(objectClass="
		classFilter += globalObjClasses[index]
		classFilter += ")"
	}
	classFilter += ")"

	searchRequest := ldap.NewSearchRequest(
		basedn,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		classFilter,
		globalAttrsString(),
		nil,
	)

	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {
		logging.Error("SearchAll", err.Error())
		return
	}
	//sr.PrettyPrint(2)

	for _, entry := range sr.Entries {
		if callback != nil {
			callback(entry)
		}
		//entry.PrettyPrint(2)
	}

}

func SearchOneFull(ldapConnection *ldap.Conn, fulldn string, callback func(*ldap.Entry)) {

	// connected ?
	if ldapConnection == nil {
		return
	}

	var classFilter string = "(|"
	for index := range globalObjClasses {
		classFilter += "(objectClass="
		classFilter += globalObjClasses[index]
		classFilter += ")"
	}
	classFilter += ")"

	searchRequest := ldap.NewSearchRequest(
		fulldn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		classFilter,
		globalAttrsString(),
		nil,
	)

	sr, err := ldapConnection.Search(searchRequest)
	if err != nil {
		logging.Error("SearchAll", err.Error())
		return
	}
	//sr.PrettyPrint(2)

	if len(sr.Entries) > 0 {
		if callback != nil {
			callback(sr.Entries[0])
		}
		//entry.PrettyPrint(2)
	}

}

func GetLdapObject(ldapConnection *ldap.Conn, fullDn string) (error, *ldapObject) {

	// connected ?
	if ldapConnection == nil {
		return errors.New("Not connected"), nil
	}

	var retError error
	var retLdapObject *ldapObject

	SearchOneFull(ldapConnection, fullDn, func(entry *ldap.Entry) {

		objectClass := entry.GetAttributeValues("objectClass")
		dn := entry.DN

		// get the object of the corresponding class
		err, ldapObject := ldapClassCreateLdapObject(objectClass)
		if err != nil {
			retError = err
			retLdapObject = nil
			return
		}

		// set all readed attributes
		for _, attribute := range entry.Attributes {

			// we dont set objectClass
			if attribute.Name == "objectClass" {
				continue
			}

			ldapObject.SetAttrValue(attribute.Name, attribute.Values)
		}

		// set the dn ( because it get first manipulated from SetAttrValue)
		ldapObject.Dn = dn

		retError = nil
		retLdapObject = ldapObject
		return
	})

	return retError, retLdapObject
}
