package pluginLdap

import (
	"gopkg.in/ldap.v3"
)

type gopilotLdapConnectionMocked struct {
	entries []*ldap.Entry
}

func (con *gopilotLdapConnectionMocked) Bind(username, password string) error {
	con.entries = make([]*ldap.Entry, 0)
	return nil
}
func (con *gopilotLdapConnectionMocked) Close() {

}
func (con *gopilotLdapConnectionMocked) Add(addRequest *ldap.AddRequest) error {

	var newEntry ldap.Entry
	newEntry.DN = addRequest.DN

	attributeArray := make([]*ldap.EntryAttribute, 0)

	for _, attribute := range addRequest.Attributes {
		var newEntryAttribute ldap.EntryAttribute
		newEntryAttribute.Name = attribute.Type
		newEntryAttribute.Values = attribute.Vals
		attributeArray = append(attributeArray, &newEntryAttribute)
	}

	newEntry.Attributes = attributeArray
	con.entries = append(con.entries, &newEntry)

	return nil
}
func (con *gopilotLdapConnectionMocked) Modify(modifyRequest *ldap.ModifyRequest) error {
	return nil
}
func (con *gopilotLdapConnectionMocked) ModifyDN(m *ldap.ModifyDNRequest) error {
	return nil
}
func (con *gopilotLdapConnectionMocked) Del(delRequest *ldap.DelRequest) error {
	return nil
}
func (con *gopilotLdapConnectionMocked) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	var newRequest ldap.SearchResult

	newRequest.Entries = con.entries

	return &newRequest, nil
}
