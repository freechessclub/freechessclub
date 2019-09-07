// Copyright © 2019 Free Chess Club <help@freechess.club>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/base64"
	"net"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

func sendWS(ws *websocket.Conn, bs []byte) error {
	if bs == nil || len(bs) == 0 {
		return nil
	}

	err := ws.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		log.WithFields(logrus.Fields{
			"data": bs,
			"err":  err,
			"ws":   ws,
		}).Error("error writing data to connection")
	}
	return err
}

func recvWS(ws *websocket.Conn) []byte {
	mt, data, err := ws.ReadMessage()
	if err != nil {
		l := log.WithFields(logrus.Fields{"mt": mt, "data": data, "err": err})
		if websocket.IsUnexpectedCloseError(err) {
			l.Info("websocket closed!")
		} else {
			l.Error("error reading websocket message")
		}
		return nil
	}
	return data
}

func wsHandler(user, pass, ip string, ws *websocket.Conn) {
	s, err := newSession(user, pass, ip, ws)
	if s == nil || err != nil {
		log.WithField("err", err).Println("failed to create a new session")
		return
	}

	for {
		msg := recvWS(ws)
		if msg == nil {
			s.end()
			return
		}

		// log.Printf("Sending msg to server: %s", msg.Message)
		err = s.Send(string(msg))
		if err != nil {
			log.WithField("err", err).Println("error sending message")
		}
	}
}

// handleWebsocket connection.
func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errmsg := "unable to upgrade to websockets"
		log.WithField("err", err).Println(errmsg)
		http.Error(w, errmsg, http.StatusBadRequest)
		return
	}
	ws.SetReadLimit(2048)

	user := "guest"
	pass := ""
	login := r.URL.Query().Get("login")
	if login != "" {
		msg := recvWS(ws)
		if msg == nil {
			return
		}

		up := strings.Split(string(msg), ",")
		if len(up) > 1 {
			user = up[0][1:]
			b, err := base64.StdEncoding.DecodeString(up[1][:len(up[1])-1])
			if err != nil {
				errmsg := "error decoding password"
				log.WithField("err", err).Println(errmsg)
				http.Error(w, errmsg, http.StatusUnauthorized)
				return
			}
			pass = string(b)
		} else {
			user = up[0][1 : len(up[0])-1]
		}
	}

	ip := r.Header.Get("X-Forwarded-For")
	if len(ip) == 0 {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	go wsHandler(user, pass, ip, ws)
}
