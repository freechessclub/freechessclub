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
	"github.com/pkg/errors"
)

const (
	ctl      = iota // control message
	chTell          // channel tell
	pTell           // private tell
	gameMove        // game move
	unknown         // unknown message
)

type MessageType int

type ctlMsg struct {
	Type    MessageType `json:"type"`
	Command int         `json:"command"`
	Text    string      `json:"text"`
}

type chTellMsg struct {
	Type    MessageType `json:"type"`
	Channel string      `json:"channel"`
	Handle  string      `json:"handle"`
	Text    string      `json:"text"`
}

type pTellMsg struct {
	Type   MessageType `json:"type"`
	Handle string      `json:"handle"`
	Text   string      `json:"text"`
}

type gameMoveMsg struct {
	Type     MessageType `json:"type"`
	Handle   string      `json:"handle"`
	Opponent string      `json:"opponent"`
	FEN      string      `json:"fen"`
	Text     string      `json:"text"`
}

type unknownMsg struct {
	Type MessageType `json:"type"`
	Text string      `json:"text"`
}

// validateMessage so that we know it's valid JSON and contains a Handle and
// Text
func validateMessage(data []byte) (pTellMsg, error) {
	var msg pTellMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, errors.Wrap(err, "unmarshaling message")
	}

	if msg.Text == "" {
		return msg, errors.New("message has no text")
	}
	if msg.Type < ctl || msg.Type > unknown {
		return msg, errors.New("invalid message type")
	}
	return msg, nil
}

func sendWebsocket(ws *websocket.Conn, bs []byte) error {
	var err error
	if err = ws.WriteMessage(websocket.TextMessage, bs); err != nil {
		log.WithFields(logrus.Fields{
			"data": bs,
			"err":  err,
			"ws":   ws,
		}).Error("error writting data to connection.")
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
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m := "unable to upgrade to websockets"
		log.WithField("err", err).Println(m)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	s := newSession(ws)
	keepAlive(ws, 50*time.Second)

	for {
		ws.SetReadLimit(2048)
		mt, data, err := ws.ReadMessage()
		l := log.WithFields(logrus.Fields{"mt": mt, "data": data, "err": err})
		if err != nil {
			if err == io.EOF {
				l.Info("websocket closed!")
			} else {
				l.Error("error reading websocket message")
			}
			s.end()
			break
		}
		switch mt {
		case websocket.TextMessage:
			msg, err := validateMessage(data)
			if err != nil {
				l.WithFields(logrus.Fields{"msg": msg, "err": err}).Error("invalid message")
				break
			}
			log.Printf("Sending text to server: %s", msg.Text)
			err = s.send(msg.Text)
			if err != nil {
				l.WithFields(logrus.Fields{"msg": msg, "err": err}).Error("error sending message")
			}
		default:
			l.Warning("unknown message!")
		}
	}
}
