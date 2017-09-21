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
	"regexp"

	"github.com/pkg/errors"
)

// type of server messages
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

// control message
type ctlMsg struct {
	Type    MessageType `json:"type"`
	Command int         `json:"command"`
	Text    string      `json:"text"`
}

// channel tell
type chTellMsg struct {
	Type    MessageType `json:"type"`
	Channel string      `json:"channel"`
	Handle  string      `json:"handle"`
	Text    string      `json:"text"`
}

var chTellRE *regexp.Regexp

// private tell
type pTellMsg struct {
	Type   MessageType `json:"type"`
	Handle string      `json:"handle"`
	Text   string      `json:"text"`
}

var pTellRE *regexp.Regexp

// game start
type gameStartMsg struct {
	Type      MessageType `json:"type"`
	Id        int         `json:"id"`
	PlayerOne string      `json:"playerone"`
	PlayerTwo string      `json:"playertwo"`
}

var gameStartRE *regexp.Regexp

// game move
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

var gameMoveRE *regexp.Regexp

// game end
type gameEndMsg struct {
	Type    MessageType `json:"type"`
	Id      int         `json:"id"`
	Winner  string      `json:"winner"`
	Loser   string      `json:"loser"`
	Reason  int         `json:"reason"`
	Message string      `json:"message"`
}

var gameEndRE *regexp.Regexp

// type of game end messages
const (
	Resign = iota
	Disconnect
	Checkmate
	TimeForfeit
	Draw
	Adjourn
	Abort
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

type unknownMsg struct {
	Type MessageType `json:"type"`
	Text string      `json:"text"`
}

type incomingMsg struct {
	Type MessageType `json:"type"`
}

var toldMsgRE *regexp.Regexp

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

func init() {
	// game move
	// <12> rnbqkb-r pppppppp -----n-- -------- ----P--- -------- PPPPKPPP RNBQ-BNR B -1 0 0 1 1 0 7 Newton Einstein 1 2 12 39 39 119 122 2 K/e1-e2 (0:06) Ke2 0
	gameMoveRE = regexp.MustCompile(`<12>\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([rnbqkpRNBQKP\-]{8})\s([BW\-])\s(?:\-?[0-7])\s(?:[01])\s(?:[01])\s(?:[01])\s(?:[01])\s(?:[0-9]+)\s([0-9]+)\s([a-zA-Z]+)\s([a-zA-Z]+)\s(\-?[0-3])\s([0-9]+)\s([0-9]+)\s(?:[0-9]+)\s(?:[0-9]+)\s(\-?[0-9]+)\s(\-?[0-9]+)\s(?:[0-9]+)\s(?:\S+)\s\((?:[0-9]+)\:(?:[0-9]+)\)\s(\S+)\s(?:[01])\s(?:[0-9]+)\s(?:[0-9]+)\s*`)

	// {Game 117 (GuestMDPS vs. guestl) Creating unrated blitz match.}
	gameStartRE = regexp.MustCompile(`(?s)^\s*\{Game\s([0-9]+)\s\(([a-zA-Z]+)\svs\.\s([a-zA-Z]+)\)\sCreating.*\}.*`)

	gameEndRE = regexp.MustCompile(`(?s)^[^\(\):]*(?:Game\s[0-9]+:.*)?\{Game\s([0-9]+)\s\(([a-zA-Z]+)\svs\.\s([a-zA-Z]+)\)\s([a-zA-Z]+)\s([a-zA-Z0-9\s]+)\}\s(?:[012/]+-[012/]+)?.*`)

	// channel tell
	chTellRE = regexp.MustCompile(`(?s)^([a-zA-Z]+)(?:\([A-Z\*]+\))*\(([0-9]+)\):\s+(.*)`)

	// private tell
	pTellRE = regexp.MustCompile(`(?s)^([a-zA-Z]+)(?:[\(\[][A-Z0-9\*\-]+[\)\]])* (?:tells you|says|kibitzes):\s+(.*)`)

	// told status
	toldMsgRE = regexp.MustCompile(`\(told .+\)`)
}

func decodeMessage(msg []byte) (interface{}, error) {
	msg = toldMsgRE.ReplaceAll(msg, []byte{})
	if msg == nil || bytes.Equal(msg, []byte("\n")) {
		return nil, nil
	}

	matches := gameMoveRE.FindSubmatch(msg)
	if matches != nil && len(matches) >= 18 {
		m := bytes.Split(msg, []byte("\n"))
		if len(m) > 1 {
			var msgs []interface{}
			for i := 0; i < len(m); i++ {
				if len(m[i]) > 0 {
					dm, err := decodeMessage(m[i])
					if dm != nil && err == nil {
						msgs = append(msgs, dm)
					}
				}
			}
			return msgs, nil
		}

		fen := ""
		for i := 1; i < 8; i++ {
			fen += style12ToFEN(matches[i][:])
			fen += "/"
		}
		fen += style12ToFEN(matches[8][:])

		return &gameMoveMsg{
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
		}, nil
	}

	matches = gameStartRE.FindSubmatch(msg)
	if matches != nil && len(matches) > 2 {
		return &gameStartMsg{
			Type:      gameStart,
			Id:        atoi(matches[1][:]),
			PlayerOne: string(matches[2][:]),
			PlayerTwo: string(matches[3][:]),
		}, nil
	}

	matches = gameEndRE.FindSubmatch(msg)
	if matches != nil && len(matches) > 4 {
		p1 := string(matches[2][:])
		p2 := string(matches[3][:])
		who := string(matches[4][:])
		action := string(matches[5][:])

		winner, loser, reason := decodeEndMessage(p1, p2, who, action)
		return &gameEndMsg{
			Type:    gameEnd,
			Id:      atoi(matches[1][:]),
			Winner:  winner,
			Loser:   loser,
			Reason:  reason,
			Message: string(msg),
		}, nil
	}

	matches = chTellRE.FindSubmatch(msg)
	if matches != nil && len(matches) > 3 {
		return &chTellMsg{
			Type:    chTell,
			Channel: string(matches[2][:]),
			Handle:  string(matches[1][:]),
			Text:    string(bytes.Replace(matches[3][:], []byte("\n"), []byte{}, -1)),
		}, nil
	}

	matches = pTellRE.FindSubmatch(msg)
	if matches != nil && len(matches) > 2 {
		return &pTellMsg{
			Type:   pTell,
			Handle: string(matches[1][:]),
			Text:   string(bytes.Replace(matches[2][:], []byte("\n"), []byte{}, -1)),
		}, nil
	}

	return &unknownMsg{
		Type: unknown,
		Text: string(msg[:]),
	}, nil
}
