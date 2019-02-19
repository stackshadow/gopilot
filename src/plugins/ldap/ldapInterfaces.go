package ldapclient

//import "errors"
import (
	"gopkg.in/ldap.v3"
)

type ILdap interface {
	NewSearchRequest(
		BaseDN string,
		Scope, DerefAliases, SizeLimit, TimeLimit int,
		TypesOnly bool,
		Filter string,
		Attributes []string,
		Controls []ldap.Control,
	) *ldap.SearchRequest
}

type ILdapConnection interface {
	Bind(username, password string) error
	Close()
	Add(addRequest *ldap.AddRequest) error
	Modify(modifyRequest *ldap.ModifyRequest) error
	ModifyDN(m *ldap.ModifyDNRequest) error
	Del(delRequest *ldap.DelRequest) error
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
}
