package main

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type Hub struct {
	sync.RWMutex

	publish    chan *OutputMessage
	broadcast  chan *OutputMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		publish:    make(chan *OutputMessage),
		broadcast:  make(chan *OutputMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) publishInfo(room *Room, client *Client, text string) {
	h.publish <- NewInfoMessage(room, client, text)
}

func (h *Hub) sendWelcomeMessage(client *Client, room *Room) {
	client.send <- NewWelcomeMessage(*room.Name, room.getActiveClients())
}

func (h *Hub) run() {
	for {
		select {

		case client := <-h.register:
			room := getOrCreateRoom(client.room.Name)
			client.room = room
			room.register(client)
			h.sendWelcomeMessage(client, room)
			go h.publishInfo(client.room, client, "registered")

		case client := <-h.unregister:
			close(client.send)
			room := client.room
			room.unregister(client)
			if room.isEmpty() {
				room.close()
			} else {
				go h.publishInfo(client.room, client, "unregistered")
			}

		case msg := <-h.publish:
			room := msg.Data.Room
			content := marshalMessage(msg)
			publish(room.Name, content)

		case msg := <-h.broadcast:
			room := msg.Data.Room
			room.Lock()
			log.Debugf("broadcasting to room %s", *room.Name)
			for client := range room.clients {
				select {
				case client.send <- msg:
				default:
					// If anything hangs or fails, we close the channel
					room.unregister(client)
					if room.isEmpty() {
						room.close()
					} else {
						go h.publishInfo(client.room, client, "unregistered")
					}
				}
			}
			room.Unlock()
		}
	}
}
