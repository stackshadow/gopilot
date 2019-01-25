package ldapclient

import (
	"core/clog"
	"core/msgbus"
	"plugins/core"

	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"

	"gopkg.in/ldap.v3"
)

// private vars
var plugin msgbus.Plugin
var logging clog.Logger
var ldapCon *ldap.Conn = nil

var ldapConnected bool
var ldapNamespace string
var ldapBaseOrgaName string
var ldapOrgaObj *ldapObject

var ldapUserSuffixDn string
var ldapGroupSuffixDn string
var ldapDummyUserDn string

type ldapConnectionConfig struct {
	Host      string  `json:"host"`
	Port      float64 `json:"port"`
	BindDN    string  `json:"binddn"`
	Password  string  `json:"password"`
	Namespace string  `json:"namespace"`
	OrgaName  string  `json:"organame"`
}

type ldapChangeRequest struct {
	Dn          string              `json:"dn"`
	ObjectClass []string            `json:"objectClass"`
	AttrData    map[string][]string `json:"attrData"`
}

type ldapCreateRequest struct {
	DnBase      string              `json:"basedn"`
	ObjectClass []string            `json:"objectClass"`
	AttrData    map[string][]string `json:"attrData"`
}

func ParseCmdLine() {
	flag.StringVar(&ldapNamespace, "ldapNamespace", "dc=local", "The namespace where your organisation lives")
	flag.StringVar(&ldapBaseOrgaName, "ldapOrga", "shinyneworga", "Name of your organisation")
}

func Init() {

	logging = clog.New("LDAP")

	// we init the global class storage
	ldapClassInit()

	// we create for every type a class to create the objectClass and attributes search strings
	ldapClassOrganizationRegister()
	ldapClassOrganizationalUnitRegister()
	ldapClassInetOrgPersonRegister()
	ldapClassgroupOfNamesRegister()

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

func SetLdapConfig(newConfig ldapConnectionConfig) error {

	// to json
	groupObjectBytes, err := json.Marshal(newConfig)
	if err != nil {
		fmt.Println("LDAP:SetLdapConfig", err)
		return err
	}

	// from json
	var jsonObject map[string]interface{}
	err = json.Unmarshal(groupObjectBytes, &jsonObject)
	if err != nil {
		logging.Error("LDAP:SetLdapConfig", err.Error())
		return err
	}

	// save it
	core.SetJsonObject("ldap", jsonObject)
	core.ConfigSave()
	return nil
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
	_, ldapOrgaObj = organizationCreate(namespace, orgaName)
	if ldapOrgaObj != nil {
		ldapOrgaObj.Add(ldapCon)
	}

	_, ldapDummy := inetOrgPersonCreate(ldapOrgaObj.Dn, "dummy", "dummy", "Default")
	ldapDummy.Add(ldapCon)
	ldapDummy.ToJsonString()

	_, ldapGroupsFolder := organizationalUnitCreate(ldapOrgaObj.Dn, "groups", "The Folder for all groups")
	ldapGroupsFolder.Add(ldapCon)

	_, ldapUsersFolder := organizationalUnitCreate(ldapOrgaObj.Dn, "users", "The Folder for all users")
	ldapUsersFolder.Add(ldapCon)

	_, ldapNextcloudGroup := groupOfNamesCreate(ldapGroupsFolder.Dn, "nextcloud", ldapDummy.Dn)
	ldapNextcloudGroup.Add(ldapCon)

	_, ldapAdminUser := inetOrgPersonCreate(ldapUsersFolder.Dn, "admin", "admin", "The")
	ldapAdminUser.Add(ldapCon)
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
	modify := ldap.NewModifyRequest(userdn, nil)
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

		var jsonLdapConfig ldapConnectionConfig = GetLdapConfig()

		// we protect our password !
		jsonLdapConfig.Password = ""

		groupObjectBytes, err := json.Marshal(jsonLdapConfig)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		message.Answer(&plugin, "config", string(groupObjectBytes))
		return
	}

	if command == "saveConfig" {

		// get the json
		var configValues ldapConnectionConfig
		err := json.Unmarshal([]byte(payload), &configValues)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// get the original config
		var jsonLdapConfig ldapConnectionConfig = GetLdapConfig()

		// patch it
		if configValues.Host != "" {
			jsonLdapConfig.Host = configValues.Host
		}
		if configValues.Port != 0 {
			jsonLdapConfig.Port = configValues.Port
		}
		if configValues.BindDN != "" {
			jsonLdapConfig.BindDN = configValues.BindDN
		}
		if configValues.Password != "" {
			jsonLdapConfig.Password = configValues.Password
		}
		if configValues.Namespace != "" {
			jsonLdapConfig.Namespace = configValues.Namespace
		}
		if configValues.OrgaName != "" {
			jsonLdapConfig.OrgaName = configValues.OrgaName
		}

		SetLdapConfig(jsonLdapConfig)
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

		// we want the base-dn
		if payload == "" {

			oldBaseDn := ldapOrgaObj.DnBase

			ldapOrgaObj.DnBase = "#"
			message.Answer(&plugin, "objects", ldapOrgaObj.ToJsonString())
			message.Answer(&plugin, "objectsFinish", payload)

			ldapOrgaObj.DnBase = oldBaseDn

			return
		}

		SearchAllFull(ldapCon, payload, func(entry *ldap.Entry) {

			// get the object of the corresponding class
			objectClass := entry.GetAttributeValues("objectClass")
			err, ldapObject := ldapClassCreateLdapObject(objectClass)
			if err != nil {
				message.Answer(&plugin, "error", err.Error())
				return
			}

			// set the dn
			ldapObject.DnBase = payload

			// set all readed attributes
			for _, attribute := range entry.Attributes {

				// ignore objectClass
				if attribute.Name == "objectClass" {
					continue
				}

				ldapObject.SetAttrValue(attribute.Name, attribute.Values)
			}

			message.Answer(&plugin, "objects", ldapObject.ToJsonString())
		})
		message.Answer(&plugin, "objectsFinish", payload)

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

		// get ldapObject from fulldn
		err, ldapObject := GetLdapObject(ldapCon, payload)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		message.Answer(&plugin, "object", ldapObject.ToJsonString())
		return
	}

	if command == "getTemplate" {

		var ldapClass []string
		err := json.Unmarshal([]byte(payload), &ldapClass)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// we create an ldapObject from class to get an template
		err, ldapTemplateObject := ldapClassCreateLdapObject(ldapClass)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		message.Answer(&plugin, "template", ldapTemplateObject.ToJsonString())
	}

	if command == "createObject" {

		// parse json
		var newObject ldapCreateRequest
		err := json.Unmarshal([]byte(payload), &newObject)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// check if dn exist
		if newObject.DnBase == "" {
			message.Answer(&plugin, "error", "DnBase is missing in object")
			return
		}
		// check if objectClass exist
		if newObject.ObjectClass == nil {
			message.Answer(&plugin, "error", "objectClass is missing in object")
			return
		}

		// create an empty object
		err, ldapObjectToCreate := ldapClassCreateLdapObject(newObject.ObjectClass)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// set base-dn
		ldapObjectToCreate.DnBase = newObject.DnBase

		// set all attributes that should be changed
		for attrName, attrValue := range newObject.AttrData {
			err = ldapObjectToCreate.SetAttrValue(attrName, attrValue)
			if err != nil {
				message.Answer(&plugin, "error", err.Error())
				return
			}
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
		err = ldapObjectToCreate.Add(ldapCon)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		message.Answer(&plugin, "created", ldapObjectToCreate.Dn)
		return
	}

	if command == "modifyObject" {

		// parse json
		var changeRequestObject ldapChangeRequest
		err := json.Unmarshal([]byte(payload), &changeRequestObject)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// check if dn exist
		if changeRequestObject.Dn == "" {
			message.Answer(&plugin, "error", "DN is missing in object")
			return
		}
		// check if objectClass exist
		if changeRequestObject.ObjectClass == nil {
			message.Answer(&plugin, "error", "objectClass is missing in object")
			return
		}

		// create an empty object
		err, ldapObjectToChange := ldapClassCreateLdapObject(changeRequestObject.ObjectClass)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// we need to grab the base dn from the dn
		splitDN := strings.SplitN(changeRequestObject.Dn, ",", 2)
		if len(splitDN) <= 1 {
			message.Answer(&plugin, "error", "Could not read basedn from dn")
			return
		}
		ldapObjectToChange.DnBase = splitDN[1]

		// set all attributes that should be changed
		for attrName, attrValue := range changeRequestObject.AttrData {
			err = ldapObjectToChange.SetAttrValue(attrName, attrValue)
			if err != nil {
				message.Answer(&plugin, "error", err.Error())
				return
			}
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

		// if DN was not changed by mainAttr, we use the dn from the change request
		if ldapObjectToChange.Dn == "" {
			ldapObjectToChange.Dn = changeRequestObject.Dn
		}

		// compare if dn is changed
		if ldapObjectToChange.Dn != changeRequestObject.Dn {
			ldapObjectToChange.Rename(ldapCon, changeRequestObject.Dn)
		}

		// send change request
		err = ldapObjectToChange.Change(ldapCon)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		message.Answer(&plugin, "modifyed", ldapObjectToChange.Dn)
		return
	}

	if command == "deleteObject" {

		var classes []string
		ldapObject := ldapObjectCreate(classes, "", "")
		ldapObject.Dn = payload

		// read config-object from config-file
		jsonLdapConfig := GetLdapConfig()
		// try to connect
		err := BindConnect(jsonLdapConfig.Host, int(jsonLdapConfig.Port), jsonLdapConfig.BindDN, jsonLdapConfig.Password)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}
		// disconnect on return
		defer disconnect()

		// remove
		err = ldapObject.Remove(ldapCon)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		message.Answer(&plugin, "deleted", ldapObject.Dn)
		return
	}
}
