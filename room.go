package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	trace "github.com/kubatrt/goTrace"
)

type room struct {
	// forward is a channel where incoming messages will be forward to web browser
	forward chan []byte

	// join is a channel for clients who wants to join
	join chan *client

	// leave is a channel fot clienets who wants to leave
	leave chan *client

	// clients contains all the connected users
	clients map[*client]bool

	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("Client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			// send message to all connected clients
			for client := range r.clients {
				client.send <- msg
				r.tracer.Trace("sent to client")
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// to use socket its reqauired to 'upgrade' HTTP connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: messageBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil) // get socket
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()
	go client.write() // start goroutine
	client.read()     // in main thread
}
