package model

import (
	"encoding/json"
	"fmt"
)

type RawMessage struct {
	Id      int64      `json:"id"`
	ChatId  int64      `json:"chat_id"`
	Content CommonResp `json:"content"`
}

type MessageText struct {
	Content MsgCntText `json:"content"`
}

type MessagePhoto struct {
	Content MsgCntPhoto `json:"content"`
}

type MessageDoc struct {
	Content MsgCntDoc `json:"content"`
}

type MessageVideo struct {
	Content MsgCntText `json:"content"`
}

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

type RawMessages struct {
	Msgs []RawMessage `json:"messages"`
}

type TextMessages struct {
	Msgs []MessageText `json:"messages"`
}

type PhotoMessages struct {
	Msgs []MessagePhoto `json:"messages"`
}

type DocMessages struct {
	Msgs []MessageDoc `json:"messages"`
}

type VideoMessages struct {
	Msgs []MessageVideo `json:"messages"`
}

func (m *MsgsResp) UnmarshalJSON(data []byte) error {
	var msgs RawMessages
	if err := json.Unmarshal(data, &msgs); err != nil {
		return err
	}
	var result []Message
	for _, msg := range msgs.Msgs {
		switch msg.Content.Type {
		case "messageText":
			var msgCnt TextMessages
			if err := json.Unmarshal(data, &msgCnt); err != nil {
				fmt.Println(err)
				return err
			}
			for _, message := range msgCnt.Msgs {
				result = append(result, Message{
					Id:      msg.Id,
					ChatId:  msg.ChatId,
					Content: message.Content,
				})
			}
		case "messagePhoto":
			var msgCnt PhotoMessages
			if err := json.Unmarshal(data, &msgCnt); err != nil {
				fmt.Println(err)
				return err
			}
			for _, message := range msgCnt.Msgs {
				result = append(result, Message{
					Id:      msg.Id,
					ChatId:  msg.ChatId,
					Content: message.Content,
				})
			}
		case "messageDocument":
			var msgCnt DocMessages
			if err := json.Unmarshal(data, &msgCnt); err != nil {
				fmt.Println(err)
				return err
			}
			for _, message := range msgCnt.Msgs {
				result = append(result, Message{
					Id:      msg.Id,
					ChatId:  msg.ChatId,
					Content: message.Content,
				})
			}
		case "messageVideo":
			var msgCnt VideoMessages
			if err := json.Unmarshal(data, &msgCnt); err != nil {
				fmt.Println(err)
				return err
			}
			for _, message := range msgCnt.Msgs {
				result = append(result, Message{
					Id:      msg.Id,
					ChatId:  msg.ChatId,
					Content: message.Content,
				})
			}
		default:
			result = append(result, Message{
				Id:      msg.Id,
				ChatId:  msg.ChatId,
				Content: MsgCntText{Formatted: FormattedText{Text: "Like this message"}},
			})
		}
	}
	*m = MsgsResp{Msgs: result}
	return nil
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
			Id:      msg.Id,
			ChatId:  msg.ChatId,
			Content: msgCnt.Content,
		}
	case "messagePhoto":
		var msgCnt MessagePhoto
		json.Unmarshal(data, &msgCnt)
		*m = Message{
			Id:      msg.Id,
			ChatId:  msg.ChatId,
			Content: msgCnt.Content,
		}
	case "messageDocument":
		var msgCnt MessageDoc
		json.Unmarshal(data, &msgCnt)
		*m = Message{
			Id:      msg.Id,
			ChatId:  msg.ChatId,
			Content: msgCnt.Content,
		}
	case "messageVideo":
		var msgCnt MessageVideo
		json.Unmarshal(data, &msgCnt)
		*m = Message{
			Id:      msg.Id,
			ChatId:  msg.ChatId,
			Content: msgCnt.Content,
		}
	default:
		*m = Message{
			Id:      msg.Id,
			ChatId:  msg.ChatId,
			Content: MsgCntText{Formatted: FormattedText{Text: "Like this message"}},
		}
	}
	return nil
}
