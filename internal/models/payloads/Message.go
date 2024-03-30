package payloads

import "time"

type RawMessage struct {
	Action   string   `json:"action"`
	Metadata Metadata `json:"metadata"`
	Data     any      `json:"data"`
}

type Metadata struct {
	UserId    int       `json:"userid"`
	DeviceId  int       `json:"deviceid"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}

type NewChat struct {
	ChatName string `json:"chat_name"`
	Members  []int  `json:"members"`
}

type NewChatResponse struct {
	ChatId int `json:"chatid"`
	NewChat
}

type TextMessage struct {
	ChatId int    `json:"chatid"`
	Text   string `json:"text"`
}

type TMessage[T any] struct {
	Action   string   `json:"action"`
	Metadata Metadata `json:"metadata"`
	Data     T        `json:"data"`
}

type RawResponse struct {
	RawMessage
}
