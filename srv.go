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
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ziutek/telnet"
)

const (
	loginPrompt    = "login:"
	passwordPrompt = "password:"
	newLine        = "\n"
	ficsPrompt     = "fics%"
)

type Session struct {
	conn     *telnet.Conn
	ws       *websocket.Conn
	username string
	wlock    sync.Mutex
	rlock    sync.Mutex
}

func Connect(network, addr string, timeout, retries int) (*telnet.Conn, error) {
	ts := time.Duration(timeout) * time.Second

	var conn *telnet.Conn
	var connected bool = false
	var err error = nil

	for attempts := 1; attempts <= retries && connected != true; attempts++ {
		log.Printf("Connecting to chess server %s (attempt %d of %d)...", addr, attempts, retries)
		conn, err = telnet.DialTimeout(network, addr, ts)
		if err != nil {
			continue
		}
		connected = true
	}
	if err != nil || connected == false {
		return nil, fmt.Errorf("error connecting to server %s: %v", addr, err)
	}
	log.Printf("Connected!")

	conn.SetUnixWriteMode(true)
	conn.SetReadDeadline(time.Now().Add(ts))
	conn.SetWriteDeadline(time.Now().Add(ts))
	return conn, nil
}

func sanitize(b []byte) []byte {
	b = bytes.Replace(b, []byte("\u0007"), []byte{}, -1)
	b = bytes.Replace(b, []byte("\x00"), []byte{}, -1)
	b = bytes.Replace(b, []byte("\\   "), []byte{}, -1)
	b = bytes.Replace(b, []byte("\r"), []byte{}, -1)
	b = bytes.Replace(b, []byte("fics%"), []byte{}, -1)
	return bytes.TrimSpace(b)
}

func send(conn *telnet.Conn, cmd string) error {
	conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
	buf := make([]byte, len(cmd)+1)
	copy(buf, cmd)
	buf[len(cmd)] = '\n'
	_, err := conn.Write(buf)
	return err
}

func sendAndReadUntil(conn *telnet.Conn, cmd string, delims ...string) ([]byte, error) {
	err := send(conn, cmd)
	if err != nil {
		return nil, err
	}
	b, err := conn.ReadUntil(delims...)
	return sanitize(b), err
}

func Login(conn *telnet.Conn, username, password string) (string, error) {
	var prompt string
	// guests have no passwords
	if username != "guest" && len(password) > 0 {
		prompt = passwordPrompt
	} else {
		prompt = "Press return to enter the server as"
		password = ""
	}

	// wait for the login prompt
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn.ReadUntil(loginPrompt)
	out, err := sendAndReadUntil(conn, username, prompt)
	if err != nil {
		return "", fmt.Errorf("error creating new login session for %s: %v", username, err)
	}

	re := regexp.MustCompile("Logging you in as \"([a-zA-Z]+)\"")
	user := re.FindSubmatch(out)
	if user != nil {
		username = string(user[1][:])
	}

	// wait for the password prompt
	_, err = sendAndReadUntil(conn, password, newLine)
	if err != nil {
		return "", fmt.Errorf("failed authentication for %s (password %s): %v", username, password, err)
	}

	log.Printf("Logged in as %s", username)

	//fmt.Println(string(motd[:]))
	return username, nil
}

func (s *Session) ficsReader() {
	for {
		s.conn.SetReadDeadline(time.Now().Add(3600 * time.Second))
		out, err := s.conn.ReadUntil(ficsPrompt)

		if err != nil {
			s.end()
			return
		}
		out = sanitize(out)
		if len(out) == 0 {
			continue
		}

		msgs, err := decodeMessage(out)
		if err != nil {
			log.Println("error decoding message")
		}

		if msgs == nil {
			continue
		}

		arr, ok := msgs.([]interface{})
		if ok && len(arr) == 0 {
			continue
		}

		bs, err := json.Marshal(msgs)
		if err != nil {
			log.Println("error marshaling message")
		}

		if bs == nil {
			continue
		}

		s.wlock.Lock()
		sendWS(s.ws, bs)
		s.wlock.Unlock()
	}
}

func (s *Session) send(msg string) error {
	return send(s.conn, msg)
}

func (s *Session) recvWS() interface{} {
	return recvWS(s.ws, &s.rlock)
}

func newSession(user, pass string, ws *websocket.Conn) (*Session, error) {
	conn, err := Connect("tcp", "freechess.org:5000", 5, 5)
	if err != nil {
		return nil, err
	}

	username, err := Login(conn, user, pass)
	if err != nil {
		return nil, err
	}

	msg := &ctlMsg{
		Type:    ctl,
		Command: 1,
		Text:    username,
	}
	bs, _ := json.Marshal(msg)
	sendWS(ws, bs)

	_, err = sendAndReadUntil(conn, "set seek 0", newLine)
	if err != nil {
		return nil, err
	}

	_, err = sendAndReadUntil(conn, "set echo 1", newLine)
	if err != nil {
		return nil, err
	}

	_, err = sendAndReadUntil(conn, "set style 12", newLine)
	if err != nil {
		return nil, err
	}

	_, err = sendAndReadUntil(conn, "set interface www.freechess.club", newLine)
	if err != nil {
		return nil, err
	}

	s := &Session{
		conn:     conn,
		ws:       ws,
		username: username,
	}

	go s.ficsReader()
	go s.keepAlive(50 * time.Second)
	return s, nil
}

func (s *Session) keepAlive(timeout time.Duration) {
	var lastResponse int64
	atomic.StoreInt64(&lastResponse, time.Now().UnixNano())
	s.rlock.Lock()
	s.ws.SetPongHandler(func(msg string) error {
		fmt.Println("pong ", &lastResponse)
		atomic.StoreInt64(&lastResponse, time.Now().UnixNano())
		return nil
	})
	s.rlock.Unlock()

	for {
		s.wlock.Lock()
		err := s.ws.WriteMessage(websocket.PingMessage, []byte("keepalive"))
		s.wlock.Unlock()
		if err != nil {
			return
		}
		time.Sleep(timeout / 2)
		if atomic.LoadInt64(&lastResponse) < time.Now().Add(-timeout).UnixNano() {
			s.end()
			return
		}
	}
}

func (s *Session) end() {
	s.wlock.Lock()
	s.ws.WriteMessage(websocket.CloseMessage, []byte{})
	s.wlock.Unlock()
	send(s.conn, "exit")
	s.conn.Close()
	s.ws.Close()
}
