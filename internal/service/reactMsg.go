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
	re := regexp.MustCompile(`-?[0-9]`)
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
	if limit > 1 {
		time.Sleep(1 * time.Second)
		tdlib.Send(clientId, getChatHistory)
	}

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
	for _, msg := range msgs {
		message := model.LastMsgResp{
			LastMsg: msg,
			ChatId:  msg.ChatId,
		}
		PrepareAndSendReaction(
			clientId,
			hostel,
			message,
			fileChan,
			client,
			ctx,
			true,
		)
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
	for {
		message := <-msgChan
		PrepareAndSendReaction(
			clientId,
			hostel,
			message,
			fileChan,
			client,
			ctx,
			false,
		)
	}
}
