package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"td/internal/model"
	"td/internal/tdlib"
	"time"

	"google.golang.org/genai"
)

func ReactMessages(
	clientId int,
	callback func(chatId int64, limit uint64),
	messages chan []model.Message,
	reactChan chan model.ReactionsResp,
	hostel *model.Hostel,
	client *genai.Client,
	ctx context.Context,
	fileChan chan model.TdFile,
) {
	hostel.Mutex.Lock()
	defer hostel.Mutex.Unlock()

	var chatIdStr string
	fmt.Print("Введите id чата у которого хотите поставить реакции: ")
	fmt.Scan(&chatIdStr)
	re := regexp.MustCompile(`[0-9]`)
	chatId, _ := strconv.ParseInt(strings.Join(re.FindAllString(chatIdStr, -1), ""), 10, 64)

	var limitStr string
	fmt.Print("На сколько последних сообщений хотите поставить реакции? (не более 100): ")
	fmt.Scan(&limitStr)
	limit, _ := strconv.ParseUint(strings.Join(re.FindAllString(limitStr, -1), ""), 10, 64)

	callback(chatId, limit)

	getChatHistory := fmt.Sprintf(`{
		"@type": "getChatHistory",
		"chat_id": %d,
		"from_message_id": 0,
		"offset": 0,
		"limit": %d,
		"only_local": false
	}`, chatId, limit)
	tdlib.Send(clientId, getChatHistory)
	time.Sleep(1 * time.Second)
	tdlib.Send(clientId, getChatHistory)

	reactResp := <-reactChan
	var rawReactions []model.CommonReaction
	var (
		emojiReaction  model.EmojiReactions
		emojiReactions []model.EmojiReaction
	)
	rawReactions = append(reactResp.Resp.Top, reactResp.Resp.Popular...)
	rawReactions = append(rawReactions, reactResp.Resp.Recent...)
	for i, r := range rawReactions {
		if !r.ForVIP && r.Type.Type == "reactionTypeEmoji" {
			if len(emojiReactions) < 1 {
				json.Unmarshal(reactResp.Req, &emojiReaction)
				emojiReactions = append(emojiReaction.Top, emojiReaction.Popular...)
				emojiReactions = append(emojiReactions, emojiReaction.Recent...)
			}
			*hostel.Reactions = append(*hostel.Reactions, emojiReactions[i].Type.Emoji)
		}
	}

	msgs := <-messages
	reactions := *hostel.Reactions
	text := "bro i so tired, tomorrow ok?"
	for _, msg := range msgs {
		if _, ok := hostel.ReactedMsgs[msg.ChatId]; !ok {
			hostel.ReactedMsgs[msg.ChatId] = false
		}
		if !hostel.ReactedMsgs[msg.Id] {
			parts := []*genai.Part{{
				Text: fmt.Sprintf(`
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
				`, strings.Join(reactions, ", "), text),
			}}
			result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{{Parts: parts}}, nil)
			if err != nil {
				fmt.Println("Ошибка получения результата от AI: ", err)
			}
			SendReaction(clientId, chatId, msg.Id, result.Text(), "reactions")
			hostel.ReactedMsgs[msg.Id] = true
		}
	}
}

func ReactNewMessage(
	clientId int,
	msgChan chan model.LastMsgResp,
	hostel *model.Hostel,
	client *genai.Client,
	ctx context.Context,
	fileChan chan model.TdFile,
) {
	var (
		reactions []string
		parts     []*genai.Part
	)
	for {
		message := <-msgChan
		parts = []*genai.Part{}
		hostel.Mutex.Lock()
		if _, ok := hostel.ReactedMsgs[message.LastMsg.Id]; !ok {
			hostel.ReactedMsgs[message.LastMsg.Id] = false
		}
		if !hostel.ReactedMsgs[message.LastMsg.Id] {
			reactions = *hostel.Reactions
			switch v := message.LastMsg.Content.(type) {
			case model.MessageText:
				parts = append(parts, &genai.Part{Text: GetPrompt(reactions, v.Formatted.Text)})
			case model.MessagePhoto:
				parts = append(parts, &genai.Part{})
			}
			result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{{Parts: parts}}, nil)
			if err != nil {
				fmt.Println("Ошибка получения результата от AI: ", err)
			}
			SendReaction(clientId, message.ChatId, message.LastMsg.Id, result.Text(), "reaction")
			hostel.ReactedMsgs[message.LastMsg.Id] = true
		}
		hostel.Mutex.Unlock()
	}
}
