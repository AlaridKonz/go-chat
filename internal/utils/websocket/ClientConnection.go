package websocket

import "github.com/gorilla/websocket"

type ClientConnection struct {
	Conn     *websocket.Conn
	ClientId int
	ChatId   int
}
