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

import (
	"bufio"
	"core/clog"
	"core/msgbus"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"plugins/core"
	"strings"
)

type tlsSession struct {
	logging        clog.Logger
	plugin         msgbus.Plugin
	conn           net.Conn
	remoteNodeName string
	nodeType       int
	myChallange    string
}

func NewSession(sessionNo string, nodeType int, connection net.Conn) {
	var newSession tlsSession

	newSession.logging = clog.New("SESSION-" + sessionNo)
	newSession.plugin = msgbus.NewPlugin("SESSION-" + sessionNo)
	newSession.plugin.Register()
	newSession.conn = connection
	newSession.nodeType = nodeType

	newSession.handleClient()

	return
}

func (curSession *tlsSession) readMsg(bufReader *bufio.Reader) (msgbus.Msg, error) {

	var newMessage msgbus.Msg

	for {
		msgString, err := bufReader.ReadString('\n')
		if err != nil {
			curSession.logging.Error("readMsg", err.Error())
			return newMessage, err
		}
		msgString = strings.TrimSuffix(msgString, "\n")
		curSession.logging.Debug("readMsg", msgString)

		newMessage, _ := msgbus.FromJsonString(string(msgString))

		return newMessage, nil
	}

}

func (curSession *tlsSession) writeData(src, trg, grp, cmd, payload string) error {

	// create new message and string
	var newMessage msgbus.Msg
	newMessage.NodeSource = src
	newMessage.NodeTarget = trg
	newMessage.Group = grp
	newMessage.Command = cmd
	newMessage.Payload = payload

	return curSession.writeMsg(&newMessage)
}

func (curSession *tlsSession) writeMsg(message *msgbus.Msg) error {

	jsonByteArray, err := message.ToJsonByteArray()
	if err != nil {
		return err
	}
	jsonByteArray = append(jsonByteArray, '\n')

	_, err = curSession.conn.Write(jsonByteArray)
	if err != nil {
		return err
	}

	curSession.logging.Debug("writeMsg", string(jsonByteArray))
	return nil
}

func (curSession *tlsSession) handleClient() {
	defer curSession.conn.Close()
	defer curSession.plugin.DeRegister()
	curSession.logging.Info("handleClient", fmt.Sprintf("%s: New client", curSession.conn.RemoteAddr().String()))

	// show connection info
	tlsConn := curSession.conn.(*tls.Conn)

	// run the handshake
	err := tlsConn.Handshake()
	if err != nil {
		curSession.logging.Error("handleClient", fmt.Sprintf("Handshake failed: %s", err))
		return
	}

	// some debugging infos
	state := tlsConn.ConnectionState()
	curSession.logging.Debug("handleClient", fmt.Sprintf("Version: %d", state.Version))
	curSession.logging.Debug("handleClient", fmt.Sprintf("HandshakeComplete: %t", state.HandshakeComplete))
	curSession.logging.Debug("handleClient", fmt.Sprintf("DidResume: %t", state.DidResume))
	curSession.logging.Debug("handleClient", fmt.Sprintf("CipherSuite: %x", state.CipherSuite))

	// check certificate
	if len(state.PeerCertificates) == 0 {
		curSession.logging.Error("handleClient", "Missing peer certificate")
		return
	}
	peerCert := state.PeerCertificates[0]

	// remode-node-name
	curSession.remoteNodeName = peerCert.Subject.CommonName

	// check certificate
	peerCertCheckResult := curSession.peerCertCheck(peerCert.Signature)
	if peerCertCheckResult == certCheckReq {
		curSession.plugin.Publish(core.NodeName, core.NodeName, "tls", "nodeReq", peerCert.Subject.CommonName)
		return
	}
	if peerCertCheckResult != certCheckOk {
		return
	}

	// challange
	challangeResult := curSession.handleChallange()
	if challangeResult == false {
		return
	}

	// successfully connected
	curSession.plugin.Publish(core.NodeName, core.NodeName, "tls", "nodeConnected", curSession.remoteNodeName)
	defer curSession.plugin.Publish(core.NodeName, core.NodeName, "tls", "nodeDisconnect", curSession.remoteNodeName)
	curSession.plugin.ListenForGroup("", curSession.onMessage)

	r := bufio.NewReader(curSession.conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			curSession.logging.Error("handleClient", err.Error())
			return
		}
		msg = strings.TrimSuffix(msg, "\n")

		// convert to Msg
		curMessage, err := msgbus.FromJsonString(string(msg))
		if err != nil {
			curSession.logging.Error("handleClient", err.Error())
			continue
		}

		// publish it to BUS
		curSession.plugin.PublishMsg(curMessage)

		//curSession.logging.Info("TLS", msg)

	}

}

const certCheckErr int = -2
const certCheckMisMatch int = -1
const certCheckOk int = 0
const certCheckReq int = 1

func (curSession *tlsSession) peerCertCheck(peerCertSignature []byte) int {

	nodeObject, err := core.GetNodeObject(curSession.remoteNodeName)
	if err != nil {
		logging.Error("CLIENT", err.Error())
		return certCheckErr
	}

	// no cert signature present
	// as server we save it to
	// as client we "cherry pick"
	if (*nodeObject)["peerCertSignature"] == nil {

		if curSession.nodeType == core.NodeTypeIncoming {

			logging.Info("CLIENT", fmt.Sprintf(
				"Peer Certificate missing for '%s', save it to requested keys. Signature: %x",
				curSession.remoteNodeName, peerCertSignature),
			)

			(*nodeObject)["peerCertSignatureReq"] = fmt.Sprintf("%x", peerCertSignature)
			core.ConfigSave()
			return certCheckReq
		}

		if curSession.nodeType == core.NodeTypeClient {

			logging.Info("CLIENT", fmt.Sprintf(
				"Cherry pick Signature: %x for '%s'",
				peerCertSignature, curSession.remoteNodeName),
			)

			(*nodeObject)["peerCertSignature"] = fmt.Sprintf("%x", peerCertSignature)
			core.ConfigSave()
			return certCheckOk
		}

	}

	// cert of remote-node is present, check it against tls-cert
	if (*nodeObject)["peerCertSignature"] != fmt.Sprintf("%x", peerCertSignature) {
		logging.Error("CLIENT", fmt.Sprintf("Peer Certificate '%x' not accepted for this node", peerCertSignature))
		return certCheckMisMatch
	}
	logging.Info("CLIENT", "Peer Certificate accepted")

	return certCheckOk
}

func (curSession *tlsSession) handleChallange() bool {

	nodeObject, err := core.GetNodeObject(curSession.remoteNodeName)
	bufReader := bufio.NewReader(curSession.conn)

	// no shared secret
	if (*nodeObject)["sharedSecret"] == nil {

		// is this an incoming connection -> send "newSecret" command
		if curSession.nodeType == core.NodeTypeIncoming {

			randomBytes, err := GenerateRandomBytes(32)
			if err != nil {
				return false
			}
			randomString := base64.StdEncoding.EncodeToString(randomBytes)

			(*nodeObject)["sharedSecret"] = randomString
			core.ConfigSave()

			curSession.logging.Error("CHALLANGE", fmt.Sprintf(
				"New SharedSecret generated %s, send it to client",
				randomString,
			))

			err = curSession.writeData(
				core.NodeName, curSession.remoteNodeName,
				"challange", "newSecret", randomString,
			)
			if err != nil {
				curSession.logging.Error("CHALLANGE", err.Error())
				return false
			}

			// wait for respond
			message, err := curSession.readMsg(bufReader)
			if err != nil {
				return false
			}
			if message.Command != "newSecretSaved" {
				curSession.logging.Error(
					"newSecret",
					fmt.Sprintf("We expect the 'newSecretSaved' command, but we get '%s', this is illegal", message.Command),
				)
				return false
			}

		}

		// we wait for "newSecret" command
		if curSession.nodeType == core.NodeTypeClient {

			// wait for newSecret
			message, err := curSession.readMsg(bufReader)
			if err != nil {
				return false
			}
			if message.Command != "newSecret" {
				curSession.logging.Error(
					"CHALLANGE",
					fmt.Sprintf("We expect the newSecret command, but we get %s, this is illegal", message.Command),
				)
				return false
			}

			(*nodeObject)["sharedSecret"] = message.Payload
			core.ConfigSave()

			// answer
			err = curSession.writeData(
				core.NodeName, curSession.remoteNodeName,
				"challange", "newSecretSaved", "",
			)
			if err != nil {
				curSession.logging.Error("CHALLANGE", err.Error())
				return false
			}

		}

	}

	// get shared secret
	var sharedSecret string
	sharedSecret = (*nodeObject)["sharedSecret"].(string)

	// create random bytes and remember
	randomMessageBytes, err := GenerateRandomBytes(32)
	if err != nil {
		return false
	}
	curSession.myChallange = base64.StdEncoding.EncodeToString(randomMessageBytes)

	err = curSession.writeData(
		core.NodeName, curSession.remoteNodeName,
		"challange", "challangeRequest", curSession.myChallange,
	)
	if err != nil {
		curSession.logging.Error("CHALLANGE", err.Error())
		return false
	}
	curSession.logging.Info("CHALLANGE", fmt.Sprintf(
		"Send challange %s and await %s",
		curSession.myChallange,
		ComputeHmac256(curSession.myChallange, sharedSecret),
	))

	for {
		message, err := curSession.readMsg(bufReader)
		if err != nil {
			return false
		}

		if message.Command == "challangeRequest" {

			challangeResponse := ComputeHmac256(message.Payload, sharedSecret)

			err := curSession.writeData(
				core.NodeName, curSession.remoteNodeName,
				"challange", "challangeResponse", challangeResponse,
			)
			if err != nil {
				curSession.logging.Error("CHALLANGE", err.Error())
				return false
			}

			continue
		}

		if message.Command == "challangeResponse" {

			challangeResponse := ComputeHmac256(curSession.myChallange, sharedSecret)

			if challangeResponse == message.Payload {
				curSession.logging.Info(
					"CHALLANGE",
					fmt.Sprintf("Challange ok"),
				)

				return true
			}

			curSession.logging.Error(
				"CHALLANGE",
				fmt.Sprintf("Challange %s != %s", challangeResponse, message.Payload),
			)
			break
		}

		break

	}

	return false
}

func (curSession *tlsSession) onMessage(message *msgbus.Msg, group, command, payload string) {

	// messages to us, will not sended
	if message.NodeTarget == core.NodeName {
		curSession.logging.Debug("onMessage", fmt.Sprintf(
			"I will not send out messages, which are dedicated to me",
		))
		return
	}

	// only message to remoteNodeName will be sended
	if /* message.NodeTarget != curSession.remoteNodeName || */ len(message.NodeTarget) == 0 {
		curSession.logging.Debug("onMessage", fmt.Sprintf(
			"I will not send out messages, which are not for target. '%s' != '%s'",
			message.NodeTarget, curSession.remoteNodeName,
		))
		return
	}

	curSession.writeMsg(message)
}
