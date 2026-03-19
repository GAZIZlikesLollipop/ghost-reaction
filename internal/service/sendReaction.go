package service

import (
	"fmt"
	"td/internal/tdlib"
	"time"
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
	time.Sleep(1 * time.Second)
}
