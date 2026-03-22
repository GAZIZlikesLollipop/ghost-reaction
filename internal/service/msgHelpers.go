package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"td/internal/model"
	"td/internal/tdlib"
	"time"

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
	strDelay := strings.TrimSpace(os.Getenv("REQUEST_DELAY"))
	delay, err := strconv.Atoi(strDelay)
	if err != nil {
		time.Sleep(10 * time.Second)
	} else {
		time.Sleep(time.Duration(delay) * time.Second)
	}
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
		case model.MsgCntText:
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Formatted.Text)})
		case model.MsgCntPhoto:
			hasX := false
			if len(v.PhotoInfo.PhotoSizes) < 1 {
				fmt.Println("Invalid photo!")
				return
			}
			for _, size := range v.PhotoInfo.PhotoSizes {
				if size.Type == "x" {
					hasX = true
					StartDownload(clientId, size.File.Id, 103809024)
				}
			}
			if !hasX {
				StartDownload(clientId, v.PhotoInfo.PhotoSizes[0].File.Id, 103809024)
			}
			file := <-fileChan
			data, err := os.ReadFile(file.Local.Path)
			if err != nil {
				fmt.Println("Ошибка чтения файла фотографии: ", err)
				return
			}
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Caption.Text)})
			parts = append(parts, &genai.Part{InlineData: &genai.Blob{
				Data:     data,
				MIMEType: "image/jpeg",
			}})
		case model.MsgCntDoc:
			StartDownload(clientId, v.Document.File.Id, 103809024)
			document := <-fileChan
			data, err := os.ReadFile(document.Local.Path)
			if err != nil {
				fmt.Println("Ошибка чтения файла: ", err)
				return
			}
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Caption.Text)})
			parts = append(parts, &genai.Part{InlineData: &genai.Blob{
				Data:     data,
				MIMEType: "text/plain",
			}})
		case model.MsgCntVideo:
			StartDownload(clientId, v.Video.Video.Id, 20971520)
			video := <-fileChan
			file, err := os.Open(video.Local.Path)
			if err != nil {
				fmt.Println("Ошибка чтения документа: ", err)
				return
			}
			data, err := io.ReadAll(file)
			if err != nil {
				fmt.Println("Ошибка чтения файла: ", err)
				return
			}
			mimeType := mime.TypeByExtension(filepath.Ext(file.Name()))
			parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Caption.Text)})
			parts = append(parts, &genai.Part{InlineData: &genai.Blob{
				Data:     data,
				MIMEType: fmt.Sprintf("%s/%s", mimeType[:strings.Index(mimeType, "/")], mimeType[strings.Index(mimeType, "/")+1:]),
			}})
		}
		result, err := client.Models.GenerateContent(ctx, os.Getenv("AI_MODEL"), []*genai.Content{{Parts: parts}}, nil)
		if err != nil {
			fmt.Println("Ошибка получения результата от AI: ", err)
			time.Sleep(1 * time.Second)
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
