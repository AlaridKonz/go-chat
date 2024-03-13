package main

import (
	"container/list"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"gochat/internal/models/payloads"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connections map[int]*websocket.Conn = make(map[int]*websocket.Conn)

// TODO: Implement XList, use generics, and use XList[Message]
// TODO: Or even make a larger struct that handles all the pending messages inside instead of a simple map
var pendingMessages = make(map[int]list.List)
var pendingNotifs = make(map[int]list.List)

func main() {
	http.HandleFunc("/notify", newSubscriptionHandler)
	http.HandleFunc("/chat", openChatHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func newSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var sub payloads.Subscription
	if err := conn.ReadJSON(&sub); err != nil {
		log.Println(err)
		return
	}

	// future: clean previous dead connections
	// for now I'll use one connection per user
	connections[sub.UserId] = conn

	if err := conn.WriteMessage(websocket.TextMessage, []byte("subscribed")); err != nil {
		log.Println(err)
		return
	}

	if list, ok := pendingNotifs[sub.UserId]; ok && list.Len() > 0 {
		sendPending(conn, &list)
	}
}

func openChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	chatid, _ := connectToChat(conn)
	chatLifeHandler(conn, chatid)
}

func connectToChat(conn *websocket.Conn) (int, error) {
	var chat payloads.Chat
	if err := conn.ReadJSON(&chat); err != nil {
		log.Println(err)
		return -1, err
	}
	return chat.Meta.ChatId, nil
}

func chatLifeHandler(conn *websocket.Conn, chatid int) {
	for {
		if list, ok := pendingMessages[chatid]; ok && list.Len() > 0 {
			sendPending(conn, &list)
		}
		var msg payloads.Chat
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println(err)
			return
		}
		if err := conn.WriteJSON(msg); err != nil {
			log.Println(err)
			return
		}
		addToPending(pendingMessages, chatid, msg)
		for _, recipient := range msg.Meta.Members {
			if recipient == msg.Meta.UserId {
				continue
			}
			notify(recipient, chatid)
		}
	}
}

func notify(userid int, chatid int) {
	conn := connections[userid]
	notification := struct{ Content string }{
		Content: fmt.Sprintf("update on chat %d", chatid),
	}
	if err := conn.WriteJSON(notification); err != nil {
		log.Println("Could not send notification")
		addToPending(pendingNotifs, userid, notification)

		return
	}
}

func addToPending(pendingMap map[int]list.List, key int, json interface{}) {
	if _, ok := pendingMap[key]; !ok {
		pendingMap[key] = *list.New()
	}
	list := pendingMap[key]
	list.PushBack(json)
}

func sendPending(conn *websocket.Conn, pendingList *list.List) {
	for el := pendingList.Front(); pendingList.Len() > 0; el = pendingList.Front() {
		output, err := el.Value.(string)
		if err {
			log.Println("not a string")
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(output)); err != nil {
			log.Println("error writing message")
		}
		pendingList.Remove(pendingList.Front())
	}
}
