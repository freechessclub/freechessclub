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
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	ctl       = iota // control message
	chTell           // channel tell
	pTell            // private tell
	gameMove         // game move
	gameStart        // game start
	gameEnd          // game end
	unknown          // unknown message
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

type gameStartMsg struct {
	Type      MessageType `json:"type"`
	Id        int         `json:"id"`
	PlayerOne string      `json:"playerone"`
	PlayerTwo string      `json:"playertwo"`
}

const (
	Resign = iota
	Disconnect
	Checkmate
	Draw
	Adjourn
	Abort
	TimeForfeit
	unknownReason
)

func decodeEndMessage(p1, p2, who, action string) (string, string, int) {
	switch action {
	case "resigns":
		if p1 == who {
			return p2, p1, Resign
		} else if p2 == who {
			return p1, p2, Resign
		}
	case "forfeits by disconnection":
		if p1 == who {
			return p2, p1, Disconnect
		} else if p2 == who {
			return p1, p2, Disconnect
		}
	case "checkmated":
		if p1 == who {
			return p2, p1, Checkmate
		} else if p2 == who {
			return p1, p2, Checkmate
		}
	case "forfeits on time":
		if p1 == who {
			return p2, p1, TimeForfeit
		} else if p2 == who {
			return p1, p2, TimeForfeit
		}
	case "aborted on move 1":
	case "aborted by mutual agreement":
		return p1, p2, Abort
	case "drawn by mutual agreement":
	case "drawn because both players ran out of time":
	case "drawn by repetition":
	case "drawn by the 50 move rule":
	case "drawn due to length":
	case "was drawn":
	case "player has mating material":
	case "drawn by adjudication":
	case "drawn by stalemate":
		return p1, p2, Draw
	case "adjourned by mutual agreement":
		return p1, p2, Adjourn
	}
	return p1, p2, unknownReason
}

type gameEndMsg struct {
	Type   MessageType `json:"type"`
	Id     int         `json:"id"`
	Winner string      `json:"winner"`
	Loser  string      `json:"loser"`
	Reason int         `json:"reason"`
}

type gameMoveMsg struct {
	Type  MessageType `json:"type"`
	FEN   string      `json:"fen"`
	Turn  string      `json:"turn"`
	Game  int         `json:"game"`
	WName string      `json:"wname"`
	BName string      `json:"bname"`
	Role  int         `json:"role"`
	Time  int         `json:"time"`
	Inc   int         `json:"inc"`
	WTime int         `json:"wtime"`
	BTime int         `json:"btime"`
	Move  string      `json:"move"`
}

type unknownMsg struct {
	Type MessageType `json:"type"`
	Text string      `json:"text"`
}

type incomingMsg struct {
	Type MessageType `json:"type"`
}

// validateMessage so that we know it's valid JSON and contains a Handle and
// Text
func validateMessage(data []byte) (interface{}, error) {
	var msg incomingMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, errors.Wrap(err, "unmarshaling message")
	}

	switch msg.Type {
	case ctl:
		var c ctlMsg
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, errors.Wrap(err, "unmarshaling ctl message")
		}
		return c, nil
	case chTell:
		var c chTellMsg
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, errors.Wrap(err, "unmarshaling chTell message")
		}
		return c, nil
	case pTell:
		var p pTellMsg
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, errors.Wrap(err, "unmarshaling pTell message")
		}
		return p, nil
	}
	return nil, errors.New("invalid message type")
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

func recvWebsocket(ws *websocket.Conn) interface{} {
	ws.SetReadLimit(2048)
	mt, data, err := ws.ReadMessage()
	l := log.WithFields(logrus.Fields{"mt": mt, "data": data, "err": err})
	if err != nil {
		if err == io.EOF {
			l.Info("websocket closed!")
		} else {
			l.Error("error reading websocket message")
		}
		return nil
	}

	switch mt {
	case websocket.TextMessage:
		msg, err := validateMessage(data)
		if err != nil {
			l.WithFields(logrus.Fields{"msg": msg, "err": err}).Error("invalid message")
			return nil
		}
		return msg
	default:
		l.Warning("unknown message!")
		return nil
	}
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

	user := "guest"
	pass := ""
	login := r.URL.Query().Get("login")
	if login != "" {
		msg := recvWebsocket(ws)
		if msg == nil {
			return
		}
		switch msg.(type) {
		case ctlMsg:
			m := msg.(ctlMsg)
			if m.Command == 1 {
				up := strings.Split(m.Text, ",")
				if len(up) != 2 {
					log.WithField("err", err).Println("ignoring malformed user/pass request")
					return
				}
				user = up[0][1:]
				b, err := base64.StdEncoding.DecodeString(up[1][:len(up[1])-1])
				if err != nil {
					log.WithField("err", err).Println("error decoding password")
					return
				}
				pass = string(b)
			}
		}
	}

	s, err := newSession(user, pass, ws)
	keepAlive(ws, 50*time.Second)

	for {
		msg := recvWebsocket(ws)
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
						log.WithField("err", err).Println("error sending message")
					}
				}
			} else {
				log.WithField("err", err).Println("unknown ctl command")
			}
		default:
			log.WithField("err", err).Println("ignoring unknown message from client")
		}
	}
}
