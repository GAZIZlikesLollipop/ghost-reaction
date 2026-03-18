package main

//#include "include/td_json_client.h"
//#cgo LDFLAGS: -L${SRCDIR}/lib -ltdjson -Wl,-rpath,${SRCDIR}
import "C"

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
type CommonReactType struct {
	Type string `json:"@type"`
}

type CommonReaction struct {
	Type   CommonReactType `json:"type"`
	ForVIP bool            `json:"needs_premium"`
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
	callback func(chatId int64),
	messages chan []Message,
	reactChan chan ReactionsResp,
	done chan bool,
	hostel *Hostel,
) {
	var chatIdStr string
	fmt.Print("Введите id чата у которого хотите поставить реакции: ")
	fmt.Scan(&chatIdStr)
	re := regexp.MustCompile(`[0-9]`)
	chatId, _ := strconv.ParseInt(strings.Join(re.FindAllString(chatIdStr, -1), ""), 10, 64)
	callback(chatId)
	reInt := regexp.MustCompile(`[0-9]`)
	var limitStr string
	fmt.Print("На сколько последних сообщений хотите поставить реакции? (не более 100): ")
	fmt.Scan(&limitStr)
	limit, _ := strconv.ParseInt(strings.Join(reInt.FindAllString(limitStr, -1), ""), 10, 64)
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
	msgs := <-messages
	// reactResp := <-reactChan
	// var rawReactions []CommonReaction
	// var (
	// 	emojiReaction  EmojiReactions
	// 	emojiReactions []EmojiReaction
	// )
	// rawReactions = append(reactResp.Resp.Top, reactResp.Resp.Popular...)
	// rawReactions = append(rawReactions, reactResp.Resp.Recent...)
	// for i, r := range rawReactions {
	// 	if !r.ForVIP && r.Type.Type == "reactionTypeEmoji" {
	// 		if len(emojiReactions) < 1 {
	// 			json.Unmarshal(reactResp.Req, &emojiReaction)
	// 			emojiReactions = append(emojiReaction.Top, emojiReaction.Popular...)
	// 			emojiReactions = append(emojiReactions, emojiReaction.Recent...)
	// 		}
	// 		reactions = append(reactions, emojiReactions[i].Type.Emoji)
	// 	}
	// }
	// updateReact(reactions)
	for _, msg := range msgs {
		var emoji string
		if len(hostel.reactions) < 1 {
			emoji = "👍"
		} else {
			emoji = hostel.reactions[rand.Intn(len(hostel.reactions))]
		}
		hostel.checkReacted(msg.ChatId)
		hostel.mutex.Lock()
		if !hostel.reactedMsgs[msg.ChatId] {
			sendReaction(clientId, chatId, msg.Id, emoji, "reactions")
			hostel.reactedMsgs[msg.ChatId] = true
		}
		hostel.mutex.Unlock()
	}
	done <- true
}

func reactNewMessage(
	clientId int,
	msgChan chan LastMsgResp,
	hostel *Hostel,
) {
	reactions := hostel.reactions
	var emoji string
	if len(reactions) < 1 {
		emoji = "👍"
	} else {
		emoji = reactions[rand.Intn(len(reactions))]
	}
	for {
		message := <-msgChan
		hostel.mutex.Lock()
		if !hostel.reactedMsgs[message.ChatId] {
			sendReaction(clientId, message.ChatId, message.LastMsg.Id, emoji, "reaction")
			hostel.reactedMsgs[message.ChatId] = true
		}
		hostel.mutex.Unlock()
	}
}

type Hostel struct {
	mutex        sync.Mutex
	reactions    []string
	reactedMsgs  map[int64]bool
	doneChan     chan bool
	messages     chan []Message
	newMessage   chan LastMsgResp
	reactionChan chan ReactionsResp
}

func registerHostel() *Hostel {
	return &Hostel{
		mutex:       sync.Mutex{},
		reactions:   []string{},
		reactedMsgs: make(map[int64]bool),
	}
}

func (h *Hostel) checkReacted(chatId int64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if _, ok := h.reactedMsgs[chatId]; !ok {
		h.reactedMsgs[chatId] = false
	}
}

func (h *Hostel) editReactions(reactions []string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.reactions = reactions
}

func main() {
	hostel := registerHostel()
	doneChan := make(chan bool)
	messages := make(chan []Message)
	newMessage := make(chan LastMsgResp)
	reactionChan := make(chan ReactionsResp)
	authState := "authorizationStateWaitTdlibParameters"
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

			// if resData.Type == "availableReactions" {
			// 	var reactResp CommonReactions
			// 	if err := json.Unmarshal(res, &reactResp); err != nil {
			// 		fmt.Println("Ошибка преобразования reactions: ", err)
			// 		break
			// 	}
			// 	reactionChan <- ReactionsResp{
			// 		Resp: reactResp,
			// 		Req:  res,
			// 	}
			// }

			if resData.Type == "messages" {
				var msgsResp MsgsResp
				if err := json.Unmarshal(res, &msgsResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				// C.td_send(
				// 	clientId,
				// 	C.CString(
				// 		fmt.Sprintf(
				// 			`{
				// 				"@type": "getMessageAvailableReactions",
				// 				"chat_id": %d,
				// 				"message_id": %d,
				// 				"row_size": 25
				// 			}`,
				// 			msgsResp.Msgs[0].ChatId,
				// 			msgsResp.Msgs[0].Id,
				// 		),
				// 	),
				// )
				if len(msgsResp.Msgs) > 1 {
					messages <- msgsResp.Msgs
				}
			}

			if resData.Type == "updateChatLastMessage" {
				go func() {
					var msgResp LastMsgResp
					if err := json.Unmarshal(res, &msgResp); err != nil {
						fmt.Println("Ошибка преобразования json: ", err)
					}
					hostel.checkReacted(msgResp.ChatId)
					if selectedChat == msgResp.LastMsg.ChatId {
						hostel.newMessage <- LastMsgResp{
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
				if authResp.State.Type == "authorizationStateReady" {
					fmt.Println("Вы успешно зарегестрированы!")
				}
				authState = authResp.State.Type
				if authState == "authorizationStateReady" {
					go reactMessages(
						int(clientId),
						func(chatId int64) {
							selectedChat = chatId
						},
						messages,
						reactionChan,
						doneChan,
						hostel,
					)
					go reactNewMessage(int(clientId), newMessage, hostel)
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
				go func() {
					var errResp ErrorResp
					if err := json.Unmarshal(res, &errResp); err != nil {
						fmt.Println("Ошибка парсинга ошибки: ", err)
					}
					messages <- []Message{}
					isDone := <-doneChan
					if isDone {
						go reactMessages(
							int(clientId),
							func(chatId int64) {
								selectedChat = chatId
							},
							messages,
							reactionChan,
							doneChan,
							hostel,
						)
						fmt.Println(errResp.Msg)
					}
				}()
			}
		}
	}
}
