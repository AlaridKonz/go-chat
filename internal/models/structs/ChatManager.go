package structs

import (
	"encoding/json"
	"fmt"
	"gochat/internal/models/payloads"
	ds "gochat/internal/utils/datastructures"
	ws "gochat/internal/utils/websocket"
	"log"

	"github.com/gorilla/websocket"
)

type ChatRooms map[int]*ds.XList[*payloads.ChatMessage]

func (cr *ChatRooms) Get(chatroomid int) *ds.XList[*payloads.ChatMessage] {
	_, ok := (*cr)[chatroomid]
	if !ok {
		(*cr)[chatroomid] = new(ds.XList[*payloads.ChatMessage]).Init()
	}
	chatroom := (*cr)[chatroomid]
	log.Printf("list tostring: %s", chatroom.ToString())
	return chatroom
}

type ChatManager struct {
	Users map[int]ChatRooms
}

func NewChatNotificationManager() *ChatManager {
	return new(ChatManager).Init()
}

func (cm *ChatManager) Init() *ChatManager {
	return &ChatManager{
		Users: make(map[int]ChatRooms),
	}
}

func (cm *ChatManager) backupInit() *ChatManager {
	if cm.Users == nil {
		return cm.Init()
	}
	return cm
}

func (cm *ChatManager) AddNewChatMessage(newChat *payloads.ChatMessage) {
	cm.backupInit()
	for _, user := range newChat.Meta.Members {
		chatrooms := cm.Get(user)
		chatMessages := chatrooms.Get(newChat.Meta.ChatId)
		chatMessages.PushBack(newChat)
		chatjson, _ := json.Marshal(newChat)
		log.Println(string(chatjson))
		cmjson, _ := json.Marshal(cm)
		log.Println(string(cmjson))
		log.Printf("Pushed new message for user %d, in chatroom %d. Now the size is %d", user, newChat.Meta.ChatId, chatMessages.Size())
	}
}

func (cm *ChatManager) Get(userid int) *ChatRooms {
	_, ok := cm.Users[userid]
	if !ok {
		cm.Users[userid] = make(ChatRooms)
	}
	chatroom := cm.Users[userid]
	return &chatroom
}

func (cm *ChatManager) NeedsFetching(userid int, chatroomid int) bool {
	chatrooms := cm.Get(userid)
	chatMessages := chatrooms.Get(chatroomid)
	return !chatMessages.IsEmpty()
}

func (cm *ChatManager) NotifyUser(userid int, chatroomid int, conn *websocket.Conn) {

	log.Printf("Attempting to notify user %d on chatroom %d", userid, chatroomid)
	successfullySent := new(ds.XList[*payloads.ChatMessage])
	chatrooms := cm.Get(userid)
	chatMessages := chatrooms.Get(chatroomid)
	log.Printf("chatMessages size is %d", chatMessages.Size())

	chatMessages.ForEach(func(cm **payloads.ChatMessage) {
		log.Printf("entered foreach")

		jsonStr, err := json.Marshal(cm)
		if err != nil {
			log.Printf("Error on marshalling")
			return
		}
		if err := ws.SendMessage(conn, jsonStr); err != nil {
			log.Printf("Error on sending message %s", jsonStr)
			return
		}
		successfullySent.PushFront(*cm)
		log.Printf("Notified user %d on chatroom %d", userid, chatroomid)
	})
	log.Printf("messages: %d", chatMessages.Size())
	log.Printf("successful: %d", successfullySent.Size())
	chatMessages.RemoveIf(successfullySent.Contains)
	log.Printf("messages after: %d", chatMessages.Size())
}

func (cm *ChatManager) Debug() string {
	str := "{\n"
	for userid, chatrooms := range cm.Users {
		str += fmt.Sprintf("\t%d\n", userid)
		for chatroomid, chatmessages := range chatrooms {
			str += fmt.Sprintf("\t\t%d\n", chatroomid)
			str += fmt.Sprintf("\t\t\t%s\n", chatmessages.ToString())
		}
	}
	return str + "\n}"

}
