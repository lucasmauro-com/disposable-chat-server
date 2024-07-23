package main

import (
	"encoding/json"
	"net/http"
	"os"

	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == os.Getenv("ACCEPTED_ORIGIN")
	},
}

type Client struct {
	Id   string              `json:"id"   redis:"id"`
	Name string              `json:"name" redis:"name"`
	room *Room               `json:"-"    redis:"-"`
	conn *websocket.Conn     `json:"-"    redis:"-"`
	send chan *OutputMessage `json:"-"    redis:"-"`
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func serveWs(name string, roomName string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("failed to serve client (headers: %+v) (error: %+v)", r.Header, err)
		w.Write([]byte(err.Error()))
		return
	}

	client := &Client{
		Id:   uuid.New().String(),
		Name: name,
		room: &Room{Name: &roomName},
		conn: conn,
		send: make(chan *OutputMessage, 1024),
	}

	hub.register <- client

	go client.write()
	go client.read()
}

func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(marshalMessage(msg))

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) read() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, text, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		msg := &InputMessage{}

		json.Unmarshal(text, msg)

		invalidations := []string{}
		if len(msg.Text) > 300 {
			invalidations = append(invalidations, "'text' must not contain over 300 characters")
		}

		if len(msg.Text) == 0 {
			invalidations = append(invalidations, "'text' must not be empty")
		}

		if len(invalidations) == 0 {
			hub.publish <- NewChatMessage(c.room, c, msg.Text)
		} else {
			invalidationsStr := strings.Join(invalidations, ", ")
			c.send <- NewErrorMessage(invalidationsStr)
		}
	}
}
