// Copyright © 2017 The Free Chess Club <help@freechess.club>
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
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

func sendWebsocket(ws *websocket.Conn, handle, msg string) error {
	m := &message{
		Handle: handle,
		Text:   msg,
	}
	bs, err := json.Marshal(m)
	if err = ws.WriteMessage(websocket.TextMessage, bs); err != nil {
		log.WithFields(logrus.Fields{
			"data": bs,
			"err":  err,
			"ws":   ws,
		}).Error("Error writting data to connection.")
	}
	return err
}

func keepAlive(ws *websocket.Conn, timeout time.Duration) {
	lastResponse := time.Now()
	ws.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		for {
			err := ws.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				return
			}
			time.Sleep(timeout / 2)
			if time.Now().Sub(lastResponse) > timeout {
				ws.Close()
				return
			}
		}
	}()
}

// handleWebsocket connection.
func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m := "Unable to upgrade to websockets"
		log.WithField("err", err).Println(m)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	keepAlive(ws, 50*time.Second)
	s := newSession(ws)
	for {
		ws.SetReadLimit(2048)
		mt, data, err := ws.ReadMessage()
		l := log.WithFields(logrus.Fields{"mt": mt, "data": data, "err": err})
		if err != nil {
			if err == io.EOF {
				l.Info("Websocket closed!")
			} else {
				l.Error("Error reading websocket message")
			}
			s.end()
			break
		}
		switch mt {
		case websocket.TextMessage:
			msg, err := validateMessage(data)
			if err != nil {
				l.WithFields(logrus.Fields{"msg": msg, "err": err}).Error("Invalid Message")
				break
			}
			log.Printf("Sending text to server: %s", msg.Text)
			err = s.send(msg.Text)
			if err != nil {
				l.WithFields(logrus.Fields{"msg": msg, "err": err}).Error("Error sending message")
			}
		default:
			l.Warning("Unknown Message!")
		}
	}
}
