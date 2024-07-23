package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var rooms = make(map[string]*Room)
var mutex = sync.RWMutex{}

var cacheKey = "active_clients"

type Room struct {
	sync.RWMutex

	id            *string               `json:"-"`
	Name          *string               `json:"name"`
	clients       map[*Client]bool      `json:"-"`
	subscription  *redis.PubSub         `json:"-"`
	listener      <-chan *redis.Message `json:"-"`
	ActiveClients *[]Client             `json:"activeClients,omitempty"`
}

func ensureRoomCacheEntry(roomName *string) {
	root := "$"
	cacheRoot := getCache(cacheKey, root)
	if len(*cacheRoot) == 0 || *cacheRoot == "[]" {
		setCache(cacheKey, root, "{}")
	}

	roomRoot := fmt.Sprintf("$.%s", *roomName)
	cachedClients := getCache(cacheKey, roomRoot)
	if len(*cachedClients) == 0 || *cachedClients == "[]" {
		setCache(cacheKey, roomRoot, "{}")
	}
}

func getOrCreateRoom(name *string) *Room {
	mutex.RLock()
	room, exists := rooms[*name]
	mutex.RUnlock()
	if exists {
		return room
	}

	sub := subscribe(name)

	mutex.Lock()
	id := uuid.New().String()
	room = &Room{
		id:           &id,
		Name:         name,
		clients:      make(map[*Client]bool),
		subscription: sub,
		listener:     sub.Channel(),
	}
	rooms[*name] = room

	mutex.Unlock()

	ensureRoomCacheEntry(name)
	go room.listen()
	return room
}

func (r *Room) listen() {
	for {
		select {
		case message := <-r.listener:
			msg := unmarshalMessage([]byte(message.Payload))
			msg.Data.Room = r
			hub.broadcast <- msg
		}
	}
}

func (r *Room) unregister(client *Client) {
	r.Lock()
	defer r.Unlock()
	delete(r.clients, client)

	path := fmt.Sprintf("$.%s.%s", *r.Name, client.Id)
	removeCache(cacheKey, path)
}

func (r *Room) register(client *Client) {
	r.Lock()
	defer r.Unlock()
	r.clients[client] = true

	path := fmt.Sprintf("$.%s.%s", *r.Name, client.Id)
	setCache(cacheKey, path, client)
}

func (r *Room) getActiveClients() *[]Client {
	var clients []Client
	val := getCache(cacheKey, *r.Name)

	var clientsMap map[string]Client
	err := json.Unmarshal([]byte(*val), &clientsMap)
	if err != nil {
		log.Error(err)
		return &clients
	}

	for _, client := range clientsMap {
		clients = append(clients, client)
	}
	return &clients
}

func (r *Room) close() {
	r.RLock()
	if len(r.clients) != 0 {
		r.RUnlock()
		return
	}

	r.RUnlock()
	mutex.Lock()
	delete(rooms, *r.Name)
	mutex.Unlock()

	path := fmt.Sprintf("$.%s", *r.Name)
	if *getCache(cacheKey, path) == "[{}]" {
		removeCache(cacheKey, path)
	}

	r.subscription.Unsubscribe(ctx, *r.Name)
}

func (r *Room) isEmpty() bool {
	return len(r.clients) == 0
}
