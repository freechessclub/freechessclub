package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

// Receiver receives messages and broadcasts them to all
// registered websocket connections that are Registered.
type Receiver struct {
	messages       chan []byte
	newConnections chan *websocket.Conn
	rmConnections  chan *websocket.Conn
}

// newReceiver creates a Receiver.
func newReceiver() Receiver {
	return Receiver{
		messages:       make(chan []byte, 1000), // 1000 is arbitrary
		newConnections: make(chan *websocket.Conn),
		rmConnections:  make(chan *websocket.Conn),
	}
}

// broadcast the provided message to all connected websocket connections.
// If an error occurs while writting a message to a websocket connection it is
// closed and deregistered.
func (r *Receiver) broadcast(msg []byte) {
	r.messages <- msg
}

// register the websocket connection with the receiver.
func (r *Receiver) register(conn *websocket.Conn) {
	r.newConnections <- conn
}

// deRegister the connection by closing it and removing it from our list.
func (r *Receiver) deRegister(conn *websocket.Conn) {
	r.rmConnections <- conn
}

func (r *Receiver) run() {
	conns := make([]*websocket.Conn, 0)
	for {
		select {
		case msg := <-r.messages:
			for _, conn := range conns {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.WithFields(logrus.Fields{
						"data": msg,
						"err":  err,
						"conn": conn,
					}).Error("Error writting data to connection! Closing and removing Connection")
					conns = removeConn(conns, conn)
				}
			}
		case conn := <-r.newConnections:
			conns = append(conns, conn)
		case conn := <-r.rmConnections:
			conns = removeConn(conns, conn)
		}
	}
}

func removeConn(conns []*websocket.Conn, remove *websocket.Conn) []*websocket.Conn {
	var i int
	var found bool
	for i = 0; i < len(conns); i++ {
		if conns[i] == remove {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("conns: %#v\nconn: %#v\n", conns, remove)
		panic("Conn not found")
	}
	copy(conns[i:], conns[i+1:]) // shift down
	conns[len(conns)-1] = nil    // nil last element
	return conns[:len(conns)-1]  // truncate slice
}

// publish to Redis via channel.
func (r *Receiver) publish(data []byte) {
	r.messages <- data
}
