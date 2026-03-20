package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"td/internal/model"
	"td/internal/tdlib"
)

func SendReaction(clientId int, chatId int64, msgId int64, emoji string, extra string) {
	tdlib.Send(clientId, fmt.Sprintf(`{
		"@type": "addMessageReaction",
		"chat_id": %d,
		"message_id": %d,
		"reaction_type": {
			"@type": "reactionTypeEmoji",
			"emoji": "%s"
		},
		"is_big": false,
		"update_recent_reactions": false,
		"@extra": "%s"
	}`, chatId, msgId, emoji, extra))
}

func PrepareAndSendReaction() {}

func StartDownload(
	ClientId int,
	FileId int32,
) {
	tdlib.Send(
		ClientId,
		fmt.Sprintf(`{
			"@type": "downloadFile",
			"file_id": %d,
			"priority": 1,
			"offset": 0,
			"limit": 0,
			"synchronous": true
		}`, FileId),
	)
}

type MessageWrapper struct {
	Content model.MessageContent
}

func (r *MessageWrapper) UnmarshalJSON(data []byte) error {
	var msg model.CommonResp
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case "messageText":
		var msgText model.MessageText
		json.Unmarshal(data, &msgText)
		r.Content = msgText
	case "messagePhoto":
		var msgPhoto model.MessagePhoto
		json.Unmarshal(data, &msgPhoto)
		r.Content = msgPhoto
	case "messageDocument":
		var msgDoc model.MessageDocument
		json.Unmarshal(data, &msgDoc)
		r.Content = msgDoc
	case "messageVideo":
		var msgVideo model.MessageVideo
		json.Unmarshal(data, &msgVideo)
		r.Content = msgVideo
	}
	return nil
}

func GetPrompt(
	reactions []string,
	text string,
) string {
	return fmt.Sprintf(`
		You are reacting to a message in a chat, exactly like a real person would — choosing the single most fitting emoji reaction.

		You will receive:
		1. A list of available reaction emojis
		2. A message to react to — this could be text, an image, a video thumbnail, a sticker, a voice message description, a document, or any other type of content. If media is attached, it will be provided directly.

		Look at EVERYTHING provided — read the text, look at the image, understand the vibe — then pick your reaction.

		Your job: choose the ONE emoji from the list that best fits your reaction — emotionally, visually, or contextually.

		Strict rules:
		- Output ONLY a single emoji character — nothing else
		- The emoji MUST come from the provided list — never use emojis outside of it
		- No words, no punctuation, no explanation, no line breaks — just the emoji

		Available reactions:
		%s

		Message / Content to react to:
		%s
	`, strings.Join(reactions, ","), text)
}
