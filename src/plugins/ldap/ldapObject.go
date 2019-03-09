package pluginldap

//import "errors"
import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/ldap.v3"
	"strings"
)

type ldapObject struct {
	DnBase      string   `json:"basedn"`
	ObjectClass []string `json:"objectClass"`
	Dn          string   `json:"dn"`

	AttrMain string              `json:"attrMain"`
	AttrMust map[string][]string `json:"attrMust"`
	AttrMay  map[string][]string `json:"attrMay"`
}

func ldapObjectCreate(objectClass []string, basedn, attrMain string) ldapObject {

	newObject := ldapObject{
		DnBase:      basedn,
		ObjectClass: objectClass,
		Dn:          "",
		AttrMain:    attrMain,
	}

	newObject.AttrMust = make(map[string][]string)
	newObject.AttrMay = make(map[string][]string)

	logging.Debug("ldapObject::Create", "Create new object with DN '"+newObject.Dn+"'")

	return newObject
}

func (curObject *ldapObject) SetAttrDefinition(attrsMust []string, attrsMay []string) error {

	// can not overwrite already set
	if len(curObject.AttrMust) > 0 {
		return errors.New("Can not overwrite existing Must-Attributes")
	}
	if len(curObject.AttrMay) > 0 {
		return errors.New("Can not overwrite existing May-Attributes")
	}

	// add Must-Attributes
	for _, attrName := range attrsMust {
		curObject.AttrMust[attrName] = []string{}
	}

	// add May-Attributes
	for _, attrName := range attrsMay {
		curObject.AttrMay[attrName] = []string{}
	}

	return nil
}

func (curObject *ldapObject) SetAttrValue(attributeName string, values []string) error {

	// check if main attr is set and recreate DN
	if attributeName == curObject.AttrMain {
		curObject.Dn = curObject.AttrMain + "=" + values[0]

		// add the base dn ( if not empty )
		if curObject.DnBase != "" {
			curObject.Dn += "," + curObject.DnBase
		}

	}

	// add Must-Attributes
	for attrName := range curObject.AttrMust {
		if attrName == attributeName {
			curObject.AttrMust[attrName] = values

			logging.Debug("ldapObject::SetAttrValue", "Set must-attribute-value '"+attributeName+"' for object with dn '"+curObject.Dn+"'")
			return nil
		}
	}

	// add May-Attributes
	for attrName := range curObject.AttrMay {
		if attrName == attributeName {
			curObject.AttrMay[attrName] = values

			logging.Debug("ldapObject::SetAttrValue", "Set may-attribute-value '"+attributeName+"' for object with dn '"+curObject.Dn+"'")
			return nil
		}
	}

	logging.Error("ldapObject::SetAttrValue", "Attribute '"+attributeName+"' don't exist in definition")
	return errors.New("Attribute don't exist in Object")
}

func (curObject *ldapObject) GetAttrValue(attributeName string) (error, []string) {

	// add Must-Attributes
	for attrName := range curObject.AttrMust {
		if attrName == attributeName {
			return nil, curObject.AttrMust[attrName]
		}
	}

	// add May-Attributes
	for attrName := range curObject.AttrMay {
		if attrName == attributeName {
			return nil, curObject.AttrMay[attrName]
		}
	}

	err := errors.New("Object don't know the '" + attributeName + "' attribute.")
	return err, []string{}
}

/*
@detail
This function don't log
*/
func (curObject *ldapObject) Exists(client SLdapClient) (error, bool) {

	// connected ?
	if client.isConnected == false {
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
	sr, err := client.conn.Search(searchRequest)
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

func (curObject *ldapObject) Add(client SLdapClient) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	// check if already exist
	err, exist := curObject.Exists(client)
	if err != nil {
		logging.Error(curObject.Dn, err.Error())
		return err
	}
	if exist == true {
		logging.Info(curObject.Dn, "Already exsits, do nothing")
		return nil
	}

	// create add-request
	addReq := ldap.NewAddRequest(curObject.Dn, nil)

	// add objectClass
	addReq.Attribute("objectClass", curObject.ObjectClass)

	// add Must-Attributes
	for key, value := range curObject.AttrMust {

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
	for key, value := range curObject.AttrMay {

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
	err = client.conn.Add(addReq)
	if err != nil {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Added")
	return nil
}

func (curObject *ldapObject) Remove(client SLdapClient) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	delRequest := ldap.NewDelRequest(curObject.Dn, nil)
	err := client.conn.Del(delRequest)
	if err != nil {
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Deleted")
	return nil
}

func (curObject *ldapObject) Change(client SLdapClient) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	// create change-request
	changeReq := ldap.NewModifyRequest(curObject.Dn, nil)

	// add Must-Attributes
	for key, value := range curObject.AttrMust {

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
	for key, value := range curObject.AttrMay {

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
	err := client.conn.Modify(changeReq)
	if err != nil {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Changed")
	return nil
}

func (curObject *ldapObject) Rename(client SLdapClient, oldDN string) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	var mainAttrValue []string
	var ok bool
	if mainAttrValue, ok = curObject.AttrMust[curObject.AttrMain]; !ok {
		err := errors.New("'" + curObject.AttrMain + "' Not found in .AttrMust")
		return err
	}

	// check if there are values
	if len(mainAttrValue) == 0 {
		err := errors.New("No Values for AttrMain")
		return err
	}

	req := ldap.NewModifyDNRequest(oldDN, curObject.AttrMain+"="+mainAttrValue[0], true, "")

	if err := client.conn.ModifyDN(req); err != nil {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		return err
	}

	// set the new DN
	curObject.Dn = curObject.AttrMain + "=" + mainAttrValue[0] + "," + curObject.DnBase

	return nil
}

func (curObject *ldapObject) AddAttribute(client SLdapClient, attrName string, attrValues []string) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	// create change-request
	changeReq := ldap.NewModifyRequest(curObject.Dn, nil)
	changeReq.Add(attrName, attrValues)

	// Send out the request
	err := client.conn.Modify(changeReq)
	if err != nil {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Changed")
	return nil
}

func (curObject *ldapObject) ReplaceAttribute(client SLdapClient, attrName string, attrValues []string) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	// create change-request
	changeReq := ldap.NewModifyRequest(curObject.Dn, nil)
	changeReq.Replace(attrName, attrValues)

	// Send out the request
	err := client.conn.Modify(changeReq)
	if err != nil {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		return err
	}

	logging.Info(fmt.Sprintf("%s", curObject.Dn), "Changed")
	return nil
}

func (curObject *ldapObject) RemoveAttribute(client SLdapClient, attrName string, attrValues []string) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	// create change-request
	changeReq := ldap.NewModifyRequest(curObject.Dn, nil)
	changeReq.Delete(attrName, attrValues)

	// Send out the request
	err := client.conn.Modify(changeReq)
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

func (curObject *ldapObject) ToJsonString() string {

	// convert to bytearray
	groupObjectBytes, err := json.Marshal(&curObject)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}

	logging.Debug(fmt.Sprintf("%s", curObject.Dn), string(groupObjectBytes))

	return string(groupObjectBytes)
}

func (curObject *ldapObject) GetClassElements(client SLdapClient, basedn string, callback func(*ldap.Entry)) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
	}

	var classFilter string = "(|"
	for index := range curObject.ObjectClass {
		classFilter += "(objectClass="
		classFilter += curObject.ObjectClass[index]
		classFilter += ")"
	}
	classFilter += ")"

	// attr list
	var attrs []string
	for key, _ := range curObject.AttrMust {
		attrs = append(attrs, key)
	}
	for key, _ := range curObject.AttrMay {
		attrs = append(attrs, key)
	}

	searchRequest := ldap.NewSearchRequest(
		basedn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		classFilter,
		attrs,
		nil,
	)

	sr, err := client.conn.Search(searchRequest)
	if err != nil {
		logging.Error("SearchAll", err.Error())
		return err
	}

	for _, entry := range sr.Entries {
		if callback != nil {
			callback(entry)
		}
	}

	return nil
}

func SearchOneLevel(client SLdapClient, basedn string, callback func(*ldap.Entry)) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
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

	sr, err := client.conn.Search(searchRequest)
	if err != nil {
		logging.Error("SearchAll", err.Error())
		return err
	}
	//sr.PrettyPrint(2)

	for _, entry := range sr.Entries {
		if callback != nil {
			callback(entry)
		}
		//entry.PrettyPrint(2)
	}

	return nil
}

func SearchOneFull(client SLdapClient, fulldn string, callback func(*ldap.Entry)) error {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected")
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

	sr, err := client.conn.Search(searchRequest)
	if err != nil {
		logging.Error("SearchAll", err.Error())
		return err
	}
	//sr.PrettyPrint(2)

	if len(sr.Entries) > 0 {
		if callback != nil {
			callback(sr.Entries[0])
		}
		//entry.PrettyPrint(2)
	}

	return nil
}

/**
Get an ldapObject from your ldap-server
 - Will automatic calculate the base DN
**/
func GetLdapObject(client SLdapClient, fullDn string) (error, *ldapObject) {

	// connected ?
	if client.isConnected == false {
		return errors.New("Not connected"), nil
	}

	var retError error
	var retLdapObject *ldapObject

	err := SearchOneFull(client, fullDn, func(entry *ldap.Entry) {

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

		// we need to grab the base dn from the dn
		splitDN := strings.SplitN(dn, ",", 2)
		if len(splitDN) <= 1 {
			retError = errors.New("Could not read basedn from dn")
			retLdapObject = ldapObject
			return
		}
		ldapObject.DnBase = splitDN[1]

		retError = nil
		retLdapObject = ldapObject
		return
	})

	if err != nil {
		retError = err
	}

	return retError, retLdapObject
}
