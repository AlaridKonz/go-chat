package payloads

type Chat struct {
	Meta    ChatMetadata `json:"meta"`
	Content string       `json:"content"`
}

type ChatMetadata struct {
	ChatId    int    `json:"chatid"`
	UserId    int    `json:"userid"`
	Timestamp string `json:"timestamp"`
	Members   []int  `json:"members"`
}
