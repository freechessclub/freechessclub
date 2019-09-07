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
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/freechessclub/icsgo"
	"github.com/gorilla/websocket"
)

// control message
type ctlMsg struct {
	Command int    `json:"command"`
	Control string `json:"control"`
}

// Session represents a new game session
type Session struct {
	client   *icsgo.Client
	ws       *websocket.Conn
	username string
	wlock    sync.Mutex
}

func newSession(user, pass, ip string, ws *websocket.Conn) (*Session, error) {
	cmd := 1
	var username string

	// create a new FICS client
	client, err := icsgo.NewClient(
		&icsgo.Config{},
		"freechess.org:5000", user, pass)
	if err != nil {
		cmd = 2
		username = err.Error()
		return nil, fmt.Errorf("failed to create a new ICS client: %v", err)
	}

	// log.Printf("Registering IP: %s", "%i"+ip)
	// if err := client.Send("%i" + ip); err != nil {
	// 	return nil, err
	// }

	username = client.Username()
	bs, _ := json.Marshal(&ctlMsg{
		Command: cmd,
		Control: username,
	})
	err = sendWS(ws, bs)
	if err != nil {
		return nil, err
	}

	if err := client.Send("set seek 0"); err != nil {
		return nil, err
	}

	if err := client.Send("set echo 1"); err != nil {
		return nil, err
	}

	if err := client.Send("set style 12"); err != nil {
		return nil, err
	}

	if err := client.Send("set interface www.freechess.club"); err != nil {
		return nil, err
	}

	s := &Session{
		client:   client,
		ws:       ws,
		username: username,
	}

	go s.keepAlive(80 * time.Second)
	go s.ficsReader()
	return s, nil
}

func (s *Session) ficsReader() {
	for {
		msgs, err := s.client.Recv()
		if err != nil {
			s.end()
			break
		}
		if msgs == nil {
			continue
		}

		var m interface{}
		if len(msgs) == 1 {
			m = msgs[0]
		} else {
			m = msgs
		}

		bs, err := json.Marshal(m)
		if err != nil {
			log.Println("Error marshaling message")
		}
		if len(bs) == 0 {
			continue
		}

		s.wlock.Lock()
		sendWS(s.ws, []byte(bs))
		s.wlock.Unlock()
	}
}

func (s *Session) keepAlive(timeout time.Duration) {
	var lastResponse int64
	atomic.StoreInt64(&lastResponse, time.Now().UnixNano())
	s.ws.SetPongHandler(func(msg string) error {
		atomic.StoreInt64(&lastResponse, time.Now().UnixNano())
		return nil
	})

	for {
		s.wlock.Lock()
		err := s.ws.WriteMessage(websocket.PingMessage, []byte("keepalive"))
		s.wlock.Unlock()
		if err != nil {
			s.end()
			return
		}
		time.Sleep(timeout / 2)
		if atomic.LoadInt64(&lastResponse) < time.Now().Add(-timeout).UnixNano() {
			s.end()
			return
		}
	}
}

// Send sends a message to the server
func (s *Session) Send(msg string) error {
	return s.client.Send(msg)
}

func (s *Session) end() {
	s.wlock.Lock()
	s.ws.WriteMessage(websocket.CloseMessage, []byte{})
	s.wlock.Unlock()
	s.client.Send("exit")
	s.client.Destroy()
	s.ws.Close()
}
