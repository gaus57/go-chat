package main

import "log"

type Room struct {
	name string

	chat *Chat

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newRoom(chat *Chat, name string) *Room {
	return &Room{
		name:       name,
		chat:       chat,
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (room *Room) run() {
	for {
		select {
		case client := <-room.register:
			log.Println("register client in room ", room.name)
			client.room = room
			room.clients[client] = true
		case client := <-room.unregister:
			log.Println("unregister client in room ", room.name)
			if _, ok := room.clients[client]; ok {
				delete(room.clients, client)
				if len(room.clients) == 0 {
					room.chat.close <- room
				}
				close(client.send)
			}
		case message := <-room.broadcast:
			log.Println("broadcast message in room ", room.name)
			for client := range room.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(room.clients, client)
				}
			}
		}
	}
}
