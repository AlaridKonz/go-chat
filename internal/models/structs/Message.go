package structs

import (
	"gochat/internal/models/payloads"
	"gochat/internal/utils"
)

type ChatNotification struct {
	ChatMessage          payloads.Chat
	UnnotifiedRecipients *utils.XSet[int]
}

func NewChatNotification(chat payloads.Chat) *ChatNotification {
	return &ChatNotification{
		ChatMessage:          chat,
		UnnotifiedRecipients: utils.NewXSet[int](chat.Meta.Members...),
	}
}

func (cn *ChatNotification) Notified(recipient int) {
	cn.UnnotifiedRecipients.Remove(recipient)
}

func (cn *ChatNotification) HasNotifiedAll() bool {
	return cn.UnnotifiedRecipients.IsEmpty()
}
