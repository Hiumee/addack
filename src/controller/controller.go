package controller

import (
	"addack/src/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Config struct {
	ExploitsPath string
	TickTime     int64
}

type Controller struct {
	DB            *database.Database
	Hub           *Hub
	Config        *Config
	ExploitRunner *ExploitRunner
}

func (c *Controller) GetIndex(context *gin.Context) {
	context.HTML(http.StatusOK, "index", gin.H{})
	return
}

func SendError(context *gin.Context, err string) {
	context.Header("HX-Retarget", "#blackhole")
	context.Header("HX-Reswap", "innerHTML")
	context.HTML(http.StatusOK, "error", gin.H{"error": err})
}

func (c *Controller) BroadcastMessage(message []byte) {
	c.Hub.Broadcast <- message
}

// https://github.com/gorilla/websocket/blob/main/examples/chat/hub.go

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	Clients map[*websocket.Conn]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *websocket.Conn

	// Unregister requests from clients.
	Unregister chan *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
		Clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(h.Clients, client)
				}
			}
		}
	}
}
