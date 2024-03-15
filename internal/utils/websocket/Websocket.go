package utils

import "github.com/gorilla/websocket"

func Notify(conn *websocket.Conn, message []byte) error {
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return err
	}
	return nil
}
