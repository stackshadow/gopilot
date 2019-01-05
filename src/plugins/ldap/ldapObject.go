package ldapclient

//import "errors"
import "encoding/json"
import "fmt"
import "errors"
import "gopkg.in/ldap.v2"

var objectClasses []string
var objectAttributes = []string{"objectClass", "memberof"}

type ldapObject struct {
	DnBase      string   `json:"basedn"`
	ObjectClass []string `json:"objectClass"`
	Dn          string   `json:"dn"`

	attrMain string
	attrMust map[string]string
	attrMay  map[string]string
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

	newObject.attrMust = make(map[string]string)
	newObject.attrMay = make(map[string]string)

	// set the main attr as MUST ( because it is )
	newObject.SetMustAttr(attrMain, attrMainValue)

	//
	objectClasses = append(objectClasses, objectClass...)

	return newObject
}

func (curObject *ldapObject) SetMustAttr(name, value string) {
	curObject.attrMust[name] = value
	objectAttributes = append(objectAttributes, name)
}

func (curObject *ldapObject) SetMayAttr(name, value string) {
	curObject.attrMay[name] = value
	objectAttributes = append(objectAttributes, name)
}

func (curObject *ldapObject) Add(ldapConnection *ldap.Conn) error {

	// connected ?
	if ldapConnection == nil {
		return errors.New("Not connected")
	}

	logging.Debug(fmt.Sprintf("%s", curObject.Dn), "Check if DN already exist")
	searchRequest := ldap.NewSearchRequest(
		curObject.Dn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass="+curObject.ObjectClass[0]+"))",
		[]string{"dn"},
		nil,
	)

	sr, err := ldapConnection.Search(searchRequest)
	if err == nil {
		if len(sr.Entries) > 0 {
			logging.Debug(fmt.Sprintf("%s", curObject.Dn), "Already exists, thats okay.")
			return nil
		}
	} else {
		logging.Error(fmt.Sprintf("%s", curObject.Dn), err.Error())
		//return err
	}

	// create add-request
	addReq := ldap.NewAddRequest(curObject.Dn)
	addReq.Attribute("objectClass", curObject.ObjectClass)
	for key, value := range curObject.attrMust {
		if value != "" {
			addReq.Attribute(key, []string{value})
		}
	}
	for key, value := range curObject.attrMay {
		if value != "" {
			addReq.Attribute(key, []string{value})
		}
	}
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

func (curObject *ldapObject) ToJson(ldapConnection *ldap.Conn) string {

	// connected ?
	if ldapConnection == nil {
		return "Not connected"
	}

	// build mapping
	var newJson = make(map[string]interface{})
	newJson["basedn"] = curObject.DnBase
	newJson["objectClass"] = curObject.ObjectClass
	newJson["dn"] = curObject.Dn
	for key, value := range curObject.attrMust {
		newJson[key] = value
	}
	for key, value := range curObject.attrMay {
		newJson[key] = value
	}

	// convert to json
	groupObjectBytes, err := json.Marshal(&newJson)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}

	logging.Debug(fmt.Sprintf("%s", curObject.Dn), string(groupObjectBytes))

	return string(groupObjectBytes)
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
	for index := range objectClasses {
		classFilter += "(objectClass="
		classFilter += objectClasses[index]
		classFilter += ")"
	}
	classFilter += ")"

	searchRequest := ldap.NewSearchRequest(
		basedn,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		classFilter,
		objectAttributes,
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
	for index := range objectClasses {
		classFilter += "(objectClass="
		classFilter += objectClasses[index]
		classFilter += ")"
	}
	classFilter += ")"

	searchRequest := ldap.NewSearchRequest(
		fulldn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		classFilter,
		objectAttributes,
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
