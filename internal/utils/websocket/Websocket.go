package websocket

import "github.com/gorilla/websocket"

func SendMessage(conn *websocket.Conn, message []byte) error {
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return err
	}
	return nil
}
