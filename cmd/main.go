package main

import (
	"encoding/json"
	"fmt"
	"td/internal/model"
	"td/internal/service"
	"td/internal/tdlib"
	"time"
)

func main() {
	state := service.InitApp()

	for {
		res := tdlib.Receive(1.0)
		if res == nil {
			continue
		}

		var resData model.CommonResp
		if err := json.Unmarshal(res, &resData); err != nil {
			fmt.Println("Ошибка преобразования commonResp: ", err)
			break
		}

		if resData.Type == "ok" {
			var extra model.ExtraResp
			if err := json.Unmarshal(res, &extra); err == nil {
				switch extra.Extra {
				case "reactions":
					fmt.Println("Реакция успешно отправлена!")
				case "reaction":
					fmt.Println("Отправлена реакция на новое сообщение!")
				}
			}
		}

		if resData.Type == "file" {
			go func() {
				var fileResp model.TdFile
				if err := json.Unmarshal(res, &fileResp); err != nil {
					fmt.Println("Ошибка преобразования file: ", err)
					return
				}
				state.FileChan <- fileResp
			}()
		}

		if resData.Type == "availableReactions" {
			go func() {
				var reactResp model.CommonReactions
				if err := json.Unmarshal(res, &reactResp); err != nil {
					fmt.Println("Ошибка преобразования reactions: ", err)
					return
				}
				state.ReactsChan <- model.ReactionsResp{
					Resp: reactResp,
					Req:  res,
				}
			}()
		}

		if resData.Type == "messages" {
			go func() {
				var msgsResp model.MsgsResp
				if err := json.Unmarshal(res, &msgsResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					return
				}
				tdlib.Send(state.ClientId, fmt.Sprintf(`{
					"@type": "getMessageAvailableReactions",
					"chat_id": %d,
					"message_id": %d,
					"row_size": 25
				}`, msgsResp.Msgs[0].ChatId, msgsResp.Msgs[0].Id))
				if len(msgsResp.Msgs) >= int(state.Limit) {
					state.MsgsChan <- msgsResp.Msgs
				}
			}()
		}

		if resData.Type == "updateChatLastMessage" {
			go func() {
				var msgResp model.LastMsgResp
				if err := json.Unmarshal(res, &msgResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					return
				}
				if state.SelectedChat == msgResp.LastMsg.ChatId {
					state.MsgChan <- model.LastMsgResp{
						ChatId:  state.SelectedChat,
						LastMsg: msgResp.LastMsg,
					}
				}
			}()
		}

		if resData.Type == "updateAuthorizationState" {
			var authResp model.AuthResp
			if err := json.Unmarshal(res, &authResp); err != nil {
				fmt.Println("Ошибка преобразования json: ", err)
				break
			}
			state.AuthState = authResp.State.Type
			if state.AuthState == "authorizationStateReady" {
				fmt.Println("Вы успешно зарегистрированы!")
				if err := service.NextStep(state.ClientId, state.AuthState); err != nil {
					fmt.Println("Ошибка перехода на следующий шаг: ", err)
					break
				}
				time.Sleep(1 * time.Second)
				go service.ReactMessages(
					state.ClientId,
					func(chatId int64, lmt uint64) {
						state.SelectedChat = chatId
						state.Limit = lmt
					},
					state.MsgsChan,
					state.ReactsChan,
					state.Hostel,
					state.Client,
					state.Ctx,
					state.FileChan,
				)
				go service.ReactNewMessage(state.ClientId, state.MsgChan, state.Hostel, state.Client, state.Ctx, state.FileChan)
			} else {
				if err := service.NextStep(state.ClientId, state.AuthState); err != nil {
					fmt.Println("Ошибка перехода на следующий шаг: ", err)
					break
				}
			}
		}

		if resData.Type == "error" {
			var errResp model.ErrorResp
			if err := json.Unmarshal(res, &errResp); err != nil {
				fmt.Println("Ошибка парсинга ошибки: ", err)
				break
			}
			fmt.Println(errResp.Msg)
			break
		}
	}
}
