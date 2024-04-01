package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"

	"gochat/internal/models/payloads"
	"gochat/internal/models/structs"
	ds "gochat/internal/utils/datastructures"
	ws "gochat/internal/utils/websocket"
)

var connections map[int]*websocket.Conn = make(map[int]*websocket.Conn)
var chatManager = new(structs.ChatManager).Init()
var pendingNotifs = make(map[int]*ds.XList[interface{}])

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var conns map[*websocket.Conn]UserData = make(map[*websocket.Conn]UserData)

// var msgs map[int]payloads.RawResponse = make(map[int]payloads.RawResponse)
var test sync.Map = sync.Map{}

type UserData struct {
	DeviceId int
	UserId   int
}

var chats map[int]payloads.NewChat = make(map[int]payloads.NewChat)

type Members []int

const (
	Subscribe   = "sub"
	NewChat     = "newchat"
	SendMessage = "sendmsg"
)

func main() {
	http.HandleFunc("/sub", subHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func subHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn.SetCloseHandler(func(code int, text string) error {
		delete(conns, conn)
		return nil
	})
	go subMessageListener(conn)
}

func subMessageSpeaker(conn *websocket.Conn, userid int) {
	conn.WriteMessage(websocket.TextMessage, []byte("Connected"))
	for {
		time.Sleep(1 * time.Second)
		if _, exists := conns[conn]; !exists {
			return
		}
		if _, exists := test.Load(userid); exists {
			msg, exists := test.LoadAndDelete(userid)
			if !exists {
				log.Print(exists)
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Print(err)
			}
		}

	}
}

func subMessageListener(conn *websocket.Conn) {
	defer conn.Close()
	for {
		var sub payloads.RawMessage
		if err := conn.ReadJSON(&sub); err != nil {
			log.Println(err)
			return
		}

		switch sub.Action {
		case Subscribe:
			var metadata payloads.Metadata
			mapstructure.Decode(sub.Data, &metadata)
			if _, exists := conns[conn]; !exists {
				conns[conn] = UserData{DeviceId: metadata.DeviceId, UserId: metadata.UserId}
				go subMessageSpeaker(conn, metadata.UserId)
			}

		case NewChat:
			var newchat payloads.NewChat
			mapstructure.Decode(sub.Data, &newchat)
			newchatId := rand.Int()
			chats[newchatId] = newchat
			response := payloads.RawResponse{
				RawMessage: sub,
			}
			response.Data = payloads.NewChatResponse{
				ChatId:  newchatId,
				NewChat: newchat,
			}
			for _, id := range newchat.Members {
				test.Store(id, response)
			}

		case SendMessage:
			var text payloads.TextMessage
			mapstructure.Decode(sub.Data, &text)
			chatroom, exists := chats[text.ChatId]
			if !exists {
				continue
			}
			response := payloads.RawResponse{
				RawMessage: sub,
			}
			for _, id := range chatroom.Members {
				test.Store(id, response)
			}
		}
	}
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

	if list, ok := pendingNotifs[sub.UserId]; ok && list.Size() > 0 {
		sendPending(conn, list)
	}
}

func openChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	clientConn, _ := connectToChat(conn)
	chatLifeHandler(clientConn)
}

func connectToChat(conn *websocket.Conn) (*ws.ClientConnection, error) {
	var chat payloads.ChatMessage
	if err := conn.ReadJSON(&chat); err != nil {
		log.Println(err)
		return nil, err
	}
	chatManager.AddNewChatMessage(&chat)
	for _, recipient := range chat.Meta.Members {
		if recipient == chat.Meta.UserId {
			continue
		}
		notify(recipient, chat.Meta.ChatId)
	}
	return &ws.ClientConnection{
		Conn:     conn,
		ClientId: chat.Meta.UserId,
		ChatId:   chat.Meta.ChatId,
	}, nil
}

func chatLifeHandler(clientConn *ws.ClientConnection) {
	for {
		chatManager.NotifyUser(clientConn.ClientId, clientConn.ChatId, clientConn.Conn)

		var msg payloads.ChatMessage
		log.Println("getting msg")

		chatManager.AddNewChatMessage(&msg)
		for _, recipient := range msg.Meta.Members {
			if recipient == msg.Meta.UserId {
				continue
			}
			notify(recipient, clientConn.ChatId)
		}

		if err := clientConn.Conn.ReadJSON(&msg); err != nil {
			log.Println(err)
			return
		}
		log.Println("got msg")
	}
}

func notify(userid int, chatid int) {
	conn := connections[userid]
	notification := struct{ Content string }{
		Content: fmt.Sprintf("update on chat %d", chatid),
	}
	if err := conn.WriteJSON(notification); err != nil {
		log.Println("Could not send notification")
		addToPendingNotifications(userid, notification)

		return
	}
}

func addToPendingNotifications(userid int, json interface{}) {
	if _, ok := pendingNotifs[userid]; !ok {
		pendingNotifs[userid] = new(ds.XList[interface{}]).Init()
	}
	list := pendingNotifs[userid]
	list.PushBack(json)
}

func sendPending[T interface{}](conn *websocket.Conn, pendingList *ds.XList[T]) {
	for el := pendingList.Front(); pendingList.Size() > 0; el = pendingList.Front() {
		output, err := json.Marshal(el)
		if err != nil {
			log.Println("not a string")
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(output)); err != nil {
			log.Println("error writing message")
		}
		pendingList.Remove(pendingList.Front())
	}
}
