package service

import (
	"fmt"
	"td/internal/tdlib"
)

func NextStep(clientId int, state string) error {
	switch state {
	case "authorizationStateWaitPhoneNumber":
		var phoneNumber string
		fmt.Print("Введите ваш номер телефона: ")
		fmt.Scan(&phoneNumber)
		tdlib.Send(clientId, fmt.Sprintf(`{
			"@type": "setAuthenticationPhoneNumber",
			"phone_number": "%s"
		}`, phoneNumber))

	case "authorizationStateWaitCode":
		var code string
		fmt.Print("Введите код подтверждения: ")
		fmt.Scan(&code)
		if len(code) < 5 {
			tdlib.Send(clientId, `{
				"@type": "resendAuthenticationCode",
				"reason": {
					"@type": "resendCodeReasonUserRequest"
				}
			}`)
		} else {
			tdlib.Send(clientId, fmt.Sprintf(`{
				"@type": "checkAuthenticationCode",
				"code": "%s"
			}`, code))
		}

	case "authorizationStateWaitPassword":
		var password string
		fmt.Print("Введите пароль от аккаунта: ")
		fmt.Scan(&password)
		tdlib.Send(clientId, fmt.Sprintf(`{
			"@type": "checkAuthenticationPassword",
			"password": "%s"
		}`, password))

	case "authorizationStateReady":
		tdlib.Send(clientId, `{
			"@type": "loadChats",
			"chat_list": {
				"@type": "chatListMain"
			},
			"limit": 100
		}`)
		tdlib.Send(clientId, `{
			"@type": "loadChats",
			"chat_list": {
				"@type": "chatListArchive"
			},
			"limit": 100
		}`)
	}
	return nil
}
