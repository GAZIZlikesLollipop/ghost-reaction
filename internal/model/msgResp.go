package model

type Message struct {
	Id      int64          `json:"id"`
	ChatId  int64          `json:"chat_id"`
	Content MessageContent `json:"content"`
}

type LastMsgResp struct {
	ChatId  int64   `json:"chat_id"`
	LastMsg Message `json:"last_message"`
}

type MsgsResp struct {
	Msgs []Message `json:"messages"`
}
