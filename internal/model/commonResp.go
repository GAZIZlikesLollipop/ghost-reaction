package model

type CommonResp struct {
	Type string `json:"@type"`
}

type AuthResp struct {
	State CommonResp `json:"authorization_state"`
}

type ErrorResp struct {
	Msg string `json:"message"`
}

type ExtraResp struct {
	Extra string `json:"@extra"`
}
