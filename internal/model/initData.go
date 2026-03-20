package model

import (
	"context"
	"sync"

	"google.golang.org/genai"
)

type Hostel struct {
	Mutex       sync.Mutex
	Reactions   *[]string
	ReactedMsgs map[int64]bool
}

type State struct {
	Hostel       *Hostel
	Ctx          context.Context
	Client       *genai.Client
	MsgsChan     chan []Message
	MsgChan      chan LastMsgResp
	ReactsChan   chan ReactionsResp
	FileChan     chan TdFile
	AuthState    string
	Limit        uint64
	SelectedChat int64
	ClientId     int
}
