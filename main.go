package main

//#include "include/td_json_client.h"
//#cgo LDFLAGS: -L${SRCDIR}/lib -ltdjson -Wl,-rpath,${SRCDIR}
import "C"

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AuthState struct {
	Type string `json:"@type"`
}

type AuthResp struct {
	State AuthState `json:"authorization_state"`
}

type CommonResp struct {
	Type string `json:"@type"`
}

type ErrorResp struct {
	Msg string `json:"message"`
}

type Message struct {
	Id     int64 `json:"id"`
	ChatId int64 `json:"chat_id"`
}

type NewLastMsgResp struct {
	ChatId  int64   `json:"chat_id"`
	LastMsg Message `json:"last_message"`
}

type EmojiResp struct {
	Emojis []string `json:"emojis"`
}

type EmojiTypeResp struct {
	Emoji string `json:"emoji"`
}

type ReactionType struct {
	Type string `json:"@type"`
}

type ChatAvailableReactions struct {
	Type string `json:"@type"`
}

type ChatReactResp struct {
	ChatId    int64                  `json:"chat_id"`
	Reactions ChatAvailableReactions `json:"available_reactions"`
}

type ChatAvailableReactionsSome struct {
	Reactions []ReactionType `json:"reactions"`
}

type ReactionSomeResp struct {
	Reactions ChatAvailableReactionsSome `json:"available_reactions"`
}

type ChatEmojiReactionSome struct {
	Reactions []EmojiTypeResp `json:"reactions"`
}

type ReactionEmojiResp struct {
	Reactions ChatEmojiReactionSome `json:"available_reactions"`
}

type MsgsResp struct {
	Msgs []Message `json:"messages"`
}

func nextStep(
	clientId int,
	state string,
	isErr bool,
	onChatId func(data int64),
) error {
	switch state {
	case "authorizationStateWaitPhoneNumber":
		var phoneNumber string
		fmt.Print("Введите ваш номер телефона: ")
		fmt.Scan(&phoneNumber)
		C.td_send(
			C.int(clientId),
			C.CString(
				fmt.Sprintf(
					`{
						"@type": "setAuthenticationPhoneNumber",
						"phone_number": "%s"
					}`,
					phoneNumber,
				),
			),
		)
	case "authorizationStateWaitCode":
		var code string
		fmt.Print("Введите код потдверждения: ")
		fmt.Scan(&code)
		if len(code) < 5 {
			C.td_send(
				C.int(clientId),
				C.CString(
					`{
						"@type": "resendAuthenticationCode",
						"reason": {
							"@type": "resendCodeReasonUserRequest"
						}
					}`,
				),
			)
		} else {
			C.td_send(
				C.int(clientId),
				C.CString(
					fmt.Sprintf(
						`{
							"@type": "checkAuthenticationCode",
							"code": "%s"
						}`,
						code,
					),
				),
			)
		}
	case "authorizationStateWaitPassword":
		var password string
		fmt.Print("Введите пароль от аккаунта: ")
		fmt.Scan(&password)
		C.td_send(
			C.int(clientId),
			C.CString(
				fmt.Sprintf(
					`{
						"@type": "checkAuthenticationPassword",
						"password": "%s"
					}`,
					password,
				),
			),
		)
	case "authorizationStateReady":
		var chatIdStr string
		fmt.Print("Введите id чата у которого хотите поставить реакции: ")
		fmt.Scan(&chatIdStr)
		re := regexp.MustCompile(`[0-9]`)
		chatId, err := strconv.ParseInt(strings.Join(re.FindAllString(chatIdStr, -1), ""), 10, 64)
		if err != nil {
			return err
		}
		onChatId(chatId)
		C.td_send(
			C.int(clientId),
			C.CString(
				fmt.Sprintf(
					`{
						"@type": "getChat",
						"chat_id": %d
					}`,
					chatId,
				),
			),
		)
		fmt.Println("Загружаем чаты...")
		if !isErr {
			C.td_send(
				C.int(clientId),
				C.CString(
					`{
						"@type": "loadChats",	
						"chat_list": {
							"@type": "chatListMain"	
						},
						"limit": 100
					}`,
				),
			)
			C.td_send(
				C.int(clientId),
				C.CString(
					`{
						"@type": "loadChats",	
						"chat_list": {
							"@type": "chatListArchive"	
						},
						"limit": 100
					}`,
				),
			)
		}
		reInt := regexp.MustCompile(`[0-9]`)
		var limitStr string
		fmt.Print("На сколько последних сообщений хотите поставить реакции? (не более 100): ")
		fmt.Scan(&limitStr)
		limit, err := strconv.ParseInt(strings.Join(reInt.FindAllString(limitStr, -1), ""), 10, 64)
		if err != nil {
			return err
		}
		C.td_send(
			C.int(clientId),
			C.CString(
				fmt.Sprintf(
					`{
						"@type": "getChatHistory",	
						"chat_id": %d,
						"from_message_id": 0,
						"offset": 0,
						"limit": %d,
						"only_local": false,
						"@extra": "start"
					}`,
					chatId,
					limit,
				),
			),
		)
	}
	return nil
}

func main() {
	var (
		authState       string = "authorizationStateWaitTdlibParameters"
		selectedChat    int64
		activeReactions []string
	)
	availableReactions := make(map[int64][]string)
	var clientId = C.td_create_client_id()
	var timeout = C.double(1.0)
	var result = C.td_execute(
		C.CString(
			`{
				"@type": "setLogVerbosityLevel",
				"new_verbosity_level": 3
			}`,
		),
	)
	res := []byte(C.GoString(result))
	var resData CommonResp
	if err := json.Unmarshal(res, &resData); err != nil {
		panic(fmt.Sprintln("Ошибка преобразования json: ", err))
	}
	if resData.Type == "error" {
		var errResp ErrorResp
		if err := json.Unmarshal(res, &errResp); err != nil {
			panic(fmt.Sprintln("Ошибка преобразования json: ", err))
		}
		panic(fmt.Sprintln("Ошибка измнения потока логов: ", errResp.Msg))
	}
	if err := os.Remove("tdlib.log"); err != nil {
		panic(fmt.Sprintln("Ошибка удаления файла: ", err))
	}
	result = C.td_execute(
		C.CString(
			`{
				"@type": "setLogStream",	
				"log_stream": {
					"@type": "logStreamFile",	
					"path": "./tdlib.log",
					"max_file_size": 10485760,
					"redirect_stderr": true
				}
			}`,
		),
	)
	res = []byte(C.GoString(result))
	if err := json.Unmarshal(res, &resData); err != nil {
		panic(fmt.Sprintln("Ошибка преобразования json: ", err))
	}
	if resData.Type == "error" {
		var errResp ErrorResp
		if err := json.Unmarshal(res, &errResp); err != nil {
			panic(fmt.Sprintln("Ошибка преобразования json: ", err))
		}
		panic(fmt.Sprintln("Ошибка измнения потока логов: ", errResp.Msg))
	}
	C.td_send(
		clientId,
		C.CString(
			fmt.Sprintf(`{
				"@type": "setTdlibParameters",
				"use_test_dc": false,
				"database_directory": "./tdlib-db",
				"files_directory": "./tdlib-files",
				"database_encryption_key": "",
				"use_file_database": true,
				"use_chat_info_database": true,
				"use_message_database": true,
				"use_secret_chats": false,
				"api_id": %s,
				"api_hash": "%s",
				"system_language_code": "en",
				"device_model": "Desktop",
				"system_version": "Ubuntu 24.04",
				"application_version": "1.0"
			}`, os.Getenv("APP_ID"), os.Getenv("APP_HASH")),
		),
	)
	for {
		var result = C.td_receive(timeout)
		if result != nil {
			res := []byte(C.GoString(result))
			var resData CommonResp
			if err := json.Unmarshal(res, &resData); err != nil {
				fmt.Println("Ошибка преобразования commonResp: ", err)
				break
			}

			if resData.Type == "updateActiveEmojiReactions" {
				var emojiResp EmojiResp
				if err := json.Unmarshal(res, &emojiResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				activeReactions = emojiResp.Emojis
			}

			if resData.Type == "updateChatAvailableReactions" {
				var reactResp ChatReactResp
				if err := json.Unmarshal(res, &reactResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				switch reactResp.Reactions.Type {
				case "chatAvailableReactionsAll":
					availableReactions[reactResp.ChatId] = activeReactions
				case "chatAvailableReactionsSome":
					var someResp ReactionSomeResp
					if err := json.Unmarshal(res, &someResp); err != nil {
						fmt.Println("Ошибка преобразования json: ", err)
						break
					}
					availableReactions = map[int64][]string{}
					for i, r := range someResp.Reactions.Reactions {
						if r.Type == "reactionTypeEmoji" {
							var emojResp ReactionEmojiResp
							if err := json.Unmarshal(res, &emojResp); err != nil {
								fmt.Println("Ошибка преобразования json: ", err)
								break
							}
							availableReactions[reactResp.ChatId] = append(availableReactions[reactResp.ChatId], emojResp.Reactions.Reactions[i].Emoji)
						}
					}
				default:
					fmt.Println("Неправильный тип доступной реакции: ", reactResp.Reactions.Type)
					return
				}
			}

			if resData.Type == "updateChatLastMessage" {
				var msgResp NewLastMsgResp
				if err := json.Unmarshal(res, &msgResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				if msgResp.ChatId == selectedChat {
					var emoji string
					if len(availableReactions[selectedChat]) < 1 {
						emoji = "👍"
					} else {
						emoji = availableReactions[selectedChat][rand.Intn(len(availableReactions[selectedChat]))]
					}
					C.td_send(
						clientId,
						C.CString(
							fmt.Sprintf(
								`{
									"@type": "addMessageReaction",	
									"chat_id": %d,
									"message_id": %d,
									"reaction_type": {
										"@type": "reactionTypeEmoji",
										"emoji": "%s"
									},
									"is_big": false,
									"update_recent_reactions": false
								}`,
								selectedChat,
								msgResp.LastMsg.Id,
								emoji,
							),
						),
					)
				}
				log.Println("Реакция отправлена!")
				time.Sleep(1 * time.Second)
			}

			if resData.Type == "messages" {
				var msgsResp MsgsResp
				if err := json.Unmarshal(res, &msgsResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				fmt.Println("Реакции начались отправляться!")
				for _, msg := range msgsResp.Msgs {
					var emoji string
					if len(availableReactions[selectedChat]) < 1 {
						emoji = "👍"
					} else {
						emoji = availableReactions[selectedChat][rand.Intn(len(availableReactions[selectedChat]))]
					}
					C.td_send(
						clientId,
						C.CString(
							fmt.Sprintf(
								`{
									"@type": "addMessageReaction",	
									"chat_id": %d,
									"message_id": %d,
									"reaction_type": {
										"@type": "reactionTypeEmoji",
										"emoji": "%s"
									},
									"is_big": false,
									"update_recent_reactions": false
								}`,
								selectedChat,
								msg.Id,
								emoji,
							),
						),
					)
					log.Println("Реакция отправлена!")
					time.Sleep(1 * time.Second)
				}
			}

			if resData.Type == "updateAuthorizationState" {
				var authResp AuthResp
				if err := json.Unmarshal(res, &authResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				if authResp.State.Type == "authorizationStateReady" {
					fmt.Println("Вы успшено зарегестрированы!")
				}
				authState = authResp.State.Type
				if err := nextStep(
					int(clientId),
					authState,
					false,
					func(data int64) {
						if authState == "authorizationStateReady" {
							selectedChat = data
						}
					}); err != nil {
					fmt.Println("Ошибка перехода на следующий шаг: ", err)
					break
				}
			}
			if resData.Type == "error" {
				var errResp ErrorResp
				if err := json.Unmarshal(res, &errResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				fmt.Println(errResp.Msg)
				if err := nextStep(
					int(clientId),
					authState,
					true,
					func(data int64) {
						if authState == "authorizationStateReady" {
							selectedChat = data
						}
					}); err != nil {
					fmt.Println("Ошибка перехода на следующий шаг: ", err)
					break
				}
			}
		}
	}
}
