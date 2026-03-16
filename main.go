package main

//#include "include/td_json_client.h"
//#cgo LDFLAGS: -L${SRCDIR}/lib -ltdjson -Wl,-rpath,${SRCDIR}
import "C"

import (
	"encoding/json"
	"fmt"
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

func reactMessages(
	clientId int,
	callback func(chatId int64),
	messages chan []Message,
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
	for _, msg := range msgs {
		var emoji string = "👍"
		sendReaction(clientId, chatId, msg.Id, emoji, "reactions")
	}
}

func reactNewMessage(
	clientId int,
	msgChan chan LastMsgResp,
) {
	var emoji string = "👍"
	for {
		message := <-msgChan
		sendReaction(clientId, message.ChatId, message.LastMsg.Id, emoji, "reaction")
	}
}

func main() {
	var (
		authState    string = "authorizationStateWaitTdlibParameters"
		selectedChat int64
	)
	isRunning := false
	messages := make(chan []Message)
	newMessage := make(chan LastMsgResp)
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

			if resData.Type == "ok" {
				var extra ExtraResp
				err := json.Unmarshal(res, &extra)
				if err == nil {
					switch extra.Extra {
					case "reactions":
						fmt.Println("Реакция успешно отправлена!")
					case "reaction":
						fmt.Println("Реакция успешно отправлена! (введите new для выбора чата)")
						if !isRunning {
							go func(callback func(data bool)) {
								callback(true)
								var choice string
								fmt.Scan(&choice)
								choice = strings.TrimSpace(choice)
								switch choice {
								case "new":
									go reactMessages(
										int(clientId),
										func(chatId int64) {
											selectedChat = chatId
										},
										messages,
									)
								case "stop":
									fmt.Println("Программа завершила работу")
									os.Exit(0)
								}
								callback(false)
							}(func(data bool) {
								isRunning = data
							})
						}
					}
				} else {
				}
			}

			if resData.Type == "messages" {
				var msgsResp MsgsResp
				if err := json.Unmarshal(res, &msgsResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
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
					if selectedChat == msgResp.ChatId {
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
					)
					go reactNewMessage(int(clientId), newMessage)
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
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				fmt.Println(errResp.Msg)
			}
		}
	}
}
