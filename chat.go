package main

import "log"

type Chat struct {
	rooms map[string]*Room

	guests map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	enter chan *Invite

	close chan *Room
}

type Invite struct {
	name string

	client *Client
}

func newChat() *Chat {
	return &Chat{
		rooms:      make(map[string]*Room),
		guests:     make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		enter:      make(chan *Invite),
		close:      make(chan *Room),
	}
}

func (chat *Chat) run() {
	for {
		select {
		case client := <-chat.register:
			log.Println("register guest")
			chat.guests[client] = true
		case client := <-chat.unregister:
			log.Println("unregister guest")
			if _, ok := chat.guests[client]; ok {
				delete(chat.guests, client)
				close(client.send)
			}
		case invite := <-chat.enter:
			log.Println("invite room ", invite.name)
			if room, ok := chat.rooms[invite.name]; ok {
				room.register <- invite.client
				delete(chat.guests, invite.client)
			} else {
				room := newRoom(chat, invite.name)
				go room.run()
				room.register <- invite.client
				delete(chat.guests, invite.client)
			}
			// if !ok {
			// 	room := newRoom(chat, invite.name)
			// 	go room.run()
			// }
			// room.register <- invite.client
			// delete(chat.guests, invite.client)
		case room := <-chat.close:
			log.Println("close room ", room.name)
			if _, ok := chat.rooms[room.name]; ok {
				delete(chat.rooms, room.name)
				close(room.broadcast)
				close(room.register)
				close(room.unregister)
			}
		}
	}
}
