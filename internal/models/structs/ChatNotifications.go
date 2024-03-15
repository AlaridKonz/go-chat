package structs

import (
	"encoding/json"
	"gochat/internal/models/payloads"
	ds "gochat/internal/utils/datastructures"
	ws "gochat/internal/utils/websocket"

	"github.com/gorilla/websocket"
)

type ChatNotification struct {
	ChatMessage          payloads.ChatMessage
	UnnotifiedRecipients *ds.XSet[int]
}

func NewChatNotification(chat payloads.ChatMessage) *ChatNotification {
	return &ChatNotification{
		ChatMessage:          chat,
		UnnotifiedRecipients: ds.NewXSet[int](chat.Meta.Members...),
	}
}

func (cn *ChatNotification) Notified(recipient int) {
	cn.UnnotifiedRecipients.Remove(recipient)
}

func (cn *ChatNotification) HasNotifiedAll() bool {
	return cn.UnnotifiedRecipients.IsEmpty()
}

func (cn *ChatNotification) NotifyUser(conn *websocket.Conn, recipientid int) error {
	if !cn.UnnotifiedRecipients.Contains(recipientid) {
		return nil
	}
	jsonStr, err := json.Marshal(cn.ChatMessage)
	if err != nil {
		return err
	}
	if err := ws.SendMessage(conn, jsonStr); err != nil {
		return err
	}
	cn.Notified(recipientid)
	return nil
}
