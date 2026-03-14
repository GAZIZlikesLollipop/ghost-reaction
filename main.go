package main

//#include "include/td_json_client.h"
//#cgo LDFLAGS: -L${SRCDIR}/lib -ltdjson -Wl,-rpath,${SRCDIR}
import "C"

import (
	"encoding/json"
	"fmt"
	"os"
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

func main() {
	var authState string = "authorizationStateWaitTdlibParameters"
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
	if err := os.RemoveAll("tdlib.log"); err != nil {
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
				fmt.Println("Ошибка преобразования json: ", err)
				break
			}
			if resData.Type == "updateAuthorizationState" {
				var authResp AuthResp
				if err := json.Unmarshal(res, &authResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				authState = authResp.State.Type
				nextStep(int(clientId), authState)
			}
			if resData.Type == "error" {
				var errResp ErrorResp
				if err := json.Unmarshal(res, &errResp); err != nil {
					fmt.Println("Ошибка преобразования json: ", err)
					break
				}
				fmt.Println(errResp.Msg)
				if authState != "authorizationStateReady" {
					nextStep(int(clientId), authState)
				}
			}
		}
	}
}
