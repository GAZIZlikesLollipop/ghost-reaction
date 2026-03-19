package main

//#include "include/td_json_client.h"
//#cgo LDFLAGS: -L${SRCDIR}/lib -ltdjson -Wl,-rpath,${SRCDIR}/lib
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

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

func nextStep(
	clientId int,
	state string,
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
	return nil
}

func sendReaction(
	clientId int,
	chatId int64,
	msgId int64,
	emoji string,
	extra string,
) {
	C.td_send(
		C.int(clientId),
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
					"update_recent_reactions": false,
					"@extra": "%s"
				}`,
				chatId,
				msgId,
				emoji,
				extra,
			),
		),
	)
	time.Sleep(1 * time.Second)
}

// raw
type CommonReaction struct {
	Type   CommonResp `json:"type"`
	ForVIP bool       `json:"needs_premium"`
}

type CommonReactions struct {
	Top     []CommonReaction `json:"top_reactions"`
	Recent  []CommonReaction `json:"recent_reactions"`
	Popular []CommonReaction `json:"popular_reactions"`
}

type ReactionsResp struct {
	Resp CommonReactions
	Req  []byte
}

// formated
type EmojiReactType struct {
	Emoji string `json:"emoji"`
}

type EmojiReaction struct {
	Type EmojiReactType `json:"type"`
}

type EmojiReactions struct {
	Top     []EmojiReaction `json:"top_reactions"`
	Recent  []EmojiReaction `json:"recent_reactions"`
	Popular []EmojiReaction `json:"popular_reactions"`
}

func reactMessages(
	clientId int,
	callback func(chatId int64, limit uint64),
	messages chan []Message,
	reactChan chan ReactionsResp,
	hostel *Hostel,
	client *genai.Client,
	ctx context.Context,
) {
	hostel.mutex.Lock()
	defer hostel.mutex.Unlock()
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
					"only_local": false
				}`,
				chatId,
				limit,
			),
		),
	)
	time.Sleep(1 * time.Second)
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
					"only_local": false
				}`,
				chatId,
				limit,
			),
		),
	)
	reactResp := <-reactChan
	var rawReactions []CommonReaction
	var (
		emojiReaction  EmojiReactions
		emojiReactions []EmojiReaction
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
			*hostel.reactions = append(*hostel.reactions, emojiReactions[i].Type.Emoji)
		}
	}
	msgs := <-messages
	reactions := *hostel.reactions
	text := "bro i so tired, tomorrow ok?"
	for _, msg := range msgs {
		if _, ok := hostel.reactedMsgs[msg.ChatId]; !ok {
			hostel.reactedMsgs[msg.ChatId] = false
		}
		if !hostel.reactedMsgs[msg.Id] {
			parts := []*genai.Part{
				{
					Text: fmt.Sprintf(`
						You will be given a list of emojis and a text. Your task is to select the most fitting emoji from the provided list based on the meaning or mood of the text.

						Rules:
						- You MUST only return an emoji that exists in the provided list
						- Return ONLY the emoji itself, nothing else — no words, no punctuation, no explanation
						- The response must be a single emoji character

						List of emojis:
						%s

						Text:
						%s
					`,
						strings.Join(reactions, ", "),
						text,
					),
				},
			}
			result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{{Parts: parts}}, nil)
			if err != nil {
				fmt.Println("Ошибка получения результата от AI: ", err)
			}
			sendReaction(clientId, chatId, msg.Id, result.Text(), "reactions")
			hostel.reactedMsgs[msg.Id] = true
		}
	}
}

func reactNewMessage(
	clientId int,
	msgChan chan LastMsgResp,
	hostel *Hostel,
	client *genai.Client,
	ctx context.Context,
) {
	var (
		reactions []string
		text      string
	)
	text = "ohh, yeah i think"
	for {
		message := <-msgChan
		hostel.mutex.Lock()
		if _, ok := hostel.reactedMsgs[message.LastMsg.Id]; !ok {
			hostel.reactedMsgs[message.LastMsg.Id] = false
		}
		if !hostel.reactedMsgs[message.LastMsg.Id] {
			reactions = *hostel.reactions
			parts := []*genai.Part{
				{
					Text: fmt.Sprintf(`
						You will be given a list of emojis and a text. Your task is to select the most fitting emoji from the provided list based on the meaning or mood of the text.

						Rules:
						- You MUST only return an emoji that exists in the provided list
						- Return ONLY the emoji itself, nothing else — no words, no punctuation, no explanation
						- The response must be a single emoji character

						List of emojis:
						%s

						Text:
						%s
					`,
						strings.Join(reactions, ","),
						text,
					),
				},
			}
			result, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{{Parts: parts}}, nil)
			if err != nil {
				fmt.Println("Ошибка получения результата от AI: ", err)
			}
			sendReaction(clientId, message.ChatId, message.LastMsg.Id, result.Text(), "reaction")
			hostel.reactedMsgs[message.LastMsg.Id] = true
		}
		hostel.mutex.Unlock()
	}
}

type Hostel struct {
	mutex        sync.Mutex
	reactions    *[]string
	reactedMsgs  map[int64]bool
	doneChan     chan bool
	messages     chan []Message
	newMessage   chan LastMsgResp
	reactionChan chan ReactionsResp
}

func registerHostel() *Hostel {
	return &Hostel{
		mutex:       sync.Mutex{},
		reactions:   &[]string{},
		reactedMsgs: make(map[int64]bool),
	}
}

func main() {
	hostel := registerHostel()
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		panic(fmt.Sprint("Ошибка создания AI client: ", err))
	}
	messages := make(chan []Message)
	newMessage := make(chan LastMsgResp)
	reactionChan := make(chan ReactionsResp)
	authState := "authorizationStateWaitTdlibParameters"
	var limit uint64
	var selectedChat int64 = 0
	clientId := C.td_create_client_id()
	timeout := C.double(1.0)
	result := C.td_execute(
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
		if !os.IsNotExist(err) {
			panic(fmt.Sprintln("Ошибка удаления файла: ", err))
		}
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

			if resData.Type == "ok" {
				var extra ExtraResp
				err := json.Unmarshal(res, &extra)
				if err == nil {
					switch extra.Extra {
					case "reactions":
						fmt.Println("Реакция успешно отправлена!")
					case "reaction":
						fmt.Println("Отправлена реакция на новое сообщение!")
					}
				} else {
				}
			}

			if resData.Type == "availableReactions" {
				go func() {
					var reactResp CommonReactions
					if err := json.Unmarshal(res, &reactResp); err != nil {
						fmt.Println("Ошибка преобразования reactions: ", err)
						return
					}
					reactionChan <- ReactionsResp{
						Resp: reactResp,
						Req:  res,
					}
				}()
			}

			if resData.Type == "messages" {
				var msgsResp MsgsResp
				if err := json.Unmarshal(res, &msgsResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				C.td_send(
					clientId,
					C.CString(
						fmt.Sprintf(
							`{
								"@type": "getMessageAvailableReactions",
								"chat_id": %d,
								"message_id": %d,
								"row_size": 25
							}`,
							msgsResp.Msgs[0].ChatId,
							msgsResp.Msgs[0].Id,
						),
					),
				)
				if len(msgsResp.Msgs) >= int(limit) {
					messages <- msgsResp.Msgs
				}
			}

			if resData.Type == "updateChatLastMessage" {
				go func() {
					var msgResp LastMsgResp
					if err := json.Unmarshal(res, &msgResp); err != nil {
						fmt.Println("Ошибка преобразования json: ", err)
					}
					if selectedChat == msgResp.LastMsg.ChatId {
						newMessage <- LastMsgResp{
							ChatId:  selectedChat,
							LastMsg: msgResp.LastMsg,
						}
					}
				}()
			}

			if resData.Type == "updateAuthorizationState" {
				var authResp AuthResp
				if err := json.Unmarshal(res, &authResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				authState = authResp.State.Type
				if authState == "authorizationStateReady" {
					fmt.Println("Вы успешно зарегестрированы!")
					if err := nextStep(
						int(clientId),
						authState,
					); err != nil {
						fmt.Println("Ошибка перехода на следующий шаг: ", err)
						break
					}
					time.Sleep(1 * time.Second)
					go reactMessages(
						int(clientId),
						func(chatId int64, lmt uint64) {
							selectedChat = chatId
							limit = lmt
						},
						messages,
						reactionChan,
						hostel,
						client,
						ctx,
					)
					go reactNewMessage(int(clientId), newMessage, hostel, client, ctx)
				} else {
					if err := nextStep(
						int(clientId),
						authState,
					); err != nil {
						fmt.Println("Ошибка перехода на следующий шаг: ", err)
						break
					}
				}
			}

			if resData.Type == "error" {
				var errResp ErrorResp
				if err := json.Unmarshal(res, &errResp); err != nil {
					fmt.Println("Ошибка парсинга ошибки: ", err)
				}
				fmt.Println(errResp.Msg)
				break
			}
		}
	}
}
