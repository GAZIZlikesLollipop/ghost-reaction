package model

type CommonResp struct {
	Type string `json:"@type"`
}

type AuthResp struct {
	State CommonResp `json:"authorization_state"`
}

type ErrorResp struct {
	Msg string `json:"message"`
}

type Message struct {
	Id      int64      `json:"id"`
	ChatId  int64      `json:"chat_id"`
	Content CommonResp `json:"content"`
}

type LastMsgResp struct {
	ChatId  int64   `json:"chat_id"`
	LastMsg Message `json:"last_message"`
}

type MsgsResp struct {
	Msgs []Message `json:"messages"`
}

type ExtraResp struct {
	Extra string `json:"@extra"`
}
