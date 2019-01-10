package ldapclient

//import "errors"
import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/ldap.v2"
)

var globalObjClasses []string

var globalAttrs map[string]bool = make(map[string]bool)

func globalAttrAdd(attrName string) {
	globalAttrs[attrName] = false
}
func globalAttrGet() []string {
	var objectAttributes = []string{}

	for key := range globalAttrs {
		objectAttributes = append(objectAttributes, key)
	}

	return objectAttributes
}

type ldapObject struct {
	DnBase      string   `json:"basedn"`
	ObjectClass []string `json:"objectClass"`
	Dn          string   `json:"dn"`

	attrMain string
	attrMust map[string][]string
	attrMay  map[string][]string
}

type ldapObjectJson struct {
	attrMust map[string]interface{}
}

func ldapObjectCreate(objectClass []string, basedn, attrMain, attrMainValue string) ldapObject {

	newObject := ldapObject{
		DnBase:      basedn,
		ObjectClass: objectClass,
		Dn:          attrMain + "=" + attrMainValue,
		attrMain:    attrMain,
	}

	if basedn != "" {
		newObject.Dn += "," + basedn
	}

	newObject.attrMust = make(map[string][]string)
	newObject.attrMay = make(map[string][]string)

	// set the main attr as MUST ( because it is )
	newObject.SetMustAttr(attrMain, []string{attrMainValue})

	//
	globalObjClasses = append(globalObjClasses, objectClass...)

	return newObject
}

func (curObject *ldapObject) SetMustAttr(name string, value []string) {

	//
	if name == "" {
		return
	}

	// set must-attribute
	curObject.attrMust[name] = value

	// if this attribute was the main attribute, we recreate the dn
	if name == curObject.attrMain {
		curObject.Dn = name + "=" + value[0] + "," + curObject.DnBase
	}

	// remember it the attribute
	globalAttrAdd(name)
}

func (curObject *ldapObject) SetMayAttr(name string, value []string) {

	//
	if name == "" {
		return
	}

	curObject.attrMay[name] = value

	// remember it the attribute
	globalAttrAdd(name)
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
		addReq.Attribute(key, value)
	}

	// add May-Attributes
	for key, value := range curObject.attrMay {
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
		changeReq.Replace(key, value)

	}

	// add May-Attributes
	for key, value := range curObject.attrMay {
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

func (curObject *ldapObject) ToMap() map[string]interface{} {

	// build data
	var curAttr = make(map[string]interface{})
	for key, value := range curObject.attrMust {
		curAttr[key] = value
	}
	for key, value := range curObject.attrMay {
		curAttr[key] = value
	}

	// build mapping
	var newJson = make(map[string]interface{})
	newJson["basedn"] = curObject.DnBase
	newJson["objectClass"] = curObject.ObjectClass
	newJson["dn"] = curObject.Dn
	newJson["attrMain"] = curObject.attrMain
	newJson["attr"] = curAttr

	return newJson
}

func (curObject *ldapObject) ToJsonString() string {

	// build mapping
	var newJson = curObject.ToMap()

	// convert to bytearray
	groupObjectBytes, err := json.Marshal(&newJson)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}

	logging.Debug(fmt.Sprintf("%s", curObject.Dn), string(groupObjectBytes))

	return string(groupObjectBytes)
}

func (curObject *ldapObject) ToTemplate() map[string]interface{} {

	// we use the original object
	var newJson = curObject.ToMap()

	// Create array with must attributes
	var mustAttrArray []string
	for key := range curObject.attrMust {
		mustAttrArray = append(mustAttrArray, key)
	}

	// Create array with may attributes
	var mayAttrArray []string
	for key := range curObject.attrMay {
		mayAttrArray = append(mayAttrArray, key)
	}

	newJson["must"] = mustAttrArray
	newJson["may"] = mayAttrArray

	return newJson
}

func (curObject *ldapObject) IsClass(entry ldap.Entry, className string) bool {
	classes := entry.GetAttributeValues("objectClass")

	for index := range classes {
		fmt.Println("index:", index)
	}

	return false
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
		globalAttrGet(),
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
		globalAttrGet(),
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
