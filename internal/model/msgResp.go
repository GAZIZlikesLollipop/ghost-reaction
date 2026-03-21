package model

import "encoding/json"

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

type RawMessage struct {
	Id      int64      `json:"id"`
	ChatId  int64      `json:"chat_id"`
	Content CommonResp `json:"content"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	var msg RawMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	switch msg.Content.Type {
	case "messageText":
		var msgCnt MessageText
		json.Unmarshal(data, &msgCnt)
		*m = Message{
			Id:      m.Id,
			ChatId:  m.ChatId,
			Content: msgCnt,
		}
	case "messagePhoto":
		var msgCnt MessagePhoto
		json.Unmarshal(data, &msgCnt)
		*m = Message{
			Id:      m.Id,
			ChatId:  m.ChatId,
			Content: msgCnt,
		}
	case "messageDocument":
		var msgCnt MessageDocument
		json.Unmarshal(data, &msgCnt)
		*m = Message{
			Id:      m.Id,
			ChatId:  m.ChatId,
			Content: msgCnt,
		}
	case "messageVideo":
		var msgCnt MessageVideo
		json.Unmarshal(data, &msgCnt)
		*m = Message{
			Id:      m.Id,
			ChatId:  m.ChatId,
			Content: msgCnt,
		}
	}
	return nil
}
