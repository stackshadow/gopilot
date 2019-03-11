/*
Copyright (C) 2019 by Martin Langlotz aka stackshadow

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

package pluginwebclient

import (
	"core/clog"
	"core/msgbus"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type pluginCWs struct {
	logging clog.Logger

	plugin msgbus.Plugin
	conn   *websocket.Conn
}

var startWebSocket bool
var webSocketAddr string
var startWebServer bool
var webServerRoot string
var webServerAddr string

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func ParseCmdLine() {
	flag.BoolVar(&startWebSocket, "websocket", false, "Enable Websocket-Server")
	flag.StringVar(&webSocketAddr, "websocket.addr", "localhost:3333", "Web-Socket Adress for webinterface")

	flag.BoolVar(&startWebServer, "webserver", false, "Enable Webserver")
	flag.StringVar(&webServerAddr, "webserver.addr", "localhost:9090", "Web-Server Adress for webinterface")
	flag.StringVar(&webServerRoot, "webserver.root", "/app/www/gopilot", "Root directory of webfiles")

}

func Init() pluginCWs {
	var newCWs pluginCWs
	newCWs.logging = clog.New("WS")

	if startWebSocket == true {
		newCWs.plugin = msgbus.NewPlugin("Websocket")
		newCWs.plugin.Register()
		newCWs.plugin.ListenForGroup("", newCWs.onMessage)

		go newCWs.serveWebsocket()
	}

	if startWebServer == true {
		go newCWs.serveWebserver()
	}

	// web-server
	/*
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Welcome to my website!")
		})
	*/

	return newCWs
}

func (curCWs *pluginCWs) serveWebsocket() {
	curCWs.logging.Info("WEBSOCKET", fmt.Sprintf("Start websocker-server on %s", webSocketAddr))
	http.HandleFunc("/echo-protocol", curCWs.onWebsocketMessage)
	http.ListenAndServe(webSocketAddr, nil)
}

func (curCWs *pluginCWs) serveWebserver() {
	curCWs.logging.Info("WEBSERVER", fmt.Sprintf("Start webserver on %s", webServerAddr))
	fs := http.FileServer(http.Dir(webServerRoot))
	http.Handle("/", fs)
	http.ListenAndServe(webServerAddr, nil)
}

func (curCWs *pluginCWs) onWebsocketMessage(w http.ResponseWriter, r *http.Request) {

	var err error
	curCWs.conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer curCWs.conn.Close()

	for {
		messageType, message, err := curCWs.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		if messageType == websocket.BinaryMessage {
			log.Printf("We dont handle binary-messages yet")
			continue
		}

		if messageType == websocket.TextMessage {
			curCWs.logging.Debug("RECV", string(message[:]))
		}

		curMessage, err := msgbus.FromJsonString(string(message))
		if err != nil {
			log.Println("error: ", err)
			continue
		}

		curCWs.plugin.PublishMsg(curMessage)

	}
}

func (curCWs *pluginCWs) onMessage(message *msgbus.Msg, group, command, payload string) {
	if curCWs.conn == nil {
		return
	}

	curCWs.logging.Info("WEBSOCKET", fmt.Sprintf("%s/%s", group, command))

	jsonString, err := message.ToJsonString()
	if err != nil {
		fmt.Println("error:", err)
	}

	curCWs.logging.Debug("SEND", jsonString)
	err = curCWs.conn.WriteMessage(websocket.TextMessage, []byte(jsonString))
	if err != nil {
		log.Println("write:", err)
	}
}
