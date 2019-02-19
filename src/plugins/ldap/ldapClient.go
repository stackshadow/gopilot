package ldapclient

//import "errors"
import (
	"errors"
	"fmt"
	"gopkg.in/ldap.v3"
)

type SLdapClient struct {
	conn        ILdapConnection
	isConnected bool

	baseDn string
}

func ldapClientNew() SLdapClient {
	var newClient SLdapClient
	return newClient
}

func (client *SLdapClient) BindConnect(hostname string, port int, binddn, password string) error {

	// already connected
	if client.isConnected == true {
		err := errors.New(fmt.Sprintf("Already connected"))
		logging.Info("BindConnect", err.Error())
		// yeah, we return nil here, because we would like connect, and we are connected
		return nil
	}

	var err error

	// connect
	client.conn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		client.Disconnect()
		logging.Error("BindConnect", err.Error())
		return err
	}
	logging.Info("BindConnect", "Connected")

	// authenticate
	err = client.conn.Bind(binddn, password)
	if err != nil {
		client.Disconnect()
		logging.Error("BindConnect", err.Error())
		return err
	}

	client.isConnected = true

	return nil
}

func (client *SLdapClient) Disconnect() {
	if client.conn != nil {
		client.conn.Close()
	}
	client.conn = nil
	client.isConnected = false

	logging.Info("disconnect", "Disconnected")
}
