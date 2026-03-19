package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"td/internal/model"
	"td/internal/tdlib"

	"google.golang.org/genai"
)

func registerHostel() *model.Hostel {
	return &model.Hostel{
		Mutex:       sync.Mutex{},
		Reactions:   &[]string{},
		ReactedMsgs: make(map[int64]bool),
	}
}

func InitApp() model.State {
	hostel := registerHostel()
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		panic(fmt.Sprint("Ошибка создания AI client: ", err))
	}

	res := tdlib.Execute(`{
		"@type": "setLogVerbosityLevel",
		"new_verbosity_level": 3
	}`)
	var resData model.CommonResp
	if err := json.Unmarshal(res, &resData); err != nil {
		panic(fmt.Sprintln("Ошибка преобразования json: ", err))
	}
	if resData.Type == "error" {
		var errResp model.ErrorResp
		if err := json.Unmarshal(res, &errResp); err != nil {
			panic(fmt.Sprintln("Ошибка преобразования json: ", err))
		}
		panic(fmt.Sprintln("Ошибка изменения потока логов: ", errResp.Msg))
	}

	if err := os.Remove("tdlib.log"); err != nil {
		if !os.IsNotExist(err) {
			panic(fmt.Sprintln("Ошибка удаления файла: ", err))
		}
	}

	res = tdlib.Execute(`{
		"@type": "setLogStream",
		"log_stream": {
			"@type": "logStreamFile",
			"path": "./tdlib.log",
			"max_file_size": 10485760,
			"redirect_stderr": true
		}
	}`)
	if err := json.Unmarshal(res, &resData); err != nil {
		panic(fmt.Sprintln("Ошибка преобразования json: ", err))
	}
	if resData.Type == "error" {
		var errResp model.ErrorResp
		if err := json.Unmarshal(res, &errResp); err != nil {
			panic(fmt.Sprintln("Ошибка преобразования json: ", err))
		}
		panic(fmt.Sprintln("Ошибка изменения потока логов: ", errResp.Msg))
	}

	clientId := tdlib.CreateClientId()
	tdlib.Send(clientId, fmt.Sprintf(`{
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
	}`, os.Getenv("APP_ID"), os.Getenv("APP_HASH")))

	return model.State{
		Hostel:       hostel,
		Ctx:          ctx,
		Client:       client,
		MsgsChan:     make(chan []model.Message),
		MsgChan:      make(chan model.LastMsgResp),
		ReactsChan:   make(chan model.ReactionsResp),
		AuthState:    "authorizationStateWaitTdlibParameters",
		Limit:        0,
		SelectedChat: 0,
		ClientId:     clientId,
	}
}
