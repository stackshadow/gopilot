package ldapclient

import (
	"core/clog"
	"core/msgbus"
	"plugins/core"

	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"gopkg.in/ldap.v2"
)

// private vars
var plugin msgbus.Plugin
var logging clog.Logger
var ldapCon *ldap.Conn = nil

var ldapConnected bool
var ldapNamespace string
var ldapBaseOrgaName string
var ldapOrgaObj ldapObject

var ldapUserSuffixDn string
var ldapGroupSuffixDn string
var ldapDummyUserDn string

type ldapConnectionConfig struct {
	Host      string  `json:"host,omitempty"`
	Port      float64 `json:"port,omitempty"`
	BindDN    string  `json:"binddn,omitempty"`
	Password  string  `json:"password,omitempty"`
	Namespace string  `json:"namespace,omitempty"`
	OrgaName  string  `json:"organame,omitempty"`
}

type ldapChangeRequest struct {
	Dn   string              `json:"dn"`
	Attr map[string][]string `json:"attr"`
}

func ParseCmdLine() {
	flag.StringVar(&ldapNamespace, "ldapNamespace", "dc=local", "The namespace where your organisation lives")
	flag.StringVar(&ldapBaseOrgaName, "ldapOrga", "shinyneworga", "Name of your organisation")
}

func Init() {

	logging = clog.New("LDAP")

	// we create for every type a class to create the objectClass and attributes search strings
	organizationInit("", "")
	organizationalUnitInit("", "", "")
	groupOfNamesInit("", "", "")
	inetOrgPersonInit("", "", "", "")

	// we add the global attributes objectClass and dn
	globalAttrAdd("objectClass")
	globalAttrAdd("dn")

	// register plugin on messagebus
	plugin = msgbus.NewPlugin("LDAP")
	plugin.Register()
	plugin.ListenForGroup("ldap", onMessage)
}

func GetLdapConfig() ldapConnectionConfig {
	var jsonObject *map[string]interface{}

	// get the config
	jsonObject, _ = core.GetJsonObject("ldap")
	if jsonObject == nil {
		jsonObject, _ = core.NewJsonObject("ldap")
	}

	// the final struct
	var jsonLdapConfig ldapConnectionConfig

	// convert interface to struct
	groupObjectBytes, err := json.Marshal(jsonObject)
	if err != nil {
		logging.Error("BindConnect", err.Error())
		return jsonLdapConfig
	}

	err = json.Unmarshal([]byte(groupObjectBytes), &jsonLdapConfig)
	if err != nil {
		logging.Error("BindConnect", err.Error())
		return jsonLdapConfig
	}

	// set default values

	if jsonLdapConfig.Host == "" {
		jsonLdapConfig.Host = "127.0.0.1"
	}
	if jsonLdapConfig.Port == 0 {
		jsonLdapConfig.Port = 639
	}
	if jsonLdapConfig.BindDN == "" {
		jsonLdapConfig.BindDN = "cn=admin,dc=test,dc=local"
	}
	if jsonLdapConfig.Password == "" {
		jsonLdapConfig.Password = "secret"
	}
	if jsonLdapConfig.Namespace == "" {
		jsonLdapConfig.Namespace = "dc=local"
	}
	if jsonLdapConfig.OrgaName == "" {
		jsonLdapConfig.OrgaName = "test"
	}

	return jsonLdapConfig
}

func BindConnect(hostname string, port int, binddn, password string) error {

	// already connected
	if ldapCon != nil {
		err := errors.New(fmt.Sprintf("Already connected"))
		logging.Info("BindConnect", err.Error())
		// yeah, we return nil here, because we would like connect, and we are connected
		return nil
	}

	// connect
	tempConnection, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		disconnect()
		logging.Error("BindConnect", err.Error())
		return err
	}
	ldapCon = tempConnection
	logging.Info("BindConnect", "Connected")

	// authenticate
	err = ldapCon.Bind(binddn, password)
	if err != nil {
		disconnect()
		logging.Error("BindConnect", err.Error())
		return err
	}

	ldapConnected = true

	return nil
}

func disconnect() {
	if ldapCon != nil {
		ldapCon.Close()
	}
	ldapCon = nil
	ldapConnected = false

	logging.Info("disconnect", "Disconnected")
}

func PopulateDirectory(namespace, orgaName string) {
	// If we are not connected, we dont do anything
	if ldapCon == nil {
		return
	}

	// maybe the orga changed
	ldapOrgaObj = organizationInit(namespace, orgaName)
	ldapOrgaObj.Add(ldapCon)

	ldapDummy := inetOrgPersonInit(ldapOrgaObj.Dn, "dummy", "dummy", "Default")
	ldapDummy.Add(ldapCon)
	ldapDummy.ToJsonString()

	ldapGroupsFolder := organizationalUnitInit(ldapOrgaObj.Dn, "groups", "The Folder for all groups")
	ldapGroupsFolder.Add(ldapCon)

	ldapUsersFolder := organizationalUnitInit(ldapOrgaObj.Dn, "users", "The Folder for all users")
	ldapUsersFolder.Add(ldapCon)

	ldapNextcloudGroup := groupOfNamesInit(ldapGroupsFolder.Dn, "nextcloud", ldapDummy.Dn)
	ldapNextcloudGroup.Add(ldapCon)

	ldapAdminUser := inetOrgPersonInit(ldapUsersFolder.Dn, "admin", "admin", "The")
	ldapAdminUser.Add(ldapCon)
}

func OrganisationAdd(basedn, name string) error {

	searchRequest := ldap.NewSearchRequest(
		basedn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(dc="+name+")(objectClass=organization))",
		[]string{"dn"},
		nil,
	)

	sr, err := ldapCon.Search(searchRequest)
	if err == nil {
		if len(sr.Entries) > 0 {
			logging.Debug("OrganisationAdd - Check", fmt.Sprintf("Organisation '%s' already exist, thats okay", name))
			return nil
		}
	} else {
		return err
	}

	addReq := ldap.NewAddRequest(basedn)
	addReq.Attribute("objectClass", []string{"top", "dcObject", "organization"})
	addReq.Attribute("dc", []string{name})
	addReq.Attribute("o", []string{name})
	err = ldapCon.Add(addReq)
	if err != nil {
		return err
	}

	logging.Info("OrganisationAdd - Add", fmt.Sprintf("Success for '%s'", name))
	return nil
}

func AddUserToGroup(groupdn, userdn string) error {

	// search for group
	searchRequest := ldap.NewSearchRequest(
		groupdn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=groupOfNames))",
		[]string{"dn"},
		nil,
	)

	sr, err := ldapCon.Search(searchRequest)
	if err == nil {
		if len(sr.Entries) == 0 {
			logging.Error("AddUserToGroup - Check", "Group dont exist, can not add user to group")
			return errors.New("Group dont exist, can not add user to group")
		}
	}

	// search for user
	searchRequest = ldap.NewSearchRequest(
		userdn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=inetOrgPerson))",
		[]string{"dn"},
		nil,
	)

	sr, err = ldapCon.Search(searchRequest)
	if err == nil {
		if len(sr.Entries) == 0 {
			logging.Error("AddUserToGroup - Check", "User dont exist, can not add user to group")
			return errors.New("User dont exist, can not add user to group")
		}
	}

	// Add a description, and replace the mail attributes
	modify := ldap.NewModifyRequest(userdn)
	modify.Add("memberOf", []string{groupdn})

	err = ldapCon.Modify(modify)
	if err != nil {
		logging.Error("AddUserToGroup", err.Error())
		return err
	}

	return nil
}

type seachResultFct func(string /*basedn*/, string /*dn*/, string /*type*/, string /*displayName*/)
type searchResultObjectFct func(object ldapObject)

func GetObjectAsJsonString(fulldn string) string {

	searchRequest := ldap.NewSearchRequest(
		fulldn, // The base dn to search
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(|(objectClass=organizationalUnit)(objectClass=inetOrgPerson)(objectClass=groupOfNames))",
		/* "(&(dn="+fulldn+"))", */
		[]string{"objectClass", "ou", "sn", "cn", "uid", "description", "memberOf", "userPassword", "displayName"}, // A list attributes to retrieve
		nil,
	)

	sr, err := ldapCon.Search(searchRequest)
	if err != nil {
		logging.Error("GetObject", fmt.Sprintf("[%s] %s", fulldn, err.Error()))
		return ""
	}

	if len(sr.Entries) <= 0 {
		logging.Debug("GetObject", fmt.Sprintf("[%s] No entry found", fulldn))
		return ""
	}

	entry := sr.Entries[0]
	entry.PrettyPrint(2)
	/*
		newObj := ldapObject{
			Dn:           entry.DN,
			ObjectClass:  entry.GetAttributeValue("objectClass"),
			Cn:           entry.GetAttributeValue("cn"),
			Ou:           entry.GetAttributeValue("ou"),
			Sn:           entry.GetAttributeValue("sn"),
			Uid:          entry.GetAttributeValue("uid"),
			Description:  entry.GetAttributeValue("description"),
			UserPassword: entry.GetAttributeValue("userPassword"),
			MemberOf:     entry.GetAttributeValues("memberOf"),
		}

		groupObjectBytes, err := json.Marshal(newObj)
		if err != nil {
			fmt.Println("error:", err)
			return ""
		}

		return string(groupObjectBytes)
	*/
	return ""
}

func onMessage(message *msgbus.Msg, group, command, payload string) {

	if command == "getConfig" {

		jsonLdapConfig := GetLdapConfig()

		groupObjectBytes, err := json.Marshal(jsonLdapConfig)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		message.Answer(&plugin, "config", string(groupObjectBytes))
		return
	}

	if command == "saveConfig" {

		var jsonNewConfig map[string]interface{}
		err := json.Unmarshal([]byte(payload), &jsonNewConfig)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		core.SetJsonObject("ldap", jsonNewConfig)
		core.ConfigSave()
		message.Answer(&plugin, "configSaved", "")
		return
	}

	if command == "connect" {

		// read config-object from config-file
		jsonLdapConfig := GetLdapConfig()

		// try to connect
		err := BindConnect(jsonLdapConfig.Host, int(jsonLdapConfig.Port), jsonLdapConfig.BindDN, jsonLdapConfig.Password)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// Populate directory
		PopulateDirectory(jsonLdapConfig.Namespace, jsonLdapConfig.OrgaName)

		message.Answer(&plugin, "connected", "")
		return
	}

	if command == "disconnect" {
		disconnect()
		message.Answer(&plugin, "disconnected", "No config")
		return
	}

	if command == "isConnected" {
		if ldapConnected == true {
			message.Answer(&plugin, "connected", "")
		} else {
			message.Answer(&plugin, "disconnected", "No config")
		}
	}

	if command == "getObjects" {

		// read config-object from config-file
		jsonLdapConfig := GetLdapConfig()

		// try to connect
		err := BindConnect(jsonLdapConfig.Host, int(jsonLdapConfig.Port), jsonLdapConfig.BindDN, jsonLdapConfig.Password)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// Populate directory
		PopulateDirectory(jsonLdapConfig.Namespace, jsonLdapConfig.OrgaName)

		// disconnect on return
		defer disconnect()

		tempSearchDN := ldapOrgaObj.Dn
		if payload != "" {
			tempSearchDN = payload
		} else {

			oldBaseDn := ldapOrgaObj.DnBase

			ldapOrgaObj.DnBase = "#"
			message.Answer(&plugin, "objects", ldapOrgaObj.ToJsonString())
			message.Answer(&plugin, "objectsFinish", tempSearchDN)

			ldapOrgaObj.DnBase = oldBaseDn

			return
		}

		SearchAllFull(ldapCon, tempSearchDN, func(entry *ldap.Entry) {

			// create an empty object
			ldapObject := ldapObjectCreate([]string{""}, tempSearchDN, "", "")

			// the core attributes
			ldapObject.ObjectClass = entry.GetAttributeValues("objectClass")
			ldapObject.Dn = entry.DN

			for _, attribute := range entry.Attributes {
				ldapObject.SetMustAttr(attribute.Name, attribute.Values)
			}

			message.Answer(&plugin, "objects", ldapObject.ToJsonString())
		})
		message.Answer(&plugin, "objectsFinish", tempSearchDN)

		return
	}

	if command == "getObject" {

		// read config-object from config-file
		jsonLdapConfig := GetLdapConfig()

		// try to connect
		err := BindConnect(jsonLdapConfig.Host, int(jsonLdapConfig.Port), jsonLdapConfig.BindDN, jsonLdapConfig.Password)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// Populate directory
		PopulateDirectory(jsonLdapConfig.Namespace, jsonLdapConfig.OrgaName)

		// disconnect on return
		defer disconnect()

		SearchOneFull(ldapCon, payload, func(entry *ldap.Entry) {

			// create an empty object
			ldapObject := ldapObjectCreate([]string{""}, "", "", "")

			// the core attributes
			ldapObject.ObjectClass = entry.GetAttributeValues("objectClass")
			ldapObject.Dn = entry.DN

			for _, attribute := range entry.Attributes {
				ldapObject.SetMustAttr(attribute.Name, attribute.Values)
			}

			message.Answer(&plugin, "object", ldapObject.ToJsonString())
		})

		return
	}

	if command == "getTemplate" {

		var object ldapObject
		if payload == "organizationalUnit" {
			object = organizationalUnitInit("", "", "")
		}
		if payload == "groupOfNames" {
			object = groupOfNamesInit("", "", "")
		}
		if payload == "inetOrgPerson" {
			object = inetOrgPersonInit("", "", "", "")
		}

		// convert to json
		var newJson = object.ToTemplate()
		groupObjectBytes, err := json.Marshal(&newJson)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		message.Answer(&plugin, "template", string(groupObjectBytes))
	}

	if command == "newObject" {

		// parse json
		var jsonNewConfig map[string]interface{}
		err := json.Unmarshal([]byte(payload), &jsonNewConfig)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

	}

	if command == "modifyObject" {

		// parse json
		var changeObjct ldapChangeRequest
		err := json.Unmarshal([]byte(payload), &changeObjct)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// check if dn exist
		if changeObjct.Dn == "" {
			message.Answer(&plugin, "error", "DN is missing in object")
			return
		}

		// create an empty object
		ldapObject := ldapObjectCreate([]string{""}, "", "", "")

		// the core attributes
		ldapObject.Dn = changeObjct.Dn

		// set all attr
		for attrName, attrValues := range changeObjct.Attr {
			ldapObject.SetMustAttr(attrName, attrValues)
		}

		// read config-object from config-file
		jsonLdapConfig := GetLdapConfig()

		// try to connect
		err = BindConnect(jsonLdapConfig.Host, int(jsonLdapConfig.Port), jsonLdapConfig.BindDN, jsonLdapConfig.Password)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// disconnect on return
		defer disconnect()

		// send change request
		err = ldapObject.Change(ldapCon)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		message.Answer(&plugin, "modifyed", ldapObject.Dn)
		return
	}

}
