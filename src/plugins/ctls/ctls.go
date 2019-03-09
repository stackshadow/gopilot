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
	"core/config"
	"core/msgbus"
	"core/nodes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// options
var newNode string        // create a new node with a new secret
var serverAdress string   // set server of this node
var remoteNodeName string // connect to remote node
var remoteNodeHost string // the remote host
var remoteAcceptNode string
var remoteRejectNode string // we will forget for this nodeName the sharedSecret, and TLS-Keys

// private vars
var plugin msgbus.Plugin
var logging clog.Logger
var sessionNo int

func ParseCmdLine() {
	flag.StringVar(&serverAdress, "serverAdress", "", "hostname:port - Enable the TLS-Server on hostname with port")
	flag.StringVar(&newNode, "newNode", "", "name - Use this on the server to allow an node for an incoming connection")
	flag.StringVar(&remoteNodeName, "remoteNodeName", "", "name - Connect to an remote node, this need also remoteNodeHost")
	flag.StringVar(&remoteNodeHost, "remoteNodeHost", "", "hostname:port - Connection information for remote node")
	flag.StringVar(&remoteAcceptNode, "acceptNode", "", "nodename - Accept an hash-request")
	flag.StringVar(&remoteRejectNode, "rejectNode", "", "nodename - We forget all keys and secrets for this nodeName")
}

func Init() {

	logging = clog.New("TLS")

	// create key pair
	CreateKeyPair(config.NodeName)

	// enable the tls-server
	if serverAdress != "" {

		host, portString, err := net.SplitHostPort(serverAdress)
		if err != nil {
			logging.Error("serverAdress", fmt.Sprintf("remoteNode '%s': %s", serverAdress, err.Error()))
			os.Exit(-1)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			logging.Error("serverAdress", fmt.Sprintf("Can not convert port-string to int: %s", err.Error()))
			os.Exit(-1)
		}

		logging.Info("serverAdress", fmt.Sprintf("Serve an TLS-Server on '%s': ", serverAdress))

		nodes.SaveData(config.NodeName, nodes.NodeTypeServer, host, port)
		os.Exit(0)
	}

	// create a newNode-Config
	if newNode != "" {
		nodes.SaveData(newNode, nodes.NodeTypeIncoming, "", 0)
		os.Exit(0)
	}

	// we connect to an remote-node
	if remoteNodeName != "" && remoteNodeHost != "" {

		// get host and port
		host, portString, err := net.SplitHostPort(remoteNodeHost)
		if err != nil {
			logging.Error("NODE", fmt.Sprintf("remoteNode '%s': %s", remoteNodeHost, err.Error()))
			os.Exit(-1)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			logging.Error("NODE", fmt.Sprintf("Can not convert port-string to int: %s", err.Error()))
			os.Exit(-1)
		}

		// set nodeName
		nodes.SaveData(remoteNodeName, nodes.NodeTypeClient, host, port)
		os.Exit(0)
	}

	// we accept an requested node
	if remoteAcceptNode != "" {
		peerCertAcceptReqCert(remoteAcceptNode)
		os.Exit(0)
	}

	// we forget all secrets for this node
	if remoteRejectNode != "" {
		peerCertReject(remoteRejectNode)
		os.Exit(0)
	}

	// register plugin on messagebus
	plugin = msgbus.NewPlugin("TLS")
	plugin.Register()
	plugin.ListenForGroup("tls", onMessage)

	// okay, get server-config
	nodes.IterateNodes(func(jsonNode nodes.JSONNodeType, nodeName string, nodeType int, host string, port int) {

		if nodeType == nodes.NodeTypeServer {
			go serve(fmt.Sprintf("%s:%d", host, port))
		}

		if nodeType == nodes.NodeTypeClient {
			go connect(fmt.Sprintf("%s:%d", host, port))
		}
	})

}

func CreateKeyPair(nodeName string) error {

	if _, err := os.Stat(nodeName + ".key"); os.IsNotExist(err) {
		cmdKey := exec.Command("openssl", "ecparam", "-genkey", "-name", "secp384r1",
			"-out", nodeName+".key",
		)
		logging.Info("CREATEKEY", fmt.Sprintf("Create %s.key", nodeName))
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
		logging.Info("CREATEKEY", fmt.Sprintf("Create %s.crt with %s", nodeName, cmdCRT.Args))
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

	nodeObject, err := nodes.GetNodeObject(nodeName)
	if err != nil {
		logging.Error("CLIENT", err.Error())
		return err
	}

	// already exist, do nothing
	if nodeObject["peerCertSignature"] != nil {
		logging.Error("CLIENT", fmt.Sprintf(
			"Can not overwrite an already accepted key",
		))
		return errors.New("Can not overwrite an already accepted key")
	}

	// no req-key exist, do nothing
	if nodeObject["peerCertSignatureReq"] == nil {
		logging.Error("CLIENT", fmt.Sprintf(
			"No key requested",
		))
		return errors.New("No key requested")
	}

	// set the peer
	nodeObject["peerCertSignature"] = nodeObject["peerCertSignatureReq"]
	delete(nodeObject, "peerCertSignatureReq")
	nodes.SaveNodeObject(nodeName, nodeObject)

	logging.Info("CLIENT", fmt.Sprintf("Accept requested key for node"))

	return nil
}

// delete Certificate-Signatures and shared secret for this node
func peerCertReject(nodeName string) error {

	nodeObject, err := nodes.GetNodeObject(nodeName)
	if err != nil {
		logging.Error("CLIENT", err.Error())
		return err
	}

	delete(nodeObject, "peerCertSignature")
	delete(nodeObject, "peerCertSignatureReq")
	delete(nodeObject, "sharedSecret")

	logging.Info("CLIENT", fmt.Sprintf("Remove all keys for '%s'", nodeName))
	nodes.SaveNodeObject(nodeName, nodeObject)

	return nil
}

func serve(serverString string) {

	cer, err := tls.LoadX509KeyPair(config.NodeName+".crt", config.NodeName+".key")
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
			fmt.Sprintf("%d", sessionNo),
			nodes.NodeTypeIncoming, conn,
		)

		sessionNo++
	}

}

func connect(clientString string) {

	cer, err := tls.LoadX509KeyPair(config.NodeName+".crt", config.NodeName+".key")
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
			fmt.Sprintf("%d", sessionNo),
			nodes.NodeTypeClient, conn,
		)

		conn.Close()
		time.Sleep(time.Second * 10)
	}

}

func onMessage(message *msgbus.Msg, group, command, payload string) {

	if command == "nodeAccept" {
		err := peerCertAcceptReqCert(payload)
		if err == nil {
			message.Answer(&plugin, "nodeAcceptOk", payload)
		} else {
			message.Answer(&plugin, "error", err.Error())
		}

		return
	}

	if command == "nodeReject" {
		err := peerCertReject(payload)
		if err == nil {
			message.Answer(&plugin, "nodeRejectOk", payload)
		} else {
			message.Answer(&plugin, "error", err.Error())
		}

		return
	}

	if command == "nodeAdd" { // this add per default an incoming-node-type

		// get new node
		type msgNodeAdd struct {
			Name string `json:"name"`
			Host string `json:"host"`
			Port int    `json:"port"`
		}
		var newNode msgNodeAdd
		err := json.Unmarshal([]byte(payload), &newNode)
		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		nodes.SaveData(
			newNode.Name,
			nodes.NodeTypeIncoming,
			newNode.Host,
			newNode.Port,
		)

		message.Answer(&plugin, "nodeAddOk", newNode.Name)
		return
	}

	if command == "nodeDelete" {
		nodes.Delete(payload)
		message.Answer(&plugin, "nodeDeleteOk", payload)
		return
	}

}
