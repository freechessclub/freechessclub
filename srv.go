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

const (
	tsKey = "Timestamp (FICS) v1.0 - programmed by Henrik Gram."
	hello = "TIMESEAL2|freeseal|Free Chess Club|"
)

type Session struct {
	conn     *telnet.Conn
	ws       *websocket.Conn
	username string
	wlock    sync.Mutex
	rlock    sync.Mutex
}

func Crypt(b []byte, l int) []byte {
	s := make([]byte, l+30)
	copy(s[:l], b)
	s[l] = 0x18
	l++
	ts := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	copy(s[l:], ts)
	l += len(ts)
	s[l] = 0x19
	l++
	for ; (l % 12) != 0; l++ {
		s[l] = 0x31
	}

	for n := 0; n < l; n += 12 {
		s[n] ^= s[n+11]
		s[n+11] ^= s[n]
		s[n] ^= s[n+11]
		s[n+2] ^= s[n+9]
		s[n+9] ^= s[n+2]
		s[n+2] ^= s[n+9]
		s[n+4] ^= s[n+7]
		s[n+7] ^= s[n+4]
		s[n+4] ^= s[n+7]
	}

	for n := 0; n < l; n++ {
		var x = int8(((s[n] | 0x80) ^ tsKey[n%50]) - 32)
		s[n] = byte(x)
	}

	s[l] = 0x80
	l++
	s[l] = 0x0a
	l++
	return s[:l]
}

func Connect(network, addr, ip string, timeout, retries int) (*telnet.Conn, error) {
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
	log.Printf("Connected! (IP: %s)", ip)

	conn.SetReadDeadline(time.Now().Add(ts))
	conn.SetWriteDeadline(time.Now().Add(ts))

	send(conn, "%i"+ip)
	send(conn, hello)
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
	buf := Crypt([]byte(cmd), len(cmd))
	_, err := conn.Conn.Write(buf)
	return err
}

func readUntil(conn *telnet.Conn, delims ...string) ([]byte, error) {
	b, err := conn.ReadUntil(delims...)
	for {
		i := bytes.Index(b, []byte{'[', 'G', ']', 0x00})
		if i == -1 {
			break
		}
		send(conn, string([]byte{0x02, 0x39}))
		b = append(b[:i], b[i+4:]...)
	}
	return sanitize(b), err
}

func sendAndReadUntil(conn *telnet.Conn, cmd string, delims ...string) ([]byte, error) {
	err := send(conn, cmd)
	if err != nil {
		return nil, err
	}
	return readUntil(conn, delims...)
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
	readUntil(conn, loginPrompt)
	_, err := sendAndReadUntil(conn, username, prompt)
	if err != nil {
		return "", fmt.Errorf("Error creating new login session for %s: %v", username, err)
	}

	// wait for the password prompt
	out, err := sendAndReadUntil(conn, password, "****\n")
	if err != nil {
		return "", fmt.Errorf("Failed authentication for %s: %v", username, err)
	}

	re := regexp.MustCompile("\\*\\*\\*\\* Starting FICS session as ([a-zA-Z]+)(?:\\(U\\))? \\*\\*\\*\\*")
	user := re.FindSubmatch(out)
	if user != nil && len(user) > 1 {
		username = string(user[1][:])
		log.Printf("Logged in as %s", username)
		return username, nil
	} else {
		return "", fmt.Errorf("Invalid password for %s", username)
	}
}

func (s *Session) ficsReader() {
	for {
		s.conn.SetReadDeadline(time.Now().Add(3600 * time.Second))
		out, err := readUntil(s.conn, ficsPrompt)
		if err != nil {
			s.end()
			return
		}
		if len(out) == 0 {
			continue
		}

		msgs, err := decodeMessage(out)
		if err != nil {
			log.Println("Error decoding message")
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
			log.Println("Error marshaling message")
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

func newSession(user, pass, ip string, ws *websocket.Conn) (*Session, error) {
	conn, err := Connect("tcp", "freechess.org:5000", ip, 5, 5)
	if err != nil {
		return nil, err
	}

	username, err := Login(conn, user, pass)
	cmd := 1
	if err != nil {
		cmd = 2
		username = err.Error()
	}

	msg := &ctlMsg{
		Type:    ctl,
		Command: cmd,
		Text:    username,
	}
	bs, _ := json.Marshal(msg)
	sendWS(ws, bs)

	if err != nil {
		return nil, err
	}

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

	go s.keepAlive(80 * time.Second)
	go s.ficsReader()
	return s, nil
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

func (s *Session) end() {
	s.wlock.Lock()
	s.ws.WriteMessage(websocket.CloseMessage, []byte{})
	s.wlock.Unlock()
	send(s.conn, "exit")
	s.conn.Close()
	s.ws.Close()
}
