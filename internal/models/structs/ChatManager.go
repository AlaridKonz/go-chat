package structs

import (
	"encoding/json"
	"gochat/internal/models/payloads"
	ds "gochat/internal/utils/datastructures"
	ws "gochat/internal/utils/websocket"

	"github.com/gorilla/websocket"
)

type ChatRooms map[int]ds.XList[*payloads.ChatMessage]

func (cr *ChatRooms) Get(chatroomid int) *ds.XList[*payloads.ChatMessage] {
	_, ok := (*cr)[chatroomid]
	if !ok {
		(*cr)[chatroomid] = *new(ds.XList[*payloads.ChatMessage]).Init()
	}
	chatroom := (*cr)[chatroomid]
	return &chatroom
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
	successfullySent := new(ds.XList[*payloads.ChatMessage])
	chatrooms := cm.Get(userid)
	chatMessages := chatrooms.Get(chatroomid)
	chatMessages.ForEach(func(cm *payloads.ChatMessage) {
		jsonStr, err := json.Marshal(cm)
		if err != nil {
			return
		}
		if err := ws.SendMessage(conn, jsonStr); err != nil {
			return
		}
		successfullySent.PushFront(cm)
	})

	chatMessages.RemoveIf(successfullySent.Contains)
}
