/*
Copyright (C) 2018 by Martin Langlotz aka stackshadow

This file is part of gopilot, an rewrite of the copilot-project in go

gopilot is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, version 3 of this License

gopilot is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gopilot.  If not, see <http://www.gnu.org/licenses/>.
*/

package ctls

/*
## Key considerations for algorithm "RSA" ≥ 2048-bit
openssl genrsa -out server.key 2048

# Key considerations for algorithm "ECDSA" ≥ secp384r1
# List ECDSA the supported curves (openssl ecparam -list_curves)
openssl ecparam -genkey -name secp384r1 -out server.key

openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

Testclient:
openssl s_client -connect localhost:4443 -key server.key -cert server.crt

Show info:
openssl x509 -noout -text -in cerfile.cer

*/
import (
	"core/clog"
	"core/core"
	"core/msgbus"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type pluginCtls struct {
	plugin    msgbus.Plugin
	sessionNo int
}

// options
var newNode string        // create a new node with a new secret
var serverAdress string   // set server of this node
var remoteNodeName string // connect to remote node
var remoteNodeHost string // the remote host
var remoteAcceptNode string

var logging clog.Logger

func ParseCmdLine() {
	flag.StringVar(&serverAdress, "serverAdress", "", "hostname:port - Enable the TLS-Server on hostname with port")
	flag.StringVar(&newNode, "newNode", "", "name - Use this on the server to allow an node for an incoming connection")
	flag.StringVar(&remoteNodeName, "remoteNodeName", "", "name - Connect to an remote node, this need also remoteNodeHost")
	flag.StringVar(&remoteNodeHost, "remoteNodeHost", "", "hostname:port - Connection information for remote node")
	flag.StringVar(&remoteAcceptNode, "acceptNode", "", "nodename - Accept an hash-request")
}

func Init() pluginCtls {

	var newCtls pluginCtls

	logging = clog.New("TLS")

	// create key pair
	CreateKeyPair(core.NodeName)

	// enable the tls-server
	if serverAdress != "" {

		host, portString, err := net.SplitHostPort(serverAdress)
		if err != nil {
			logging.Error("serverAdress", fmt.Sprintf("remoteNode '%s': ", serverAdress, err))
			os.Exit(-1)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			logging.Error("serverAdress", fmt.Sprintf("Can not convert port-string to int: ", err))
			os.Exit(-1)
		}

		logging.Info("serverAdress", fmt.Sprintf("Serve an TLS-Server on '%s': ", serverAdress))

		core.SetNode(core.NodeName, core.NodeTypeServer, host, port)
		core.ConfigSave()
		os.Exit(0)
	}

	// create a newNode-Config
	if newNode != "" {
		core.DeleteNode(newNode)
		core.SetNode(newNode, core.NodeTypeIncoming, "", 0)
		core.ConfigSave()
		os.Exit(0)
	}

	// we connect to an remote-node
	if remoteNodeName != "" && remoteNodeHost != "" {

		// get host and port
		host, portString, err := net.SplitHostPort(remoteNodeHost)
		if err != nil {
			logging.Error("NODE", fmt.Sprintf("remoteNode '%s': ", remoteNodeHost, err))
			os.Exit(-1)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			logging.Error("NODE", fmt.Sprintf("Can not convert port-string to int: ", err))
			os.Exit(-1)
		}

		// set nodeName
		core.SetNode(remoteNodeName, core.NodeTypeClient, host, port)
		nodeObject := core.GetNodeObject(remoteNodeName)
		if nodeObject == nil {
			os.Exit(-1)
		}

		os.Exit(0)
	}

	// we accept an requested node
	if remoteAcceptNode != "" {
		peerCertAcceptReqCert(remoteAcceptNode)
		os.Exit(0)
	}

	// register plugin on messagebus
	newCtls.plugin = msgbus.NewPlugin("TLS")
	newCtls.plugin.Register()
	newCtls.plugin.ListenForGroup("tls", newCtls.onMessage)

	// okay, get server-config
	core.IterateNodes(func(nodeName string, nodeType int, host string, port int) {

		if nodeType == core.NodeTypeServer {
			go newCtls.serve(fmt.Sprintf("%s:%d", host, port))
		}

		if nodeType == core.NodeTypeClient {
			go newCtls.connect(fmt.Sprintf("%s:%d", host, port))
		}
	})

	return newCtls
}

func CreateKeyPair(nodeName string) error {

	if _, err := os.Stat(nodeName + ".key"); os.IsNotExist(err) {
		cmdKey := exec.Command("openssl", "ecparam", "-genkey", "-name", "secp384r1",
			"-out", nodeName+".key",
		)
		logging.Error("CREATEKEY", fmt.Sprintf("Create %s.key", nodeName))
		errKey := cmdKey.Run()
		if errKey != nil {
			logging.Error("CREATEKEY", fmt.Sprintf("%s", errKey))
			return errKey
		}
	}

	if _, err := os.Stat(nodeName + ".crt"); os.IsNotExist(err) {
		cmdCRT := exec.Command("openssl", "req", "-new", "-x509", "-sha256",
			"-key", nodeName+".key",
			"-out", nodeName+".crt",
			"-days", "3650",
			"-subj", "/C=DE/ST=UNKNOWN/L=UNKNOWN/O=COPILOTD/OU=DAEMON/CN="+nodeName,
		)
		logging.Error("CREATEKEY", fmt.Sprintf("Create %s.crt with %s", nodeName, cmdCRT.Args))
		errCRT := cmdCRT.Run()
		if errCRT != nil {
			logging.Error("CREATEKEY", fmt.Sprintf("%s", errCRT))
			return errCRT
		}
	}

	return nil
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Accept requested Cert for an node
func peerCertAcceptReqCert(nodeName string) error {

	nodeObject := core.GetNodeObject(nodeName)
	if nodeObject == nil {
		logging.Error("CLIENT", fmt.Sprintf(
			"Node '%s' not exist in config",
			nodeName,
		))
		return errors.New("Node not exist in config")
	}

	// already exist, do nothing
	if (*nodeObject)["peerCertSignature"] != nil {
		logging.Error("CLIENT", fmt.Sprintf(
			"Can not overwrite an already accepted key",
		))
		return errors.New("Can not overwrite an already accepted key")
	}

	// no req-key exist, do nothing
	if (*nodeObject)["peerCertReqSignature"] == nil {
		logging.Error("CLIENT", fmt.Sprintf(
			"No key requested",
		))
		return errors.New("No key requested")
	}

	(*nodeObject)["peerCertSignature"] = (*nodeObject)["peerCertReqSignature"]
	delete(*nodeObject, "peerCertReqSignature")
	core.ConfigSave()

	logging.Info("CLIENT", fmt.Sprintf(
		"Accept requested key for node",
	))

	return nil
}

func (curCtls *pluginCtls) serve(serverString string) {

	cer, err := tls.LoadX509KeyPair(core.NodeName+".crt", core.NodeName+".key")
	if err != nil {
		logging.Error("SERVER", err.Error())
		return
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cer},
		ClientAuth:   tls.RequireAnyClientCert,
	}

	logging.Info("SERVER", fmt.Sprintf("Start serve on %s", serverString))

	ln, err := tls.Listen("tcp", serverString, config)
	if err != nil {
		logging.Error("SERVER", err.Error())
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			logging.Error("SERVER", err.Error())
			continue
		}

		go NewSession(
			fmt.Sprintf("%d", curCtls.sessionNo),
			core.NodeTypeIncoming, conn,
		)

		curCtls.sessionNo++
	}

}

func (curCtls *pluginCtls) connect(clientString string) {

	cer, err := tls.LoadX509KeyPair(core.NodeName+".crt", core.NodeName+".key")
	if err != nil {
		logging.Error("CONNECT", err.Error())
		return
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{cer},
		ClientAuth:         tls.RequireAnyClientCert,
		InsecureSkipVerify: true,
	}

	for {
		logging.Info("CONNECT", fmt.Sprintf("Try to connect to %s", clientString))
		conn, err := tls.Dial("tcp", clientString, config)
		if err != nil {
			logging.Error("CONNECT", fmt.Sprintf("Failed to connect: %s", err.Error()))
			time.Sleep(time.Second * 10)
			continue
		}

		NewSession(
			fmt.Sprintf("%d", curCtls.sessionNo),
			core.NodeTypeClient, conn,
		)

		conn.Close()
		time.Sleep(time.Second * 10)
	}

}

func (curCtls *pluginCtls) onMessage(message *msgbus.Msg, group, command, payload string) {

	if command == "nodeAccept" {
		err := peerCertAcceptReqCert(payload)
		if err == nil {
			message.Answer(&curCtls.plugin, "nodeAcceptOk", payload)
			return
		} else {
			message.Answer(&curCtls.plugin, "error", err.Error())
			return
		}
	}

}
