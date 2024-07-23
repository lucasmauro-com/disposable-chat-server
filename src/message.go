package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type InputMessage struct {
	Text string `json:"text"`
}

type Data struct {
	Room   *Room   `json:"room,omitempty"`
	Client *Client `json:"client,omitempty"`
	Action string  `json:"action,omitempty"`
	Text   string  `json:"text,omitempty"`
}

type OutputMessage struct {
	Type string `json:"type,omitempty"`
	Data *Data  `json:"data,omitempty"`
}

func NewWelcomeMessage(roomName string, clients *[]Client) *OutputMessage {
	return &OutputMessage{
		Type: "welcome",
		Data: &Data{
			Room: &Room{
				Name:          &roomName,
				ActiveClients: clients,
			},
		},
	}
}

func NewInfoMessage(room *Room, client *Client, action string) *OutputMessage {
	return &OutputMessage{
		Type: "info",
		Data: &Data{Room: room, Client: client, Action: action},
	}
}

func NewChatMessage(room *Room, client *Client, text string) *OutputMessage {
	return &OutputMessage{
		Type: "chat",
		Data: &Data{Room: room, Client: client, Text: text},
	}
}

func NewErrorMessage(text string) *OutputMessage {
	return &OutputMessage{
		Type: "error",
		Data: &Data{Text: text},
	}
}

func marshalMessage(msg *OutputMessage) []byte {
	message, err := json.Marshal(msg)
	if err != nil {
		log.Error(err)
	}
	return message
}

func unmarshalMessage(message []byte) *OutputMessage {
	var msg OutputMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		log.Error(err)
	}
	return &msg
}
