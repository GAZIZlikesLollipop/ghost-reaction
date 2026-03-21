package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"td/internal/model"
	"td/internal/tdlib"

	"google.golang.org/genai"
)

func SendReaction(
	clientId int,
	chatId int64,
	msgId int64,
	emoji string,
	extra string,
) {
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

func StartDownload(
	ClientId int,
	FileId int32,
	Limit int64,
) {
	tdlib.Send(
		ClientId,
		fmt.Sprintf(`{
			"@type": "downloadFile",
			"file_id": %d,
			"priority": 1,
			"offset": 0,
			"limit": %d,
			"synchronous": true
		}`, FileId, Limit),
	)
}

func PrepareAndSendReaction(
	clientId int,
	hostel *model.Hostel,
	message model.LastMsgResp,
	fileChan chan model.TdFile,
	client *genai.Client,
	ctx context.Context,
	isCustom bool,
) {
	var (
		parts     []*genai.Part
		reactions []string
	)
	if !isCustom {
		hostel.Mutex.Lock()
		defer hostel.Mutex.Unlock()
	}
	if _, ok := hostel.ReactedMsgs[message.LastMsg.Id]; !ok {
		hostel.ReactedMsgs[message.LastMsg.Id] = false
	}
	if !hostel.ReactedMsgs[message.LastMsg.Id] {
		reactions = *hostel.Reactions
		switch v := message.LastMsg.Content.(type) {
		case model.MessageText:
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Formatted.Text)})
		case model.MessagePhoto:
			for _, size := range v.PhotoInfo.PhotoSizes {
				if size.Type == "x" {
					StartDownload(clientId, size.File.Id, 0)
				}
			}
			file := <-fileChan
			data, err := os.ReadFile(file.Local.Path)
			if err != nil {
				fmt.Println("Ошибка чтения файла фотографии: ", err)
				return
			}
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Caption.Text)})
			parts = append(parts, &genai.Part{InlineData: &genai.Blob{
				Data: data,
			}})
		case model.MessageDocument:
			StartDownload(clientId, v.Document.File.Id, 0)
			file := <-fileChan
			data, err := os.ReadFile(file.Local.Path)
			if err != nil {
				fmt.Println("Ошибка чтения документа: ", err)
				return
			}
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Caption.Text)})
			parts = append(parts, &genai.Part{InlineData: &genai.Blob{
				Data:        data,
				DisplayName: v.Document.FileName,
				MIMEType:    v.Document.MimeType,
			}})
		case model.MessageVideo:
			StartDownload(clientId, v.Video.Video.Id, 103809024)
			file := <-fileChan
			data, err := os.ReadFile(file.Local.Path)
			if err != nil {
				fmt.Println("Ошибка чтения файла видео: ", err)
				return
			}
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Caption.Text)})
			parts = append(parts, &genai.Part{InlineData: &genai.Blob{
				Data: data,
			}})
		}
		result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{{Parts: parts}}, nil)
		if err != nil {
			fmt.Println("Ошибка получения результата от AI: ", err)
			return
		}
		SendReaction(clientId, message.ChatId, message.LastMsg.Id, result.Text(), "reaction")
		hostel.ReactedMsgs[message.LastMsg.Id] = true
	}
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
