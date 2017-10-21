// Copyright © 2017 Free Chess Club <help@freechess.club>
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
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

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
		}).Error("Error writing data to connection")
	}
	return err
}

func recvWS(ws *websocket.Conn, lock *sync.Mutex) interface{} {
	if lock != nil {
		lock.Lock()
	}
	ws.SetReadLimit(2048)
	mt, data, err := ws.ReadMessage()
	if lock != nil {
		lock.Unlock()
	}

	if err != nil {
		l := log.WithFields(logrus.Fields{"mt": mt, "data": data, "err": err})
		if err == io.EOF {
			l.Info("websocket closed!")
		} else {
			l.Error("Error reading websocket message")
		}
		return nil
	}

	switch mt {
	case websocket.TextMessage:
		msg, err := validateMessage(data)
		if err != nil {
			log.WithFields(logrus.Fields{"msg": msg, "err": err}).Error("Invalid message")
			return nil
		}
		return msg
	default:
		log.Warning("unknown message!")
		return nil
	}
}

func wsHandler(user, pass, ip string, ws *websocket.Conn) {
	s, err := newSession(user, pass, ip, ws)
	if err != nil {
		log.WithField("err", err).Println("Failed to create a new session")
		return
	}

	for {
		msg := s.recvWS()
		if msg == nil {
			if s != nil {
				s.end()
			}
			return
		}

		switch msg.(type) {
		case ctlMsg:
			m := msg.(ctlMsg)
			if m.Command == 0 {
				// log.Printf("Sending text to server: %s", msg.Text)
				if s != nil {
					err = s.send(m.Text)
					if err != nil {
						log.WithField("err", err).Println("Error sending message")
					}
				}
			} else {
				log.WithField("err", err).Println("Unknown control command")
			}
		default:
			log.WithField("err", err).Println("Ignoring unknown message from client")
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
		errmsg := "Unable to upgrade to websockets"
		log.WithField("err", err).Println(errmsg)
		http.Error(w, errmsg, http.StatusBadRequest)
		return
	}

	user := "guest"
	pass := ""
	login := r.URL.Query().Get("login")
	if login != "" {
		msg := recvWS(ws, nil)
		if msg == nil {
			return
		}
		switch msg.(type) {
		case ctlMsg:
			m := msg.(ctlMsg)
			if m.Command == 1 {
				up := strings.Split(m.Text, ",")
				if len(up) > 1 {
					user = up[0][1:]
					b, err := base64.StdEncoding.DecodeString(up[1][:len(up[1])-1])
					if err != nil {
						errmsg := "Error decoding password"
						log.WithField("err", err).Println(errmsg)
						http.Error(w, errmsg, http.StatusUnauthorized)
						return
					}
					pass = string(b)
				} else {
					user = up[0][1 : len(up[0])-1]
				}
			}
		}
	}

	ip := r.Header.Get("X-Forwarded-For")
	if len(ipHostPort) == 0 {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	}

	go wsHandler(user, pass, ip, ws)
}
