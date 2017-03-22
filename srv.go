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
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ziutek/telnet"
)

const (
	loginPrompt    = "login:"
	passwordPrompt = "password:"
	ficsPrompt     = "fics%"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Session struct {
	conn     *telnet.Conn
	ws       *websocket.Conn
	username string
}

var chTellRE *regexp.Regexp
var pTellRE *regexp.Regexp
var gameMoveRE *regexp.Regexp
var gameStartRE *regexp.Regexp
var toldMsgRE *regexp.Regexp

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
	b = bytes.TrimSuffix(b, []byte(ficsPrompt))
	b = bytes.Replace(b, []byte("\x00"), []byte{}, -1)
	b = bytes.Replace(b, []byte("\\   "), []byte{}, -1)
	b = bytes.Replace(b, []byte("\r"), []byte{}, -1)
	b = bytes.Replace(b, []byte("\n"), []byte(" "), -1)
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
	if username != "guest" {
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
	_, err = sendAndReadUntil(conn, password, ficsPrompt)
	if err != nil {
		return "", fmt.Errorf("failed authentication for %s (password %s): %v", username, password, err)
	}

	log.Printf("Logged in as %s", username)

	//fmt.Println(string(motd[:]))
	return username, nil
}

func init() {
	// game move
	// <12> rnbqkb-r pppppppp -----n-- -------- ----P--- -------- PPPPKPPP RNBQ-BNR B -1 0 0 1 1 0 7 Newton Einstein 1 2 12 39 39 119 122 2 K/e1-e2 (0:06) Ke2 0
	gameMoveRE = regexp.MustCompile(`<12>\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([BW\-])\s(?:\-?[0-7])\s(?:[01])\s(?:[01])\s(?:[01])\s(?:[01])\s(?:[0-9]+)\s([0-9]+)\s([a-zA-Z]+)\s([a-zA-Z]+)\s(\-?[0-3])\s([0-9]+)\s([0-9]+)\s(?:[0-9]+)\s(?:[0-9]+)\s(\-?[0-9]+)\s(\-?[0-9]+)\s(?:[0-9]+)\s(?:\S+)\s\((?:[0-9]+)\:(?:[0-9]+)\)\s(\S+)\s(?:[01]).*`)

	// {Game 117 (GuestMDPS vs. guestl) Creating unrated blitz match.}
	gameStartRE = regexp.MustCompile(`\s*\{Game\s([0-9]+)\s\(([a-zA-Z]+)\svs\.\s([a-zA-Z]+)\)\sCreating.*\}`)

	// channel tell
	chTellRE = regexp.MustCompile(`([a-zA-Z]+)(?:\([A-Z\*]+\))*\(([0-9]+)\):\s+(.*)`)

	// private tell
	pTellRE = regexp.MustCompile(`([a-zA-Z]+)(?:[\(\[][A-Z0-9\*\-]+[\)\]])* (?:tells you|says|kibitzes):\s+(.*)`)

	// told status
	toldMsgRE = regexp.MustCompile(`\s*\(told .+\)\s*`)
}

func style12ToFEN(b []byte) string {
	str := string(b[:])
	var fen string
	count := 0
	for i := 0; i < 8; i++ {
		if str[i] == '-' {
			count++
			if i == 7 {
				fen += strconv.Itoa(count)
			}
		} else {
			if count > 0 {
				fen += strconv.Itoa(count)
				count = 0
			}
			fen += string(str[i])
		}
	}
	return fen
}

func atoi(b []byte) int {
	i, _ := strconv.Atoi(string(b))
	return i
}

func (s *Session) decodeMessage(msg []byte) ([]byte, error) {
	msg = toldMsgRE.ReplaceAll(msg, []byte{})
	if msg == nil {
		return nil, nil
	}
	matches := gameMoveRE.FindSubmatch(msg)
	if matches != nil && len(matches) >= 18 {
		fen := ""
		for i := 1; i < 8; i++ {
			fen += style12ToFEN(matches[i][:])
			fen += "/"
		}
		fen += style12ToFEN(matches[8][:])

		m := &gameMoveMsg{
			Type:  gameMove,
			FEN:   fen,
			Turn:  string(matches[9][:]),
			Game:  atoi(matches[10][:]),
			WName: string(matches[11][:]),
			BName: string(matches[12][:]),
			Role:  atoi(matches[13][:]),
			Time:  atoi(matches[14][:]),
			Inc:   atoi(matches[15][:]),
			WTime: atoi(matches[16][:]),
			BTime: atoi(matches[17][:]),
			Move:  string(matches[18][:]),
		}
		return json.Marshal(m)
	}

	matches = gameStartRE.FindSubmatch(msg)
	if matches != nil && len(matches) > 3 {
		m := &gameStartMsg{
			Type:      gameStart,
			Id:        atoi(matches[1][:]),
			PlayerOne: string(matches[2][:]),
			PlayerTwo: string(matches[3][:]),
		}
		return json.Marshal(m)
	}

	matches = chTellRE.FindSubmatch(msg)
	if matches != nil && len(matches) > 3 {
		m := &chTellMsg{
			Type:    chTell,
			Channel: string(matches[2][:]),
			Handle:  string(matches[1][:]),
			Text:    string(matches[3][:]),
		}
		return json.Marshal(m)
	}

	matches = pTellRE.FindSubmatch(msg)
	if matches != nil && len(matches) > 2 {
		m := &pTellMsg{
			Type:   pTell,
			Handle: string(matches[1][:]),
			Text:   string(matches[2][:]),
		}
		return json.Marshal(m)
	}

	m := &unknownMsg{
		Type: unknown,
		Text: string(msg[:]),
	}
	return json.Marshal(m)
}

func (s *Session) ficsReader() {
	for {
		s.conn.SetReadDeadline(time.Now().Add(3600 * time.Second))
		out, err := s.conn.ReadUntil(ficsPrompt)
		if err != nil {
			s.ws.WriteMessage(websocket.CloseMessage, []byte{})
			log.Println("Closing session.")
			return
		}
		out = sanitize(out)
		bs, err := s.decodeMessage(out)
		if err != nil {
			log.Println("error decoding message")
		}
		if bs != nil {
			sendWebsocket(s.ws, bs)
		}
	}
}

func (s *Session) send(msg string) error {
	return send(s.conn, msg)
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
	sendWebsocket(ws, bs)

	_, err = sendAndReadUntil(conn, "set seek 0", ficsPrompt)
	if err != nil {
		return nil, err
	}

	_, err = sendAndReadUntil(conn, "set echo 0", ficsPrompt)
	if err != nil {
		return nil, err
	}

	_, err = sendAndReadUntil(conn, "set style 12", ficsPrompt)
	if err != nil {
		return nil, err
	}

	_, err = sendAndReadUntil(conn, "set interface fcc", ficsPrompt)
	if err != nil {
		return nil, err
	}

	s := &Session{
		conn:     conn,
		ws:       ws,
		username: username,
	}

	go s.ficsReader()
	return s, nil
}

func (s *Session) end() {
	s.ws.WriteMessage(websocket.CloseMessage, []byte{})
	send(s.conn, "exit")
	s.conn.Close()
}
